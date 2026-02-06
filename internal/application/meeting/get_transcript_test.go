package meeting_test

import (
	"context"
	"testing"
	"time"

	app "github.com/felixgeelhaar/granola-mcp/internal/application/meeting"
	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

func TestGetTranscript_Found(t *testing.T) {
	repo := newMockRepository()
	transcript := domain.NewTranscript("m-1", []domain.Utterance{
		domain.NewUtterance("Alice", "Hello", time.Now().UTC(), 0.95),
	})
	repo.addTranscript("m-1", &transcript)

	uc := app.NewGetTranscript(repo)
	out, err := uc.Execute(context.Background(), app.GetTranscriptInput{MeetingID: "m-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Transcript.Utterances()) != 1 {
		t.Errorf("got %d utterances, want 1", len(out.Transcript.Utterances()))
	}
}

func TestGetTranscript_NotReady(t *testing.T) {
	repo := newMockRepository()
	uc := app.NewGetTranscript(repo)

	_, err := uc.Execute(context.Background(), app.GetTranscriptInput{MeetingID: "m-1"})
	if err != domain.ErrTranscriptNotReady {
		t.Errorf("got error %v, want %v", err, domain.ErrTranscriptNotReady)
	}
}

func TestGetTranscript_EmptyMeetingID(t *testing.T) {
	repo := newMockRepository()
	uc := app.NewGetTranscript(repo)

	_, err := uc.Execute(context.Background(), app.GetTranscriptInput{MeetingID: ""})
	if err != domain.ErrInvalidMeetingID {
		t.Errorf("got error %v, want %v", err, domain.ErrInvalidMeetingID)
	}
}
