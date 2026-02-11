package cli

import (
	"fmt"
	"text/tabwriter"

	meetingapp "github.com/felixgeelhaar/acai/internal/application/meeting"
	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
	"github.com/spf13/cobra"
)

func newActionCmd(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "action",
		Short: "View and manage action items",
		Long:  "List, complete, and update action items from meetings.",
	}

	cmd.AddCommand(
		newActionListCmd(deps),
		newActionCompleteCmd(deps),
		newActionUpdateCmd(deps),
	)
	return cmd
}

func newActionListCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "list <meeting_id>",
		Short: "List action items for a meeting",
		Long:  "Display all action items for a meeting with their completion status.",
		Example: "  acai action list meeting-001\n  acai action list meeting-001 --format json",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := deps.GetActionItems.Execute(cmd.Context(), meetingapp.GetActionItemsInput{
				MeetingID: domain.MeetingID(args[0]),
			})
			if err != nil {
				return fmt.Errorf("failed to list action items: %w", err)
			}

			if len(out.Items) == 0 {
				_, _ = fmt.Fprintln(deps.Out, "No action items found for this meeting.")
				return nil
			}

			switch flagFormat {
			case "json":
				return printJSON(deps, out.Items)
			default:
				w := tabwriter.NewWriter(deps.Out, 0, 0, 2, ' ', 0)
				_, _ = fmt.Fprintln(w, "ID\tSTATUS\tOWNER\tTEXT")
				for _, item := range out.Items {
					status := "open"
					if item.IsCompleted() {
						status = "done"
					}
					_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
						item.ID(), status, item.Owner(), item.Text())
				}
				return w.Flush()
			}
		},
	}
}

func newActionCompleteCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:     "complete <meeting_id> <action_item_id>",
		Short:   "Mark an action item as completed",
		Long:    "Mark a specific action item as completed. Use 'acai action list <meeting_id>' to see action items.",
		Example: "  acai action complete meeting-001 action-001",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if deps.CompleteActionItem == nil {
				return errLocalDBRequired
			}
			out, err := deps.CompleteActionItem.Execute(cmd.Context(), meetingapp.CompleteActionItemInput{
				MeetingID:    domain.MeetingID(args[0]),
				ActionItemID: domain.ActionItemID(args[1]),
			})
			if err != nil {
				return fmt.Errorf("failed to complete action item: %w", err)
			}
			_, _ = fmt.Fprintf(deps.Out, "Action item %s completed (text: %s)\n", out.Item.ID(), out.Item.Text())
			return nil
		},
	}
}

func newActionUpdateCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:     "update <meeting_id> <action_item_id> <text>",
		Short:   "Update an action item's text",
		Long:    "Replace the text of an existing action item. Use 'acai action list <meeting_id>' to see current action items and their IDs.",
		Example: "  acai action update meeting-001 action-001 \"Revised task description\"",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if deps.UpdateActionItem == nil {
				return errLocalDBRequired
			}
			out, err := deps.UpdateActionItem.Execute(cmd.Context(), meetingapp.UpdateActionItemInput{
				MeetingID:    domain.MeetingID(args[0]),
				ActionItemID: domain.ActionItemID(args[1]),
				Text:         args[2],
			})
			if err != nil {
				return fmt.Errorf("failed to update action item: %w", err)
			}
			_, _ = fmt.Fprintf(deps.Out, "Action item %s updated (text: %s)\n", out.Item.ID(), out.Item.Text())
			return nil
		},
	}
}
