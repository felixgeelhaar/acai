package meeting_test

import (
	"context"
	"testing"
	"time"

	app "github.com/felixgeelhaar/granola-mcp/internal/application/meeting"
	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

func TestListMeetings_ReturnsAll(t *testing.T) {
	repo := newMockRepository()
	m1 := mustNewMeeting(t, "m-1", "Meeting 1")
	m2 := mustNewMeeting(t, "m-2", "Meeting 2")
	repo.addMeeting(m1)
	repo.addMeeting(m2)

	uc := app.NewListMeetings(repo)
	out, err := uc.Execute(context.Background(), app.ListMeetingsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Total != 2 {
		t.Errorf("got total %d, want 2", out.Total)
	}
	if !repo.listCalled {
		t.Error("expected List to be called on repository")
	}
}

func TestListMeetings_PassesFilterToRepository(t *testing.T) {
	repo := newMockRepository()
	uc := app.NewListMeetings(repo)

	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	source := "zoom"
	_, _ = uc.Execute(context.Background(), app.ListMeetingsInput{
		Since:  &since,
		Source: &source,
		Limit:  10,
		Offset: 5,
	})

	if repo.listFilter == nil {
		t.Fatal("filter not passed to repository")
	}
	if repo.listFilter.Limit != 10 {
		t.Errorf("got limit %d, want 10", repo.listFilter.Limit)
	}
	if repo.listFilter.Offset != 5 {
		t.Errorf("got offset %d, want 5", repo.listFilter.Offset)
	}
	if *repo.listFilter.Source != domain.SourceZoom {
		t.Errorf("got source %q, want %q", *repo.listFilter.Source, domain.SourceZoom)
	}
}

func TestListMeetings_EmptyResult(t *testing.T) {
	repo := newMockRepository()
	uc := app.NewListMeetings(repo)

	out, err := uc.Execute(context.Background(), app.ListMeetingsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Total != 0 {
		t.Errorf("got total %d, want 0", out.Total)
	}
	if len(out.Meetings) != 0 {
		t.Errorf("got %d meetings, want 0", len(out.Meetings))
	}
}

func mustNewMeeting(t *testing.T, id domain.MeetingID, title string) *domain.Meeting {
	t.Helper()
	m, err := domain.New(id, title, time.Now().UTC(), domain.SourceZoom, nil)
	if err != nil {
		t.Fatalf("failed to create meeting: %v", err)
	}
	m.ClearDomainEvents()
	return m
}
