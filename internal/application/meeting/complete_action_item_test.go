package meeting_test

import (
	"context"
	"testing"

	app "github.com/felixgeelhaar/granola-mcp/internal/application/meeting"
	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

func TestCompleteActionItem_Success(t *testing.T) {
	repo := newMockRepository()
	writeRepo := newMockWriteRepository()
	dispatcher := &mockDispatcher{}

	item, _ := domain.NewActionItem("ai-1", "m-1", "Alice", "Write report", nil)
	repo.addActionItems("m-1", []*domain.ActionItem{item})

	uc := app.NewCompleteActionItem(repo, writeRepo, dispatcher)
	out, err := uc.Execute(context.Background(), app.CompleteActionItemInput{
		MeetingID:    "m-1",
		ActionItemID: "ai-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !out.Item.IsCompleted() {
		t.Error("action item should be completed")
	}

	// Verify persisted
	saved := writeRepo.items["ai-1"]
	if saved == nil || !saved.IsCompleted() {
		t.Error("should be saved as completed in write repo")
	}

	// Verify event
	if len(dispatcher.events) != 1 {
		t.Fatalf("got %d events, want 1", len(dispatcher.events))
	}
	if dispatcher.events[0].EventName() != "action_item.completed" {
		t.Errorf("got event %q", dispatcher.events[0].EventName())
	}
}

func TestCompleteActionItem_NotFound(t *testing.T) {
	repo := newMockRepository()
	writeRepo := newMockWriteRepository()

	repo.addActionItems("m-1", []*domain.ActionItem{})

	uc := app.NewCompleteActionItem(repo, writeRepo, nil)
	_, err := uc.Execute(context.Background(), app.CompleteActionItemInput{
		MeetingID:    "m-1",
		ActionItemID: "nonexistent",
	})
	if err != domain.ErrMeetingNotFound {
		t.Errorf("got error %v, want %v", err, domain.ErrMeetingNotFound)
	}
}

func TestCompleteActionItem_EmptyMeetingID(t *testing.T) {
	uc := app.NewCompleteActionItem(newMockRepository(), newMockWriteRepository(), nil)
	_, err := uc.Execute(context.Background(), app.CompleteActionItemInput{
		MeetingID:    "",
		ActionItemID: "ai-1",
	})
	if err != domain.ErrInvalidMeetingID {
		t.Errorf("got error %v, want %v", err, domain.ErrInvalidMeetingID)
	}
}

func TestCompleteActionItem_EmptyActionItemID(t *testing.T) {
	uc := app.NewCompleteActionItem(newMockRepository(), newMockWriteRepository(), nil)
	_, err := uc.Execute(context.Background(), app.CompleteActionItemInput{
		MeetingID:    "m-1",
		ActionItemID: "",
	})
	if err != domain.ErrInvalidActionItemID {
		t.Errorf("got error %v, want %v", err, domain.ErrInvalidActionItemID)
	}
}
