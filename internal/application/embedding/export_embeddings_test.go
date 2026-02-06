package embedding

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/felixgeelhaar/granola-mcp/internal/domain/annotation"
	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

// --- Mocks ---

type mockMeetingRepo struct {
	meetings    map[domain.MeetingID]*domain.Meeting
	transcripts map[domain.MeetingID]*domain.Transcript
}

func (m *mockMeetingRepo) FindByID(_ context.Context, id domain.MeetingID) (*domain.Meeting, error) {
	if mtg, ok := m.meetings[id]; ok {
		return mtg, nil
	}
	return nil, domain.ErrMeetingNotFound
}

func (m *mockMeetingRepo) GetTranscript(_ context.Context, id domain.MeetingID) (*domain.Transcript, error) {
	if t, ok := m.transcripts[id]; ok {
		return t, nil
	}
	return nil, domain.ErrTranscriptNotReady
}

func (m *mockMeetingRepo) List(_ context.Context, _ domain.ListFilter) ([]*domain.Meeting, error) {
	return nil, nil
}
func (m *mockMeetingRepo) SearchTranscripts(_ context.Context, _ string, _ domain.ListFilter) ([]*domain.Meeting, error) {
	return nil, nil
}
func (m *mockMeetingRepo) GetActionItems(_ context.Context, _ domain.MeetingID) ([]*domain.ActionItem, error) {
	return nil, nil
}
func (m *mockMeetingRepo) Sync(_ context.Context, _ *time.Time) ([]domain.DomainEvent, error) {
	return nil, nil
}

type mockNoteRepo struct {
	notes map[string][]*annotation.AgentNote
}

func (m *mockNoteRepo) Save(_ context.Context, _ *annotation.AgentNote) error { return nil }
func (m *mockNoteRepo) FindByID(_ context.Context, _ annotation.NoteID) (*annotation.AgentNote, error) {
	return nil, annotation.ErrNoteNotFound
}
func (m *mockNoteRepo) ListByMeeting(_ context.Context, meetingID string) ([]*annotation.AgentNote, error) {
	return m.notes[meetingID], nil
}
func (m *mockNoteRepo) Delete(_ context.Context, _ annotation.NoteID) error { return nil }

// --- Tests ---

func TestExportEmbeddings_NoMeetings(t *testing.T) {
	uc := NewExportEmbeddings(&mockMeetingRepo{}, nil)
	_, err := uc.Execute(context.Background(), ExportEmbeddingsInput{})
	if err != ErrNoMeetings {
		t.Errorf("expected ErrNoMeetings, got %v", err)
	}
}

func TestExportEmbeddings_InvalidStrategy(t *testing.T) {
	uc := NewExportEmbeddings(&mockMeetingRepo{}, nil)
	_, err := uc.Execute(context.Background(), ExportEmbeddingsInput{
		MeetingIDs: []domain.MeetingID{"m-1"},
		Strategy:   "bogus",
	})
	if err != ErrInvalidStrategy {
		t.Errorf("expected ErrInvalidStrategy, got %v", err)
	}
}

