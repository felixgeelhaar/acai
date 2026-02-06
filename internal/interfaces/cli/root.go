package cli

import (
	"github.com/spf13/cobra"
)

var (
	flagFormat  string
	flagVerbose bool
)

func NewRootCmd(deps *Dependencies) *cobra.Command {
	root := &cobra.Command{
		Use:   "granola-mcp",
		Short: "Granola meeting intelligence for the MCP ecosystem",
		Long:  "A CLI and MCP server that exposes Granola meeting data as structured, queryable MCP resources.",
		SilenceUsage: true,
	}

	root.PersistentFlags().StringVar(&flagFormat, "format", "table", "Output format: table, json, md")
	root.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "Enable debug logging")

	root.AddCommand(
		newAuthCmd(deps),
		newSyncCmd(deps),
		newListCmd(deps),
		newExportCmd(deps),
		newServeCmd(deps),
		newVersionCmd(),
	)

	return root
}
