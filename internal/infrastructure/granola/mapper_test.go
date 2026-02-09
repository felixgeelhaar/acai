package granola

import (
	"testing"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

func TestMapNoteDetailToDomain(t *testing.T) {
	now := time.Now().UTC()
	markdown := "# Sprint Planning\nWe planned the sprint."
	dto := NoteDetailResponse{
		ID:    "m-1",
		Title: "Sprint Planning",
		Owner: UserDTO{Name: "Alice", Email: "alice@example.com"},
		CreatedAt: now,
		Attendees: []UserDTO{
			{Name: "Alice", Email: "alice@example.com"}, // duplicate of owner
			{Name: "Bob", Email: "bob@example.com"},
		},
		SummaryText:     "We planned the sprint.",
		SummaryMarkdown: &markdown,
	}

	mtg, err := mapNoteDetailToDomain(dto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mtg.ID() != "m-1" {
		t.Errorf("got id %q", mtg.ID())
	}
	if mtg.Title() != "Sprint Planning" {
		t.Errorf("got title %q", mtg.Title())
	}
	// Public API doesn't expose source — always SourceOther
	if mtg.Source() != domain.SourceOther {
		t.Errorf("got source %q, want %q", mtg.Source(), domain.SourceOther)
	}
	// Owner + Bob (Alice deduped)
	if len(mtg.Participants()) != 2 {
		t.Errorf("got %d participants, want 2", len(mtg.Participants()))
	}
	if mtg.Summary() == nil {
		t.Fatal("summary should be attached")
	}
	// Should prefer markdown over text
	if mtg.Summary().Content() != markdown {
		t.Errorf("got summary %q, want markdown", mtg.Summary().Content())
	}

	// Reconstitution should not produce domain events
	if len(mtg.DomainEvents()) != 0 {
		t.Errorf("reconstituted meeting should have 0 events, got %d", len(mtg.DomainEvents()))
	}
}

func TestMapNoteDetailToDomain_TextSummaryFallback(t *testing.T) {
	now := time.Now().UTC()
	dto := NoteDetailResponse{
		ID:          "m-1",
		Title:       "Meeting",
		Owner:       UserDTO{Name: "Alice", Email: "alice@example.com"},
		CreatedAt:   now,
		SummaryText: "Plain text summary",
	}

	mtg, err := mapNoteDetailToDomain(dto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mtg.Summary() == nil {
		t.Fatal("summary should be attached")
	}
	if mtg.Summary().Content() != "Plain text summary" {
		t.Errorf("got summary %q", mtg.Summary().Content())
	}
}

func TestMapNoteDetailToDomain_NoSummary(t *testing.T) {
	now := time.Now().UTC()
	dto := NoteDetailResponse{
		ID:        "m-1",
		Title:     "Meeting",
		Owner:     UserDTO{Name: "Alice", Email: "alice@example.com"},
		CreatedAt: now,
	}

	mtg, err := mapNoteDetailToDomain(dto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mtg.Summary() != nil {
		t.Error("expected no summary")
	}
}

func TestMapNoteListItemToDomain(t *testing.T) {
	now := time.Now().UTC()
	dto := NoteListItem{
		ID:        "m-1",
		Title:     "Sprint Planning",
		Owner:     UserDTO{Name: "Alice", Email: "alice@example.com"},
		CreatedAt: now,
	}

	mtg, err := mapNoteListItemToDomain(dto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mtg.ID() != "m-1" {
		t.Errorf("got id %q", mtg.ID())
	}
	if mtg.Title() != "Sprint Planning" {
		t.Errorf("got title %q", mtg.Title())
	}
	if len(mtg.Participants()) != 1 {
		t.Errorf("got %d participants, want 1", len(mtg.Participants()))
	}
	if mtg.Participants()[0].Role() != domain.RoleHost {
		t.Errorf("owner should be host, got %q", mtg.Participants()[0].Role())
	}
}

func TestMapTranscriptFromDetail(t *testing.T) {
	now := time.Now().UTC()
	dto := NoteDetailResponse{
		ID: "m-1",
		Transcript: []TranscriptItemDTO{
			{Speaker: "Alice", Text: "Hello", Timestamp: now},
			{Speaker: "Bob", Text: "Hi", Timestamp: now.Add(time.Second)},
		},
	}

	transcript := mapTranscriptFromDetail("m-1", dto)
	if transcript == nil {
		t.Fatal("expected transcript")
	}
	if transcript.MeetingID() != "m-1" {
		t.Errorf("got meeting id %q", transcript.MeetingID())
	}
	if len(transcript.Utterances()) != 2 {
		t.Errorf("got %d utterances", len(transcript.Utterances()))
	}
	if transcript.Utterances()[0].Speaker() != "Alice" {
		t.Errorf("got speaker %q", transcript.Utterances()[0].Speaker())
	}
}

func TestMapTranscriptFromDetail_Empty(t *testing.T) {
	dto := NoteDetailResponse{
		ID:         "m-1",
		Transcript: nil,
	}

	transcript := mapTranscriptFromDetail("m-1", dto)
	if transcript != nil {
		t.Error("expected nil transcript for empty transcript list")
	}
}

func TestMapUserToDomain(t *testing.T) {
	dto := UserDTO{Name: "Alice", Email: "alice@example.com"}
	p := mapUserToDomain(dto)

	if p.Name() != "Alice" {
		t.Errorf("got name %q", p.Name())
	}
	if p.Email() != "alice@example.com" {
		t.Errorf("got email %q", p.Email())
	}
	if p.Role() != domain.RoleAttendee {
		t.Errorf("got role %q, want attendee", p.Role())
	}
}
