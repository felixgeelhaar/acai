package embedding

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

func TestJSONLFormat_Empty(t *testing.T) {
	f := &JSONLFormat{}
	content, err := f.FormatChunks(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "" {
		t.Errorf("expected empty string, got %q", content)
	}
}

func TestJSONLFormat_SingleChunk(t *testing.T) {
	f := &JSONLFormat{}
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	c, _ := domain.NewChunk("m-1", 0, "Hello world", "Alice", now, now.Add(5*time.Second), domain.ChunkSourceTranscript, 3)

	content, err := f.FormatChunks([]domain.Chunk{c})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var line JSONLLine
	if err := json.Unmarshal([]byte(content), &line); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if line.MeetingID != "m-1" {
		t.Errorf("meeting_id = %q, want m-1", line.MeetingID)
	}
	if line.ChunkIndex != 0 {
		t.Errorf("chunk_index = %d, want 0", line.ChunkIndex)
	}
	if line.Content != "Hello world" {
		t.Errorf("content = %q", line.Content)
	}
	if line.Speaker != "Alice" {
		t.Errorf("speaker = %q", line.Speaker)
	}
	if line.Source != "transcript" {
		t.Errorf("source = %q", line.Source)
	}
	if line.TokenCount != 3 {
		t.Errorf("token_count = %d", line.TokenCount)
	}
}

func TestJSONLFormat_MultipleChunks(t *testing.T) {
	f := &JSONLFormat{}
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	c1, _ := domain.NewChunk("m-1", 0, "First", "Alice", now, now, domain.ChunkSourceTranscript, 1)
	c2, _ := domain.NewChunk("m-1", 1, "Second", "Bob", now, now, domain.ChunkSourceTranscript, 1)

	content, err := f.FormatChunks([]domain.Chunk{c1, c2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(content, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		var j JSONLLine
		if err := json.Unmarshal([]byte(line), &j); err != nil {
			t.Errorf("line %d invalid JSON: %v", i, err)
		}
	}
}

func TestJSONLFormat_OmitsEmptySpeaker(t *testing.T) {
	f := &JSONLFormat{}
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	c, _ := domain.NewChunk("m-1", 0, "Summary text", "", now, now, domain.ChunkSourceSummary, 2)

	content, err := f.FormatChunks([]domain.Chunk{c})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(content, `"speaker"`) {
		t.Errorf("expected speaker to be omitted, got: %s", content)
	}
}

func TestJSONLFormat_ZeroTimesOmitted(t *testing.T) {
	f := &JSONLFormat{}
	c, _ := domain.NewChunk("m-1", 0, "Note text", "agent", time.Time{}, time.Time{}, domain.ChunkSourceNote, 2)

	content, err := f.FormatChunks([]domain.Chunk{c})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(content, `"start_time"`) {
		t.Errorf("expected start_time to be omitted, got: %s", content)
	}
	if strings.Contains(content, `"end_time"`) {
		t.Errorf("expected end_time to be omitted, got: %s", content)
	}
}
