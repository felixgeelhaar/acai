package annotation_test

import (
	"context"
	"testing"

	app "github.com/felixgeelhaar/acai/internal/application/annotation"
	annotatn "github.com/felixgeelhaar/acai/internal/domain/annotation"
)

func TestDeleteNote_Success(t *testing.T) {
	noteRepo := newMockNoteRepository()
	dispatcher := &mockDispatcher{}

	note, _ := annotatn.NewAgentNote("n-1", "m-1", "claude", "observation")
	noteRepo.notes[note.ID()] = note

	uc := app.NewDeleteNote(noteRepo, dispatcher)
	_, err := uc.Execute(context.Background(), app.DeleteNoteInput{NoteID: "n-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify note deleted
	if len(noteRepo.notes) != 0 {
		t.Errorf("note repo should be empty, got %d", len(noteRepo.notes))
	}

	// Verify event dispatched
	if len(dispatcher.events) != 1 {
		t.Fatalf("got %d events, want 1", len(dispatcher.events))
	}
	if dispatcher.events[0].EventName() != "note.deleted" {
		t.Errorf("got event %q", dispatcher.events[0].EventName())
	}
}

func TestDeleteNote_NotFound(t *testing.T) {
	noteRepo := newMockNoteRepository()
	uc := app.NewDeleteNote(noteRepo, nil)
	_, err := uc.Execute(context.Background(), app.DeleteNoteInput{NoteID: "nonexistent"})
	if err != annotatn.ErrNoteNotFound {
		t.Errorf("got error %v, want %v", err, annotatn.ErrNoteNotFound)
	}
}

func TestDeleteNote_EmptyNoteID(t *testing.T) {
	uc := app.NewDeleteNote(newMockNoteRepository(), nil)
	_, err := uc.Execute(context.Background(), app.DeleteNoteInput{NoteID: ""})
	if err != annotatn.ErrInvalidNoteID {
		t.Errorf("got error %v, want %v", err, annotatn.ErrInvalidNoteID)
	}
}
