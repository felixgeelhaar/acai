package meeting

import (
	"testing"
	"time"
)

func TestNewChunk_Valid(t *testing.T) {
	now := time.Now().UTC()
	c, err := NewChunk("m-1", 0, "Hello world", "Alice", now, now.Add(10*time.Second), ChunkSourceTranscript, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.MeetingID() != "m-1" {
		t.Errorf("MeetingID = %q, want %q", c.MeetingID(), "m-1")
	}
	if c.ChunkIndex() != 0 {
		t.Errorf("ChunkIndex = %d, want 0", c.ChunkIndex())
	}
	if c.Content() != "Hello world" {
		t.Errorf("Content = %q, want %q", c.Content(), "Hello world")
	}
	if c.Speaker() != "Alice" {
		t.Errorf("Speaker = %q, want %q", c.Speaker(), "Alice")
	}
	if c.Source() != ChunkSourceTranscript {
		t.Errorf("Source = %q, want %q", c.Source(), ChunkSourceTranscript)
	}
	if c.TokenCount() != 3 {
		t.Errorf("TokenCount = %d, want 3", c.TokenCount())
	}
}

func TestNewChunk_AllSources(t *testing.T) {
	now := time.Now().UTC()
	sources := []ChunkSource{ChunkSourceTranscript, ChunkSourceSummary, ChunkSourceNote}
	for _, src := range sources {
		_, err := NewChunk("m-1", 0, "content", "", now, now, src, 1)
		if err != nil {
			t.Errorf("unexpected error for source %q: %v", src, err)
		}
	}
}

func TestNewChunk_EmptyContent(t *testing.T) {
	now := time.Now().UTC()
	_, err := NewChunk("m-1", 0, "", "Alice", now, now, ChunkSourceTranscript, 1)
	if err != ErrInvalidChunkContent {
		t.Errorf("expected ErrInvalidChunkContent, got %v", err)
	}
}

func TestNewChunk_InvalidSource(t *testing.T) {
	now := time.Now().UTC()
	_, err := NewChunk("m-1", 0, "content", "Alice", now, now, "invalid", 1)
	if err != ErrInvalidChunkSource {
		t.Errorf("expected ErrInvalidChunkSource, got %v", err)
	}
}

func TestNewChunk_NegativeTokenCount(t *testing.T) {
	now := time.Now().UTC()
	_, err := NewChunk("m-1", 0, "content", "Alice", now, now, ChunkSourceTranscript, -1)
	if err != ErrInvalidTokenCount {
		t.Errorf("expected ErrInvalidTokenCount, got %v", err)
	}
}

func TestNewChunk_ZeroTokenCount(t *testing.T) {
	now := time.Now().UTC()
	c, err := NewChunk("m-1", 0, "content", "", now, now, ChunkSourceSummary, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.TokenCount() != 0 {
		t.Errorf("TokenCount = %d, want 0", c.TokenCount())
	}
}

func TestNewChunk_EmptySpeaker(t *testing.T) {
	now := time.Now().UTC()
	c, err := NewChunk("m-1", 0, "summary text", "", now, now, ChunkSourceSummary, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Speaker() != "" {
		t.Errorf("Speaker = %q, want empty", c.Speaker())
	}
}
