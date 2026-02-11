package cli

import (
	"fmt"
	"strings"

	embeddingapp "github.com/felixgeelhaar/acai/internal/application/embedding"
	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
	"github.com/spf13/cobra"
)

func newExportCmd(deps *Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export data for external tools",
		Long:  "Export meeting data in formats suitable for external tools and pipelines.",
	}

	cmd.AddCommand(
		newExportEmbeddingsCmd(deps),
	)
	return cmd
}

func newExportEmbeddingsCmd(deps *Dependencies) *cobra.Command {
	var (
		meetings  string
		strategy  string
		maxTokens int
	)

	cmd := &cobra.Command{
		Use:   "embeddings",
		Short: "Export meeting content as chunks for embedding generation",
		Long: `Export meeting transcripts and notes as text chunks suitable for vector embedding.

Supports multiple chunking strategies: speaker_turn (one chunk per speaker),
time_window (fixed time intervals), or token_limit (max tokens per chunk).`,
		Example: "  acai export embeddings --meetings meeting-001,meeting-002\n  acai export embeddings --meetings meeting-001 --strategy token_limit --max-tokens 512",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if deps.ExportEmbeddings == nil {
				return fmt.Errorf("embedding export functionality not configured")
			}
			if meetings == "" {
				return fmt.Errorf("--meetings is required")
			}

			ids := strings.Split(meetings, ",")
			meetingIDs := make([]domain.MeetingID, len(ids))
			for i, id := range ids {
				meetingIDs[i] = domain.MeetingID(strings.TrimSpace(id))
			}

			out, err := deps.ExportEmbeddings.Execute(cmd.Context(), embeddingapp.ExportEmbeddingsInput{
				MeetingIDs: meetingIDs,
				Strategy:   strategy,
				MaxTokens:  maxTokens,
			})
			if err != nil {
				return fmt.Errorf("export failed: %w", err)
			}

			_, _ = fmt.Fprintln(deps.Out, out.Content)
			_, _ = fmt.Fprintf(deps.Out, "# %d chunks exported\n", out.ChunkCount)
			return nil
		},
	}

	cmd.Flags().StringVar(&meetings, "meetings", "", "Comma-separated meeting IDs")
	cmd.Flags().StringVar(&strategy, "strategy", "speaker_turn", "Chunking strategy: speaker_turn, time_window, token_limit")
	cmd.Flags().IntVar(&maxTokens, "max-tokens", 256, "Max tokens per chunk (for token_limit strategy)")
	return cmd
}
