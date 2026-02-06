package meeting_test

import (
	"context"
	"errors"
	"testing"
	"time"

	app "github.com/felixgeelhaar/granola-mcp/internal/application/meeting"
	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

func TestSyncMeetings_DelegatesToRepository(t *testing.T) {
	repo := newMockRepository()
	repo.syncEvents = []domain.DomainEvent{
		domain.NewMeetingCreatedEvent("m-1", "New Meeting", time.Now().UTC()),
	}

	uc := app.NewSyncMeetings(repo)
	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	out, err := uc.Execute(context.Background(), app.SyncMeetingsInput{Since: &since})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.syncCalled {
		t.Error("expected Sync to be called")
	}
	if len(out.Events) != 1 {
		t.Errorf("got %d events, want 1", len(out.Events))
	}
}

func TestSyncMeetings_PropagatesErrors(t *testing.T) {
	repo := newMockRepository()
	repo.syncErr = errors.New("network failure")

	uc := app.NewSyncMeetings(repo)
	_, err := uc.Execute(context.Background(), app.SyncMeetingsInput{})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "network failure" {
		t.Errorf("got error %q", err.Error())
	}
}
