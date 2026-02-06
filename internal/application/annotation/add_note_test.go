package annotation_test

import (
	"context"
	"testing"
	"time"

	app "github.com/felixgeelhaar/granola-mcp/internal/application/annotation"
	annotatn "github.com/felixgeelhaar/granola-mcp/internal/domain/annotation"
	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

func TestAddNote_Success(t *testing.T) {
	noteRepo := newMockNoteRepository()
	meetingRepo := newMockMeetingRepository()
	dispatcher := &mockDispatcher{}

	mtg, _ := domain.New("m-1", "Sprint Planning", time.Now(), domain.SourceZoom, nil)
	mtg.ClearDomainEvents()
	meetingRepo.addMeeting(mtg)

	uc := app.NewAddNote(noteRepo, meetingRepo, dispatcher)
	out, err := uc.Execute(context.Background(), app.AddNoteInput{
		MeetingID: "m-1",
		Author:    "claude",
		Content:   "Great discussion on architecture",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Note.MeetingID() != "m-1" {
		t.Errorf("got meeting id %q", out.Note.MeetingID())
	}
	if out.Note.Author() != "claude" {
		t.Errorf("got author %q", out.Note.Author())
	}
	if out.Note.Content() != "Great discussion on architecture" {
		t.Errorf("got content %q", out.Note.Content())
	}

	// Verify event dispatched
	if len(dispatcher.events) != 1 {
		t.Fatalf("got %d events, want 1", len(dispatcher.events))
	}
	if dispatcher.events[0].EventName() != "note.added" {
		t.Errorf("got event %q", dispatcher.events[0].EventName())
	}
}

func TestAddNote_MeetingNotFound(t *testing.T) {
	noteRepo := newMockNoteRepository()
	meetingRepo := newMockMeetingRepository()

	uc := app.NewAddNote(noteRepo, meetingRepo, nil)
	_, err := uc.Execute(context.Background(), app.AddNoteInput{
		MeetingID: "nonexistent",
		Author:    "claude",
		Content:   "Note",
	})
	if err != domain.ErrMeetingNotFound {
		t.Errorf("got error %v, want %v", err, domain.ErrMeetingNotFound)
	}
}

func TestAddNote_EmptyMeetingID(t *testing.T) {
	uc := app.NewAddNote(newMockNoteRepository(), newMockMeetingRepository(), nil)
	_, err := uc.Execute(context.Background(), app.AddNoteInput{
		MeetingID: "",
		Author:    "claude",
		Content:   "Note",
	})
	if err != annotatn.ErrInvalidMeetingID {
		t.Errorf("got error %v, want %v", err, annotatn.ErrInvalidMeetingID)
	}
}

func TestAddNote_EmptyContent(t *testing.T) {
	noteRepo := newMockNoteRepository()
	meetingRepo := newMockMeetingRepository()
	mtg, _ := domain.New("m-1", "Meeting", time.Now(), domain.SourceZoom, nil)
	mtg.ClearDomainEvents()
	meetingRepo.addMeeting(mtg)

	uc := app.NewAddNote(noteRepo, meetingRepo, nil)
	_, err := uc.Execute(context.Background(), app.AddNoteInput{
		MeetingID: "m-1",
		Author:    "claude",
		Content:   "",
	})
	if err != annotatn.ErrInvalidNoteContent {
		t.Errorf("got error %v, want %v", err, annotatn.ErrInvalidNoteContent)
	}
}

func TestAddNote_EmptyAuthor(t *testing.T) {
	noteRepo := newMockNoteRepository()
	meetingRepo := newMockMeetingRepository()
	mtg, _ := domain.New("m-1", "Meeting", time.Now(), domain.SourceZoom, nil)
	mtg.ClearDomainEvents()
	meetingRepo.addMeeting(mtg)

	uc := app.NewAddNote(noteRepo, meetingRepo, nil)
	_, err := uc.Execute(context.Background(), app.AddNoteInput{
		MeetingID: "m-1",
		Author:    "",
		Content:   "content",
	})
	if err != annotatn.ErrInvalidAuthor {
		t.Errorf("got error %v, want %v", err, annotatn.ErrInvalidAuthor)
	}
}
