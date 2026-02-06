package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func newServeCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server",
		Long:  "Start the Granola MCP server. By default serves over stdio for use with Claude Code and other MCP clients.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if deps.MCPServer == nil {
				return fmt.Errorf("MCP server not configured")
			}

			fmt.Fprintf(deps.Out, "Starting %s v%s MCP server (stdio)...\n",
				deps.MCPServer.Name(), deps.MCPServer.Version())

			ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			if err := deps.MCPServer.ServeStdio(ctx); err != nil {
				// Context cancellation is expected on shutdown
				if ctx.Err() != nil {
					fmt.Fprintln(os.Stderr, "MCP server stopped.")
					return nil
				}
				return fmt.Errorf("MCP server error: %w", err)
			}

			return nil
		},
	}
}
