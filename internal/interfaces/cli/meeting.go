package cli

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	exportapp "github.com/felixgeelhaar/acai/internal/application/export"
	meetingapp "github.com/felixgeelhaar/acai/internal/application/meeting"
	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
	"github.com/spf13/cobra"
)

func printJSON(deps *Dependencies, v interface{}) error {
	enc := json.NewEncoder(deps.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func newMeetingCmd(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "meeting",
		Short: "View and manage meetings",
		Long:  "List, view, and export meetings from Granola.",
	}

	cmd.AddCommand(
		newMeetingListCmd(deps),
		newMeetingShowCmd(deps),
		newMeetingExportCmd(deps),
	)
	return cmd
}

func newMeetingListCmd(deps *Dependencies) *cobra.Command {
	var (
		limit  int
		offset int
		source string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List meetings",
		Long:  "List meetings with optional filtering by source and pagination.",
		Example: "  acai meeting list\n  acai meeting list --limit 10 --source zoom",
		RunE: func(cmd *cobra.Command, args []string) error {
			input := meetingapp.ListMeetingsInput{
				Limit:  limit,
				Offset: offset,
			}
			if source != "" {
				input.Source = &source
			}

			out, err := deps.ListMeetings.Execute(cmd.Context(), input)
			if err != nil {
				return fmt.Errorf("failed to list meetings: %w", err)
			}

			switch flagFormat {
			case "json":
				return printJSON(deps, out.Meetings)
			default:
				return printMeetingsTable(deps, out.Meetings)
			}
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "Max results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Pagination offset")
	cmd.Flags().StringVar(&source, "source", "", "Filter by source (zoom, google_meet, teams)")

	return cmd
}

func newMeetingShowCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "show <meeting_id>",
		Short: "Show meeting details",
		Long: `Display a meeting's title, date, source, participants, summary, and action items.

Use 'acai transcript show <meeting_id>' to view the full transcript.`,
		Example: "  acai meeting show meeting-001\n  acai meeting show meeting-001 --format json",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := deps.GetMeeting.Execute(cmd.Context(), meetingapp.GetMeetingInput{
				ID: domain.MeetingID(args[0]),
			})
			if err != nil {
				return fmt.Errorf("failed to get meeting: %w", err)
			}

			m := out.Meeting
			switch flagFormat {
			case "json":
				return printJSON(deps, m)
			default:
				_, _ = fmt.Fprintf(deps.Out, "Title:   %s\n", m.Title())
				_, _ = fmt.Fprintf(deps.Out, "ID:      %s\n", m.ID())
				_, _ = fmt.Fprintf(deps.Out, "Date:    %s\n", m.Datetime().Format("2006-01-02 15:04"))
				_, _ = fmt.Fprintf(deps.Out, "Source:  %s\n", m.Source())

				if len(m.Participants()) > 0 {
					_, _ = fmt.Fprintln(deps.Out)
					_, _ = fmt.Fprintln(deps.Out, "Participants:")
					for _, p := range m.Participants() {
						if p.Email() != "" {
							_, _ = fmt.Fprintf(deps.Out, "  - %s <%s> (%s)\n", p.Name(), p.Email(), p.Role())
						} else {
							_, _ = fmt.Fprintf(deps.Out, "  - %s (%s)\n", p.Name(), p.Role())
						}
					}
				}

				if m.Summary() != nil {
					_, _ = fmt.Fprintln(deps.Out)
					_, _ = fmt.Fprintln(deps.Out, "Summary:")
					_, _ = fmt.Fprintln(deps.Out, m.Summary().Content())
				}

				if len(m.ActionItems()) > 0 {
					_, _ = fmt.Fprintln(deps.Out)
					_, _ = fmt.Fprintln(deps.Out, "Action Items:")
					for _, item := range m.ActionItems() {
						status := "[ ]"
						if item.IsCompleted() {
							status = "[x]"
						}
						_, _ = fmt.Fprintf(deps.Out, "  %s %s", status, item.Text())
						if item.Owner() != "" {
							_, _ = fmt.Fprintf(deps.Out, " (@%s)", item.Owner())
						}
						_, _ = fmt.Fprintln(deps.Out)
					}
				}

				return nil
			}
		},
	}
}

func newMeetingExportCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "export <meeting_id>",
		Short: "Export a meeting as markdown or JSON",
		Long: `Export a meeting's full content including title, participants, summary, transcript, and action items.

Defaults to markdown format. Use --format json for structured output.`,
		Example: "  acai meeting export meeting-001\n  acai meeting export meeting-001 --format json",
		Args:    cobra.ExactArgs(1),
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

			_, _ = fmt.Fprint(deps.Out, out.Content)
			return nil
		},
	}
}

func printMeetingsTable(deps *Dependencies, meetings []*domain.Meeting) error {
	w := tabwriter.NewWriter(deps.Out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tTITLE\tDATE\tSOURCE")
	for _, m := range meetings {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			m.ID(), m.Title(), m.Datetime().Format("2006-01-02 15:04"), m.Source())
	}
	return w.Flush()
}
