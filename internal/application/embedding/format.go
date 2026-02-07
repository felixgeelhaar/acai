package embedding

import (
	"encoding/json"
	"strings"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

// ExportFormat defines how chunks are serialized for output.
type ExportFormat interface {
	FormatChunks(chunks []domain.Chunk) (string, error)
}

// JSONLLine is the serialization structure for a single JSONL line.
type JSONLLine struct {
	MeetingID  string `json:"meeting_id"`
	ChunkIndex int    `json:"chunk_index"`
	Content    string `json:"content"`
	Speaker    string `json:"speaker,omitempty"`
	StartTime  string `json:"start_time,omitempty"`
	EndTime    string `json:"end_time,omitempty"`
	Source     string `json:"source"`
	TokenCount int    `json:"token_count"`
}

// JSONLFormat serializes chunks as newline-delimited JSON.
type JSONLFormat struct{}

func (f *JSONLFormat) FormatChunks(chunks []domain.Chunk) (string, error) {
	var lines []string
	for _, c := range chunks {
		line := JSONLLine{
			MeetingID:  string(c.MeetingID()),
			ChunkIndex: c.ChunkIndex(),
			Content:    c.Content(),
			Speaker:    c.Speaker(),
			Source:     string(c.Source()),
			TokenCount: c.TokenCount(),
		}
		if !c.StartTime().IsZero() {
			line.StartTime = c.StartTime().Format(time.RFC3339)
		}
		if !c.EndTime().IsZero() {
			line.EndTime = c.EndTime().Format(time.RFC3339)
		}
		data, err := json.Marshal(line)
		if err != nil {
			return "", err
		}
		lines = append(lines, string(data))
	}
	return strings.Join(lines, "\n"), nil
}
