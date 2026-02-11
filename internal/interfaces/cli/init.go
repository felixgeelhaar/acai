package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	domain "github.com/felixgeelhaar/acai/internal/domain/auth"
	infraauth "github.com/felixgeelhaar/acai/internal/infrastructure/auth"
	"github.com/felixgeelhaar/acai/internal/infrastructure/config"
)

// newInitCmd creates the "acai init" command.
// Self-bootstraps its own minimal dependencies (like the version command)
// since it runs before full app wiring.
func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Set up acai configuration interactively",
		Long:  "Guided setup wizard that creates ~/.acai/config.yaml and optionally stores your API token.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd)
		},
	}
}

func runInit(cmd *cobra.Command) error {
	out := cmd.OutOrStdout()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}

	acaiDir := filepath.Join(homeDir, ".acai")
	configPath := filepath.Join(acaiDir, "config.yaml")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		var overwrite bool
		err := huh.NewConfirm().
			Title("Config file already exists at " + configPath).
			Description("Do you want to overwrite it?").
			Value(&overwrite).
			Run()
		if err != nil {
			return err
		}
		if !overwrite {
			_, _ = fmt.Fprintln(out, "Init cancelled.")
			return nil
		}
	}

	// Auto-detect local Granola cache
	defaultCachePath := filepath.Join(homeDir, "Library", "Application Support", "Granola", "cache-v3.json")
	cacheDetected := false
	if _, err := os.Stat(defaultCachePath); err == nil {
		cacheDetected = true
	}

	// Build connection options
	var options []huh.Option[string]
	if cacheDetected {
		options = []huh.Option[string]{
			huh.NewOption("Local cache (Granola desktop app detected)", "local_cache"),
			huh.NewOption("API token (Enterprise)", "api"),
		}
	} else {
		options = []huh.Option[string]{
			huh.NewOption("API token (Enterprise)", "api"),
			huh.NewOption("Local cache (specify path)", "local_cache"),
		}
	}

	var dataSource string
	err = huh.NewSelect[string]().
		Title("How would you like to connect to Granola?").
		Options(options...).
		Value(&dataSource).
		Run()
	if err != nil {
		return err
	}

	fileCfg := config.FileConfig{
		DataSource: dataSource,
	}

	// Handle API token flow
	if dataSource == "api" {
		var token string
		err = huh.NewInput().
			Title("Enter your Granola API token").
			EchoMode(huh.EchoModePassword).
			Value(&token).
			Run()
		if err != nil {
			return err
		}

		// Validate and store token via auth service
		tokenStore := infraauth.NewFileTokenStore(acaiDir)
		authService := infraauth.NewService(tokenStore)

		_, err = authService.Login(context.Background(), domain.LoginParams{
			Method:   domain.AuthAPIToken,
			APIToken: token,
		})
		if err != nil {
			return fmt.Errorf("token validation failed: %w", err)
		}

		_, _ = fmt.Fprintln(out, "API token stored in credentials.json")
	}

	// Handle local cache path
	if dataSource == "local_cache" && !cacheDetected {
		var cachePath string
		err = huh.NewInput().
			Title("Enter the path to your Granola cache file").
			Placeholder(defaultCachePath).
			Value(&cachePath).
			Run()
		if err != nil {
			return err
		}

		if cachePath != "" {
			if _, err := os.Stat(cachePath); err != nil {
				return fmt.Errorf("cache file not found: %s", cachePath)
			}
			fileCfg.Granola.CachePath = cachePath
		}
	}

	// Write config file
	if err := config.WriteConfigFile(configPath, fileCfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Print summary
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Configuration saved to", configPath)
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "  data_source:", fileCfg.DataSource)
	if fileCfg.Granola.APIURL != "" {
		_, _ = fmt.Fprintln(out, "  api_url:    ", fileCfg.Granola.APIURL)
	}
	if fileCfg.Granola.CachePath != "" {
		_, _ = fmt.Fprintln(out, "  cache_path: ", fileCfg.Granola.CachePath)
	}
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Run 'acai meeting list' to verify your setup.")

	return nil
}