func TestExportEmbeddings_TranscriptOnly(t *testing.T) {
	now := time.Now().UTC()
	mtg, _ := domain.New("m-1", "Sprint Planning", now, domain.SourceZoom, nil)
	mtg.ClearDomainEvents()

	utts := []domain.Utterance{
		domain.NewUtterance("Alice", "Hello", now, 0.9),
		domain.NewUtterance("Bob", "Hi there", now.Add(5*time.Second), 0.95),
	}
	transcript := domain.NewTranscript("m-1", utts)

	repo := &mockMeetingRepo{
		meetings:    map[domain.MeetingID]*domain.Meeting{"m-1": mtg},
		transcripts: map[domain.MeetingID]*domain.Transcript{"m-1": &transcript},
	}

	uc := NewExportEmbeddings(repo, nil)
	out, err := uc.Execute(context.Background(), ExportEmbeddingsInput{
		MeetingIDs: []domain.MeetingID{"m-1"},
		Strategy:   "speaker_turn",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ChunkCount != 2 { // Alice + Bob = 2 speaker turns
		t.Errorf("ChunkCount = %d, want 2", out.ChunkCount)
	}
	if !strings.Contains(out.Content, "Hello") {
		t.Errorf("content should contain 'Hello': %s", out.Content)
	}
}

func TestExportEmbeddings_WithSummary(t *testing.T) {
	now := time.Now().UTC()
	mtg, _ := domain.New("m-1", "Sprint Planning", now, domain.SourceZoom, nil)
	summary := domain.NewSummary("m-1", "This is the summary", domain.SummaryAuto)
	mtg.AttachSummary(summary)
	mtg.ClearDomainEvents()

	repo := &mockMeetingRepo{
		meetings:    map[domain.MeetingID]*domain.Meeting{"m-1": mtg},
		transcripts: map[domain.MeetingID]*domain.Transcript{},
	}

	uc := NewExportEmbeddings(repo, nil)
	out, err := uc.Execute(context.Background(), ExportEmbeddingsInput{
		MeetingIDs: []domain.MeetingID{"m-1"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ChunkCount != 1 { // Just the summary
		t.Errorf("ChunkCount = %d, want 1", out.ChunkCount)
	}
	if !strings.Contains(out.Content, "summary") {
		t.Errorf("content should contain 'summary': %s", out.Content)
	}
}

func TestExportEmbeddings_WithNotes(t *testing.T) {
	now := time.Now().UTC()
	mtg, _ := domain.New("m-1", "Sprint", now, domain.SourceZoom, nil)
	mtg.ClearDomainEvents()

	note := annotation.ReconstructAgentNote("n-1", "m-1", "agent", "Agent observation", now)

	repo := &mockMeetingRepo{
		meetings:    map[domain.MeetingID]*domain.Meeting{"m-1": mtg},
		transcripts: map[domain.MeetingID]*domain.Transcript{},
	}
	noteRepo := &mockNoteRepo{
		notes: map[string][]*annotation.AgentNote{"m-1": {note}},
	}

	uc := NewExportEmbeddings(repo, noteRepo)
	out, err := uc.Execute(context.Background(), ExportEmbeddingsInput{
		MeetingIDs: []domain.MeetingID{"m-1"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ChunkCount != 1 { // Just the note
		t.Errorf("ChunkCount = %d, want 1", out.ChunkCount)
	}
	if !strings.Contains(out.Content, "note") {
		t.Errorf("content should contain source 'note': %s", out.Content)
	}
}

func TestExportEmbeddings_MultipleMeetings(t *testing.T) {
	now := time.Now().UTC()
	mtg1, _ := domain.New("m-1", "Meeting 1", now, domain.SourceZoom, nil)
	mtg1.ClearDomainEvents()
	mtg2, _ := domain.New("m-2", "Meeting 2", now, domain.SourceZoom, nil)
	mtg2.ClearDomainEvents()

	utts := []domain.Utterance{
		domain.NewUtterance("Alice", "Hello", now, 0.9),
	}
	t1 := domain.NewTranscript("m-1", utts)
	t2 := domain.NewTranscript("m-2", utts)

	repo := &mockMeetingRepo{
		meetings: map[domain.MeetingID]*domain.Meeting{
			"m-1": mtg1,
			"m-2": mtg2,
		},
		transcripts: map[domain.MeetingID]*domain.Transcript{
			"m-1": &t1,
			"m-2": &t2,
		},
	}

	uc := NewExportEmbeddings(repo, nil)
	out, err := uc.Execute(context.Background(), ExportEmbeddingsInput{
		MeetingIDs: []domain.MeetingID{"m-1", "m-2"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ChunkCount != 2 { // One chunk per meeting
		t.Errorf("ChunkCount = %d, want 2", out.ChunkCount)
	}
}

func TestExportEmbeddings_TimeWindowStrategy(t *testing.T) {
	now := time.Now().UTC()
	mtg, _ := domain.New("m-1", "Sprint", now, domain.SourceZoom, nil)
	mtg.ClearDomainEvents()

	utts := []domain.Utterance{
		domain.NewUtterance("Alice", "Hello", now, 0.9),
		domain.NewUtterance("Bob", "Hi", now.Add(5*time.Second), 0.95),
	}
	transcript := domain.NewTranscript("m-1", utts)

	repo := &mockMeetingRepo{
		meetings:    map[domain.MeetingID]*domain.Meeting{"m-1": mtg},
		transcripts: map[domain.MeetingID]*domain.Transcript{"m-1": &transcript},
	}

	uc := NewExportEmbeddings(repo, nil)
	out, err := uc.Execute(context.Background(), ExportEmbeddingsInput{
		MeetingIDs: []domain.MeetingID{"m-1"},
		Strategy:   "time_window",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ChunkCount < 1 {
		t.Errorf("expected at least 1 chunk, got %d", out.ChunkCount)
	}
}

func TestExportEmbeddings_TokenLimitStrategy(t *testing.T) {
	now := time.Now().UTC()
	mtg, _ := domain.New("m-1", "Sprint", now, domain.SourceZoom, nil)
	mtg.ClearDomainEvents()

	utts := []domain.Utterance{
		domain.NewUtterance("Alice", "This is a longer utterance with many words", now, 0.9),
		domain.NewUtterance("Bob", "Another sentence also with many words here", now.Add(5*time.Second), 0.95),
	}
	transcript := domain.NewTranscript("m-1", utts)

	repo := &mockMeetingRepo{
		meetings:    map[domain.MeetingID]*domain.Meeting{"m-1": mtg},
		transcripts: map[domain.MeetingID]*domain.Transcript{"m-1": &transcript},
	}

	uc := NewExportEmbeddings(repo, nil)
	out, err := uc.Execute(context.Background(), ExportEmbeddingsInput{
		MeetingIDs: []domain.MeetingID{"m-1"},
		Strategy:   "token_limit",
		MaxTokens:  5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ChunkCount < 1 {
		t.Errorf("expected at least 1 chunk, got %d", out.ChunkCount)
	}
}

func TestExportEmbeddings_MeetingNotFound(t *testing.T) {
	repo := &mockMeetingRepo{
		meetings:    map[domain.MeetingID]*domain.Meeting{},
		transcripts: map[domain.MeetingID]*domain.Transcript{},
	}

	uc := NewExportEmbeddings(repo, nil)
	_, err := uc.Execute(context.Background(), ExportEmbeddingsInput{
		MeetingIDs: []domain.MeetingID{"nonexistent"},
	})
	if err == nil {
		t.Error("expected error for nonexistent meeting")
	}
}

func TestExportEmbeddings_DefaultStrategy(t *testing.T) {
	now := time.Now().UTC()
	mtg, _ := domain.New("m-1", "Sprint", now, domain.SourceZoom, nil)
	mtg.ClearDomainEvents()

	utts := []domain.Utterance{
		domain.NewUtterance("Alice", "Hello", now, 0.9),
	}
	transcript := domain.NewTranscript("m-1", utts)

	repo := &mockMeetingRepo{
		meetings:    map[domain.MeetingID]*domain.Meeting{"m-1": mtg},
		transcripts: map[domain.MeetingID]*domain.Transcript{"m-1": &transcript},
	}

	uc := NewExportEmbeddings(repo, nil)
	// Empty strategy should default to speaker_turn
	out, err := uc.Execute(context.Background(), ExportEmbeddingsInput{
		MeetingIDs: []domain.MeetingID{"m-1"},
		Strategy:   "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ChunkCount != 1 {
		t.Errorf("ChunkCount = %d, want 1", out.ChunkCount)
	}
}
