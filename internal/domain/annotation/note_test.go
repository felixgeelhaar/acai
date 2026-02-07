package annotation_test

import (
	"testing"

	"github.com/felixgeelhaar/acai/internal/domain/annotation"
)

func TestNewAgentNote_Valid(t *testing.T) {
	note, err := annotation.NewAgentNote("n-1", "m-1", "claude", "This is an observation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if note.ID() != annotation.NoteID("n-1") {
		t.Errorf("got id %q", note.ID())
	}
	if note.MeetingID() != "m-1" {
		t.Errorf("got meeting id %q", note.MeetingID())
	}
	if note.Author() != "claude" {
		t.Errorf("got author %q", note.Author())
	}
	if note.Content() != "This is an observation" {
		t.Errorf("got content %q", note.Content())
	}
	if note.CreatedAt().IsZero() {
		t.Error("created_at should not be zero")
	}
}

func TestNewAgentNote_RejectsEmptyID(t *testing.T) {
	_, err := annotation.NewAgentNote("", "m-1", "claude", "content")
	if err != annotation.ErrInvalidNoteID {
		t.Errorf("got error %v, want %v", err, annotation.ErrInvalidNoteID)
	}
}

func TestNewAgentNote_RejectsEmptyMeetingID(t *testing.T) {
	_, err := annotation.NewAgentNote("n-1", "", "claude", "content")
	if err != annotation.ErrInvalidMeetingID {
		t.Errorf("got error %v, want %v", err, annotation.ErrInvalidMeetingID)
	}
}

func TestNewAgentNote_RejectsEmptyAuthor(t *testing.T) {
	_, err := annotation.NewAgentNote("n-1", "m-1", "", "content")
	if err != annotation.ErrInvalidAuthor {
		t.Errorf("got error %v, want %v", err, annotation.ErrInvalidAuthor)
	}
}

func TestNewAgentNote_RejectsEmptyContent(t *testing.T) {
	_, err := annotation.NewAgentNote("n-1", "m-1", "claude", "")
	if err != annotation.ErrInvalidNoteContent {
		t.Errorf("got error %v, want %v", err, annotation.ErrInvalidNoteContent)
	}
}

func TestReconstructAgentNote(t *testing.T) {
	note, _ := annotation.NewAgentNote("n-1", "m-1", "claude", "content")
	reconstructed := annotation.ReconstructAgentNote(
		note.ID(), note.MeetingID(), note.Author(), note.Content(), note.CreatedAt(),
	)

	if reconstructed.ID() != note.ID() {
		t.Errorf("id mismatch: got %q, want %q", reconstructed.ID(), note.ID())
	}
	if reconstructed.CreatedAt() != note.CreatedAt() {
		t.Error("created_at mismatch")
	}
}

func TestNoteAdded_Event(t *testing.T) {
	event := annotation.NewNoteAddedEvent("n-1", "m-1", "claude")

	if event.EventName() != "note.added" {
		t.Errorf("got event name %q", event.EventName())
	}
	if event.NoteID() != "n-1" {
		t.Errorf("got note id %q", event.NoteID())
	}
	if event.MeetingID() != "m-1" {
		t.Errorf("got meeting id %q", event.MeetingID())
	}
	if event.Author() != "claude" {
		t.Errorf("got author %q", event.Author())
	}
	if event.OccurredAt().IsZero() {
		t.Error("occurred_at should not be zero")
	}
}

func TestNoteDeleted_Event(t *testing.T) {
	event := annotation.NewNoteDeletedEvent("n-1", "m-1")

	if event.EventName() != "note.deleted" {
		t.Errorf("got event name %q", event.EventName())
	}
	if event.NoteID() != "n-1" {
		t.Errorf("got note id %q", event.NoteID())
	}
	if event.MeetingID() != "m-1" {
		t.Errorf("got meeting id %q", event.MeetingID())
	}
	if event.OccurredAt().IsZero() {
		t.Error("occurred_at should not be zero")
	}
}
