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
		Use:   "acai",
		Short: "Granola meeting intelligence for the MCP ecosystem",
		Long:  "A CLI and MCP server that exposes Granola meeting data as structured, queryable MCP resources.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVar(&flagFormat, "format", "table", "Output format: table, json, md")
	root.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "Enable debug logging")

	root.AddCommand(
		newInitCmd(),
		newAuthCmd(deps),
		newMeetingCmd(deps),
		newTranscriptCmd(deps),
		newNoteCmd(deps),
		newActionCmd(deps),
		newStatsCmd(deps),
		newExportCmd(deps),
		newSyncCmd(deps),
		newServeCmd(deps),
		newVersionCmd(),
	)

	return root
}
