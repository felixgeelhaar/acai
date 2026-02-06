package meeting_test

import (
	"context"
	"testing"

	app "github.com/felixgeelhaar/granola-mcp/internal/application/meeting"
	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

func TestGetActionItems_Found(t *testing.T) {
	repo := newMockRepository()
	item, _ := domain.NewActionItem("ai-1", "m-1", "Alice", "Write report", nil)
	repo.addActionItems("m-1", []*domain.ActionItem{item})

	uc := app.NewGetActionItems(repo)
	out, err := uc.Execute(context.Background(), app.GetActionItemsInput{MeetingID: "m-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Items) != 1 {
		t.Errorf("got %d items, want 1", len(out.Items))
	}
	if out.Items[0].Owner() != "Alice" {
		t.Errorf("got owner %q", out.Items[0].Owner())
	}
}

func TestGetActionItems_Empty(t *testing.T) {
	repo := newMockRepository()
	uc := app.NewGetActionItems(repo)

	out, err := uc.Execute(context.Background(), app.GetActionItemsInput{MeetingID: "m-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Items) != 0 {
		t.Errorf("got %d items, want 0", len(out.Items))
	}
}

func TestGetActionItems_EmptyMeetingID(t *testing.T) {
	repo := newMockRepository()
	uc := app.NewGetActionItems(repo)

	_, err := uc.Execute(context.Background(), app.GetActionItemsInput{MeetingID: ""})
	if err != domain.ErrInvalidMeetingID {
		t.Errorf("got error %v, want %v", err, domain.ErrInvalidMeetingID)
	}
}
