package embedding

import (
	"context"
	"errors"
	"fmt"

	"github.com/felixgeelhaar/granola-mcp/internal/domain/annotation"
	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

var (
	ErrNoMeetings       = errors.New("at least one meeting ID is required")
	ErrInvalidStrategy  = errors.New("unknown chunking strategy")
)

type ExportEmbeddingsInput struct {
	MeetingIDs []domain.MeetingID
	Strategy   string // "speaker_turn", "time_window", "token_limit"
	MaxTokens  int
	Format     string // "jsonl"
}

type ExportEmbeddingsOutput struct {
	Content    string
	ChunkCount int
}

type ExportEmbeddings struct {
	meetingRepo domain.Repository
	noteRepo    annotation.NoteRepository
}

func NewExportEmbeddings(meetingRepo domain.Repository, noteRepo annotation.NoteRepository) *ExportEmbeddings {
	return &ExportEmbeddings{meetingRepo: meetingRepo, noteRepo: noteRepo}
}

func (uc *ExportEmbeddings) Execute(ctx context.Context, input ExportEmbeddingsInput) (*ExportEmbeddingsOutput, error) {
	if len(input.MeetingIDs) == 0 {
		return nil, ErrNoMeetings
	}

	strategy, err := resolveStrategy(input.Strategy, input.MaxTokens)
	if err != nil {
		return nil, err
	}

	var allChunks []domain.Chunk

	for _, mid := range input.MeetingIDs {
		// Get transcript chunks
		transcript, err := uc.meetingRepo.GetTranscript(ctx, mid)
		if err != nil && !errors.Is(err, domain.ErrTranscriptNotReady) {
			return nil, fmt.Errorf("get transcript %s: %w", mid, err)
		}
		if transcript != nil {
			chunks, err := strategy.ChunkTranscript(mid, transcript.Utterances())
			if err != nil {
				return nil, fmt.Errorf("chunk transcript %s: %w", mid, err)
			}
			allChunks = append(allChunks, chunks...)
		}

		// Get meeting for summary
		meeting, err := uc.meetingRepo.FindByID(ctx, mid)
		if err != nil {
			return nil, fmt.Errorf("get meeting %s: %w", mid, err)
		}
		if summary := meeting.Summary(); summary != nil && summary.Content() != "" {
			chunkIdx := len(allChunks)
			c, err := domain.NewChunk(mid, chunkIdx, summary.Content(), "", meeting.Datetime(), meeting.Datetime(), domain.ChunkSourceSummary, estimateTokens(summary.Content()))
			if err != nil {
				return nil, err
			}
			allChunks = append(allChunks, c)
		}

		// Get agent notes
		if uc.noteRepo != nil {
			notes, err := uc.noteRepo.ListByMeeting(ctx, string(mid))
			if err != nil {
				return nil, fmt.Errorf("list notes %s: %w", mid, err)
			}
			for _, n := range notes {
				chunkIdx := len(allChunks)
				c, err := domain.NewChunk(mid, chunkIdx, n.Content(), n.Author(), n.CreatedAt(), n.CreatedAt(), domain.ChunkSourceNote, estimateTokens(n.Content()))
				if err != nil {
					return nil, err
				}
				allChunks = append(allChunks, c)
			}
		}
	}

	formatter := resolveFormat(input.Format)
	content, err := formatter.FormatChunks(allChunks)
	if err != nil {
		return nil, fmt.Errorf("format chunks: %w", err)
	}

	return &ExportEmbeddingsOutput{
		Content:    content,
		ChunkCount: len(allChunks),
	}, nil
}

func resolveStrategy(name string, maxTokens int) (ChunkStrategy, error) {
	switch name {
	case "", "speaker_turn":
		return &BySpeakerTurn{}, nil
	case "time_window":
		return &ByTimeWindow{}, nil
	case "token_limit":
		return &ByTokenLimit{MaxTokens: maxTokens}, nil
	default:
		return nil, ErrInvalidStrategy
	}
}

func resolveFormat(name string) ExportFormat {
	// Only JSONL is supported for now; default to it.
	return &JSONLFormat{}
}
