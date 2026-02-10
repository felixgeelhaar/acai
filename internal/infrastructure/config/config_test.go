package config_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/felixgeelhaar/acai/internal/infrastructure/config"
)

func TestDefault_HasSaneDefaults(t *testing.T) {
	cfg := config.Default()

	if cfg.Granola.APIURL != "https://public-api.granola.ai" {
		t.Errorf("got api url %q", cfg.Granola.APIURL)
	}
	if cfg.MCP.Transport != "stdio" {
		t.Errorf("got transport %q", cfg.MCP.Transport)
	}
	if cfg.MCP.ServerName != "acai" {
		t.Errorf("got server name %q", cfg.MCP.ServerName)
	}
	if cfg.Cache.TTL != 15*time.Minute {
		t.Errorf("got cache ttl %v", cfg.Cache.TTL)
	}
	if cfg.Resilience.Retry.MaxAttempts != 3 {
		t.Errorf("got max attempts %d", cfg.Resilience.Retry.MaxAttempts)
	}
	if cfg.Resilience.Timeout != 30*time.Second {
		t.Errorf("got timeout %v", cfg.Resilience.Timeout)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("got log level %q", cfg.Logging.Level)
	}
}

func TestLoad_OverridesFromEnv(t *testing.T) {
	t.Setenv("ACAI_GRANOLA_API_URL", "https://custom.api.com")
	t.Setenv("ACAI_GRANOLA_API_TOKEN", "test-token")
	t.Setenv("ACAI_LOGGING_LEVEL", "debug")
	t.Setenv("ACAI_MCP_HTTP_PORT", "9090")

	cfg := config.Load()

	if cfg.Granola.APIURL != "https://custom.api.com" {
		t.Errorf("got api url %q", cfg.Granola.APIURL)
	}
	if cfg.Granola.APIToken != "test-token" {
		t.Errorf("got api token %q", cfg.Granola.APIToken)
	}
	if cfg.Granola.AuthMethod != "api_token" {
		t.Errorf("got auth method %q, want api_token when token is set", cfg.Granola.AuthMethod)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("got log level %q", cfg.Logging.Level)
	}
	if cfg.MCP.HTTPPort != 9090 {
		t.Errorf("got http port %d", cfg.MCP.HTTPPort)
	}
}

func TestLoad_PolicyFileEnv(t *testing.T) {
	t.Setenv("ACAI_POLICY_FILE", "/etc/granola/policy.yaml")

	cfg := config.Load()

	if !cfg.Policy.Enabled {
		t.Error("expected policy enabled when file path is set")
	}
	if cfg.Policy.FilePath != "/etc/granola/policy.yaml" {
		t.Errorf("got policy file %q", cfg.Policy.FilePath)
	}
}

func TestDefault_PolicyDisabled(t *testing.T) {
	cfg := config.Default()

	if cfg.Policy.Enabled {
		t.Error("expected policy disabled by default")
	}
	if cfg.Policy.FilePath != "" {
		t.Errorf("got policy file %q, want empty", cfg.Policy.FilePath)
	}
}

func TestDefault_DataSourceAuto(t *testing.T) {
	cfg := config.Default()

	if cfg.Granola.DataSource != "auto" {
		t.Errorf("got data source %q, want auto", cfg.Granola.DataSource)
	}
	if cfg.Granola.LocalCachePath != "" {
		t.Errorf("got local cache path %q, want empty", cfg.Granola.LocalCachePath)
	}
}

func TestLoad_DataSourceEnv(t *testing.T) {
	t.Setenv("ACAI_DATA_SOURCE", "local_cache")
	t.Setenv("ACAI_GRANOLA_CACHE_PATH", "/custom/cache-v3.json")

	cfg := config.Load()

	if cfg.Granola.DataSource != "local_cache" {
		t.Errorf("got data source %q, want local_cache", cfg.Granola.DataSource)
	}
	if cfg.Granola.LocalCachePath != "/custom/cache-v3.json" {
		t.Errorf("got local cache path %q", cfg.Granola.LocalCachePath)
	}
}

func TestLoad_FileOverridesDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfgPath := filepath.Join(home, ".acai", "config.yaml")
	fileCfg := config.FileConfig{
		DataSource: "local_cache",
		Granola: config.GranolaFileConfig{
			APIURL:    "https://file-api.example.com",
			CachePath: "/file/cache.json",
		},
	}
	if err := config.WriteConfigFile(cfgPath, fileCfg); err != nil {
		t.Fatalf("WriteConfigFile: %v", err)
	}

	cfg := config.Load()

	if cfg.Granola.DataSource != "local_cache" {
		t.Errorf("DataSource = %q, want local_cache", cfg.Granola.DataSource)
	}
	if cfg.Granola.APIURL != "https://file-api.example.com" {
		t.Errorf("APIURL = %q, want https://file-api.example.com", cfg.Granola.APIURL)
	}
	if cfg.Granola.LocalCachePath != "/file/cache.json" {
		t.Errorf("LocalCachePath = %q, want /file/cache.json", cfg.Granola.LocalCachePath)
	}
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfgPath := filepath.Join(home, ".acai", "config.yaml")
	fileCfg := config.FileConfig{
		DataSource: "local_cache",
		Granola: config.GranolaFileConfig{
			APIURL: "https://file-api.example.com",
		},
	}
	if err := config.WriteConfigFile(cfgPath, fileCfg); err != nil {
		t.Fatalf("WriteConfigFile: %v", err)
	}

	// Env vars should win over file
	t.Setenv("ACAI_DATA_SOURCE", "api")
	t.Setenv("ACAI_GRANOLA_API_URL", "https://env-api.example.com")

	cfg := config.Load()

	if cfg.Granola.DataSource != "api" {
		t.Errorf("DataSource = %q, want api (env should override file)", cfg.Granola.DataSource)
	}
	if cfg.Granola.APIURL != "https://env-api.example.com" {
		t.Errorf("APIURL = %q, want https://env-api.example.com (env should override file)", cfg.Granola.APIURL)
	}
}

func TestLoad_MissingFileUsesDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// No config file exists — defaults should be used

	cfg := config.Load()

	defaults := config.Default()
	if cfg.Granola.DataSource != defaults.Granola.DataSource {
		t.Errorf("DataSource = %q, want default %q", cfg.Granola.DataSource, defaults.Granola.DataSource)
	}
	if cfg.Granola.APIURL != defaults.Granola.APIURL {
		t.Errorf("APIURL = %q, want default %q", cfg.Granola.APIURL, defaults.Granola.APIURL)
	}
}


