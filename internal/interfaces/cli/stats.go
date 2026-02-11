package cli

import (
	"fmt"
	"text/tabwriter"

	meetingapp "github.com/felixgeelhaar/acai/internal/application/meeting"
	"github.com/spf13/cobra"
)

func newStatsCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show meeting statistics",
		Long: `Display aggregate statistics across your meetings including frequency,
platform distribution, top participants, action item completion rates,
and speaker talk time.`,
		Example: "  acai stats\n  acai stats --format json",
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := deps.GetMeetingStats.Execute(cmd.Context(), meetingapp.GetMeetingStatsInput{})
			if err != nil {
				return fmt.Errorf("failed to get stats: %w", err)
			}

			switch flagFormat {
			case "json":
				return printJSON(deps, out)
			default:
				return printStatsTable(deps, out)
			}
		},
	}
}

func printStatsTable(deps *Dependencies, stats *meetingapp.GetMeetingStatsOutput) error {
	_, _ = fmt.Fprintf(deps.Out, "Meeting Statistics (%s to %s)\n\n", stats.DateRange.Earliest, stats.DateRange.Latest)
	_, _ = fmt.Fprintf(deps.Out, "Total meetings: %d\n\n", stats.TotalMeetings)

	if len(stats.PlatformDistribution) > 0 {
		_, _ = fmt.Fprintln(deps.Out, "Platforms:")
		w := tabwriter.NewWriter(deps.Out, 0, 0, 2, ' ', 0)
		for _, p := range stats.PlatformDistribution {
			_, _ = fmt.Fprintf(w, "  %s\t%d\n", p.Source, p.Count)
		}
		_ = w.Flush()
		_, _ = fmt.Fprintln(deps.Out)
	}

	if len(stats.TopParticipants) > 0 {
		_, _ = fmt.Fprintln(deps.Out, "Top Participants:")
		w := tabwriter.NewWriter(deps.Out, 0, 0, 2, ' ', 0)
		for _, p := range stats.TopParticipants {
			_, _ = fmt.Fprintf(w, "  %s\t%d meetings\n", p.Name, p.MeetingCount)
		}
		_ = w.Flush()
		_, _ = fmt.Fprintln(deps.Out)
	}

	_, _ = fmt.Fprintf(deps.Out, "Action Items: %d total, %d completed (%.0f%%)\n",
		stats.ActionItems.Total, stats.ActionItems.Completed, stats.ActionItems.CompletionRate*100)
	_, _ = fmt.Fprintf(deps.Out, "Summary Coverage: %.0f%% of meetings have summaries\n",
		stats.SummaryCoverage.CoverageRate*100)

	if len(stats.SpeakerTalkTime) > 0 {
		_, _ = fmt.Fprintln(deps.Out)
		_, _ = fmt.Fprintln(deps.Out, "Top Speakers (by word count):")
		w := tabwriter.NewWriter(deps.Out, 0, 0, 2, ' ', 0)
		for _, s := range stats.SpeakerTalkTime {
			_, _ = fmt.Fprintf(w, "  %s\t%d words\t%d utterances\n", s.Speaker, s.WordCount, s.UtteranceCount)
		}
		_ = w.Flush()
	}

	return nil
}
