package meeting_test

import (
	"context"
	"testing"

	app "github.com/felixgeelhaar/granola-mcp/internal/application/meeting"
	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

func TestGetMeeting_Found(t *testing.T) {
	repo := newMockRepository()
	m := mustNewMeeting(t, "m-1", "Sprint Planning")
	repo.addMeeting(m)

	uc := app.NewGetMeeting(repo)
	out, err := uc.Execute(context.Background(), app.GetMeetingInput{ID: "m-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Meeting.ID() != "m-1" {
		t.Errorf("got id %q", out.Meeting.ID())
	}
	if !repo.findByIDCalled {
		t.Error("expected FindByID to be called")
	}
}

func TestGetMeeting_NotFound(t *testing.T) {
	repo := newMockRepository()
	uc := app.NewGetMeeting(repo)

	_, err := uc.Execute(context.Background(), app.GetMeetingInput{ID: "nonexistent"})
	if err != domain.ErrMeetingNotFound {
		t.Errorf("got error %v, want %v", err, domain.ErrMeetingNotFound)
	}
}

func TestGetMeeting_EmptyID(t *testing.T) {
	repo := newMockRepository()
	uc := app.NewGetMeeting(repo)

	_, err := uc.Execute(context.Background(), app.GetMeetingInput{ID: ""})
	if err != domain.ErrInvalidMeetingID {
		t.Errorf("got error %v, want %v", err, domain.ErrInvalidMeetingID)
	}
}
