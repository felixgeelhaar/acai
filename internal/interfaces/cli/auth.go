package cli

import (
	"fmt"

	authapp "github.com/felixgeelhaar/acai/internal/application/auth"
	domain "github.com/felixgeelhaar/acai/internal/domain/auth"
	"github.com/spf13/cobra"
)

func newAuthCmd(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication with Granola",
	}

	cmd.AddCommand(newAuthLoginCmd(deps))
	cmd.AddCommand(newAuthStatusCmd(deps))
	cmd.AddCommand(newAuthLogoutCmd(deps))

	return cmd
}

func newAuthLoginCmd(deps *Dependencies) *cobra.Command {
	var method string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Granola (requires ACAI_GRANOLA_API_TOKEN env var)",
		Long: `Authenticate with Granola using an API token.

Set the ACAI_GRANOLA_API_TOKEN environment variable before running this command.
OAuth is not yet supported.

Example:
  export ACAI_GRANOLA_API_TOKEN=gra_xxxxx
  acai auth login`,
		RunE: func(cmd *cobra.Command, args []string) error {
			authMethod := domain.AuthAPIToken
			if method == "oauth" {
				authMethod = domain.AuthOAuth
			}

			out, err := deps.Login.Execute(cmd.Context(), authapp.LoginInput{
				Method:   authMethod,
				APIToken: deps.GranolaAPIToken,
			})
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			_, _ = fmt.Fprintf(deps.Out, "Authenticated successfully (workspace: %s)\n", out.Credential.Workspace())
			return nil
		},
	}

	cmd.Flags().StringVar(&method, "method", "api_token", "Auth method: api_token (oauth not yet supported)")

	return cmd
}

func newAuthLogoutCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deps.Logout == nil {
				return fmt.Errorf("logout not configured")
			}
			if err := deps.Logout.Execute(cmd.Context()); err != nil {
				return fmt.Errorf("logout failed: %w", err)
			}
			_, _ = fmt.Fprintln(deps.Out, "Logged out successfully. Stored credentials have been removed.")
			return nil
		},
	}
}

func newAuthStatusCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := deps.CheckStatus.Execute(cmd.Context())
			if err != nil {
				return err
			}

			if !out.Authenticated {
				_, _ = fmt.Fprintln(deps.Out, "Not authenticated. Run 'acai auth login' to authenticate.")
				return nil
			}

			_, _ = fmt.Fprintf(deps.Out, "Authenticated (workspace: %s, method: %s)\n",
				out.Credential.Workspace(), out.Credential.Method())
			return nil
		},
	}
}
