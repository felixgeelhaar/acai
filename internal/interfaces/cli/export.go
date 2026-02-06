package cli

import (
	"fmt"

	exportapp "github.com/felixgeelhaar/granola-mcp/internal/application/export"
	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
	"github.com/spf13/cobra"
)

func newExportCmd(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export meeting data",
	}

	cmd.AddCommand(newExportMeetingCmd(deps))
	return cmd
}

func newExportMeetingCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "meeting [id]",
		Short: "Export a meeting",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format := exportapp.Format(flagFormat)
			if flagFormat == "table" {
				format = exportapp.FormatMarkdown
			}

			out, err := deps.ExportMeeting.Execute(cmd.Context(), exportapp.ExportMeetingInput{
				MeetingID: domain.MeetingID(args[0]),
				Format:    format,
			})
			if err != nil {
				return fmt.Errorf("export failed: %w", err)
			}

			fmt.Fprint(deps.Out, out.Content)
			return nil
		},
	}
}
