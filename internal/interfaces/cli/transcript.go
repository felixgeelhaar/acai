package cli

import (
	"fmt"
	"text/tabwriter"

	meetingapp "github.com/felixgeelhaar/acai/internal/application/meeting"
	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
	"github.com/spf13/cobra"
)

func newTranscriptCmd(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transcript",
		Short: "View and search meeting transcripts",
		Long:  "Show the full transcript of a meeting or search across all transcripts.",
	}

	cmd.AddCommand(
		newTranscriptShowCmd(deps),
		newTranscriptSearchCmd(deps),
	)
	return cmd
}

func newTranscriptShowCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "show <meeting_id>",
		Short: "Show the transcript for a meeting",
		Long:  "Display the full transcript with speaker names and timestamps.",
		Example: "  acai transcript show meeting-001\n  acai transcript show meeting-001 --format json",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := deps.GetTranscript.Execute(cmd.Context(), meetingapp.GetTranscriptInput{
				MeetingID: domain.MeetingID(args[0]),
			})
			if err != nil {
				return fmt.Errorf("failed to get transcript: %w", err)
			}

			if out.Transcript == nil {
				_, _ = fmt.Fprintln(deps.Out, "No transcript available for this meeting.")
				return nil
			}

			switch flagFormat {
			case "json":
				return printJSON(deps, out.Transcript)
			default:
				for _, u := range out.Transcript.Utterances() {
					ts := u.Timestamp().Format("15:04:05")
					_, _ = fmt.Fprintf(deps.Out, "[%s] %s: %s\n", ts, u.Speaker(), u.Text())
				}
				return nil
			}
		},
	}
}

func newTranscriptSearchCmd(deps *Dependencies) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search across meeting transcripts",
		Long:  "Find meetings whose transcripts contain the search query. Returns matching meetings.",
		Example: "  acai transcript search \"quarterly review\"\n  acai transcript search roadmap --limit 5",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := deps.SearchTranscripts.Execute(cmd.Context(), meetingapp.SearchTranscriptsInput{
				Query: args[0],
				Limit: limit,
			})
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}

			if len(out.Meetings) == 0 {
				_, _ = fmt.Fprintln(deps.Out, "No meetings found matching your query.")
				return nil
			}

			switch flagFormat {
			case "json":
				return printJSON(deps, out.Meetings)
			default:
				_, _ = fmt.Fprintf(deps.Out, "Found %d meeting(s):\n\n", out.Total)
				w := tabwriter.NewWriter(deps.Out, 0, 0, 2, ' ', 0)
				_, _ = fmt.Fprintln(w, "ID\tTITLE\tDATE\tSOURCE")
				for _, m := range out.Meetings {
					_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
						m.ID(), m.Title(), m.Datetime().Format("2006-01-02 15:04"), m.Source())
				}
				return w.Flush()
			}
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "Max results")
	return cmd
}
