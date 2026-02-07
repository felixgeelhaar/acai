package annotation_test

import (
	"context"
	"testing"

	app "github.com/felixgeelhaar/acai/internal/application/annotation"
	annotatn "github.com/felixgeelhaar/acai/internal/domain/annotation"
)

func TestListNotes_Found(t *testing.T) {
	noteRepo := newMockNoteRepository()
	note, _ := annotatn.NewAgentNote("n-1", "m-1", "claude", "observation")
	noteRepo.notes[note.ID()] = note

	uc := app.NewListNotes(noteRepo)
	out, err := uc.Execute(context.Background(), app.ListNotesInput{MeetingID: "m-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Notes) != 1 {
		t.Errorf("got %d notes, want 1", len(out.Notes))
	}
}

func TestListNotes_Empty(t *testing.T) {
	noteRepo := newMockNoteRepository()

	uc := app.NewListNotes(noteRepo)
	out, err := uc.Execute(context.Background(), app.ListNotesInput{MeetingID: "m-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Notes) != 0 {
		t.Errorf("got %d notes, want 0", len(out.Notes))
	}
}

func TestListNotes_EmptyMeetingID(t *testing.T) {
	uc := app.NewListNotes(newMockNoteRepository())
	_, err := uc.Execute(context.Background(), app.ListNotesInput{MeetingID: ""})
	if err != annotatn.ErrInvalidMeetingID {
		t.Errorf("got error %v, want %v", err, annotatn.ErrInvalidMeetingID)
	}
}
