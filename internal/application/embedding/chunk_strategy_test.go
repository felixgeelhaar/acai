package embedding

import (
	"testing"
	"time"

	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

func TestBySpeakerTurn_Empty(t *testing.T) {
	s := &BySpeakerTurn{}
	chunks, err := s.ChunkTranscript("m-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chunks != nil {
		t.Errorf("expected nil chunks, got %d", len(chunks))
	}
}

func TestBySpeakerTurn_SingleSpeaker(t *testing.T) {
	s := &BySpeakerTurn{}
	now := time.Now().UTC()
	utterances := []domain.Utterance{
		domain.NewUtterance("Alice", "Hello", now, 0.9),
		domain.NewUtterance("Alice", "How are you?", now.Add(5*time.Second), 0.95),
	}
	chunks, err := s.ChunkTranscript("m-1", utterances)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Speaker() != "Alice" {
		t.Errorf("Speaker = %q, want Alice", chunks[0].Speaker())
	}
	if chunks[0].Content() != "Hello How are you?" {
		t.Errorf("Content = %q", chunks[0].Content())
	}
}

func TestBySpeakerTurn_MultipleSpeakers(t *testing.T) {
	s := &BySpeakerTurn{}
	now := time.Now().UTC()
	utterances := []domain.Utterance{
		domain.NewUtterance("Alice", "Hi", now, 0.9),
		domain.NewUtterance("Alice", "Good morning", now.Add(1*time.Second), 0.9),
		domain.NewUtterance("Bob", "Hey", now.Add(2*time.Second), 0.95),
		domain.NewUtterance("Alice", "Let's start", now.Add(3*time.Second), 0.9),
	}
	chunks, err := s.ChunkTranscript("m-1", utterances)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	if chunks[0].Speaker() != "Alice" {
		t.Errorf("chunk 0 speaker = %q, want Alice", chunks[0].Speaker())
	}
	if chunks[1].Speaker() != "Bob" {
		t.Errorf("chunk 1 speaker = %q, want Bob", chunks[1].Speaker())
	}
	if chunks[2].Speaker() != "Alice" {
		t.Errorf("chunk 2 speaker = %q, want Alice", chunks[2].Speaker())
	}
	// Verify chunk indices
	for i, c := range chunks {
		if c.ChunkIndex() != i {
			t.Errorf("chunk %d index = %d", i, c.ChunkIndex())
		}
		if c.Source() != domain.ChunkSourceTranscript {
			t.Errorf("chunk %d source = %q, want transcript", i, c.Source())
		}
	}
}

func TestByTimeWindow_Empty(t *testing.T) {
	s := &ByTimeWindow{Window: 30 * time.Second}
	chunks, err := s.ChunkTranscript("m-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chunks != nil {
		t.Errorf("expected nil chunks, got %d", len(chunks))
	}
}

func TestByTimeWindow_SingleWindow(t *testing.T) {
	s := &ByTimeWindow{Window: 60 * time.Second}
	now := time.Now().UTC()
	utterances := []domain.Utterance{
		domain.NewUtterance("Alice", "Hello", now, 0.9),
		domain.NewUtterance("Bob", "Hi", now.Add(10*time.Second), 0.95),
		domain.NewUtterance("Alice", "Start", now.Add(20*time.Second), 0.9),
	}
	chunks, err := s.ChunkTranscript("m-1", utterances)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0].Content() != "Hello Hi Start" {
		t.Errorf("Content = %q", chunks[0].Content())
	}
}

func TestByTimeWindow_MultipleWindows(t *testing.T) {
	s := &ByTimeWindow{Window: 10 * time.Second}
	now := time.Now().UTC()
	utterances := []domain.Utterance{
		domain.NewUtterance("Alice", "Hello", now, 0.9),
		domain.NewUtterance("Bob", "Hi", now.Add(5*time.Second), 0.95),
		domain.NewUtterance("Alice", "Later", now.Add(30*time.Second), 0.9),
	}
	chunks, err := s.ChunkTranscript("m-1", utterances)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
}

func TestByTimeWindow_DefaultWindow(t *testing.T) {
	s := &ByTimeWindow{} // Window is zero — should default to 30s
	now := time.Now().UTC()
	utterances := []domain.Utterance{
		domain.NewUtterance("Alice", "Hello", now, 0.9),
	}
	chunks, err := s.ChunkTranscript("m-1", utterances)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
}

func TestByTokenLimit_Empty(t *testing.T) {
	s := &ByTokenLimit{MaxTokens: 100}
	chunks, err := s.ChunkTranscript("m-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chunks != nil {
		t.Errorf("expected nil chunks, got %d", len(chunks))
	}
}

func TestByTokenLimit_SplitsLargeContent(t *testing.T) {
	s := &ByTokenLimit{MaxTokens: 5}
	now := time.Now().UTC()
	// Each utterance has several words, token limit is low so should create multiple chunks
	utterances := []domain.Utterance{
		domain.NewUtterance("Alice", "This is a test sentence with many words", now, 0.9),
		domain.NewUtterance("Bob", "Another sentence with more words here", now.Add(5*time.Second), 0.95),
		domain.NewUtterance("Alice", "Final words", now.Add(10*time.Second), 0.9),
	}
	chunks, err := s.ChunkTranscript("m-1", utterances)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) < 2 {
		t.Errorf("expected at least 2 chunks, got %d", len(chunks))
	}
	for _, c := range chunks {
		if c.Source() != domain.ChunkSourceTranscript {
			t.Errorf("source = %q, want transcript", c.Source())
		}
	}
}

func TestByTokenLimit_DefaultMaxTokens(t *testing.T) {
	s := &ByTokenLimit{} // MaxTokens is zero — should default to 256
	now := time.Now().UTC()
	utterances := []domain.Utterance{
		domain.NewUtterance("Alice", "Short", now, 0.9),
	}
	chunks, err := s.ChunkTranscript("m-1", utterances)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		text string
		min  int
	}{
		{"", 0},
		{"hello", 1},
		{"hello world how are you", 3}, // 5 words / 0.75 ≈ 6.67
	}
	for _, tt := range tests {
		got := estimateTokens(tt.text)
		if got < tt.min {
			t.Errorf("estimateTokens(%q) = %d, want >= %d", tt.text, got, tt.min)
		}
	}
}
