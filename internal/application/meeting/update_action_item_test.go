package meeting_test

import (
	"context"
	"testing"

	app "github.com/felixgeelhaar/acai/internal/application/meeting"
	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

func TestUpdateActionItem_Success(t *testing.T) {
	repo := newMockRepository()
	writeRepo := newMockWriteRepository()
	dispatcher := &mockDispatcher{}

	item, _ := domain.NewActionItem("ai-1", "m-1", "Alice", "Original text", nil)
	repo.addActionItems("m-1", []*domain.ActionItem{item})

	uc := app.NewUpdateActionItem(repo, writeRepo, dispatcher)
	out, err := uc.Execute(context.Background(), app.UpdateActionItemInput{
		MeetingID:    "m-1",
		ActionItemID: "ai-1",
		Text:         "Updated text",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Item.Text() != "Updated text" {
		t.Errorf("got text %q", out.Item.Text())
	}

	// Verify persisted
	saved := writeRepo.items["ai-1"]
	if saved == nil || saved.Text() != "Updated text" {
		t.Error("should be saved with updated text in write repo")
	}

	// Verify event
	if len(dispatcher.events) != 1 {
		t.Fatalf("got %d events, want 1", len(dispatcher.events))
	}
	if dispatcher.events[0].EventName() != "action_item.updated" {
		t.Errorf("got event %q", dispatcher.events[0].EventName())
	}
}

func TestUpdateActionItem_NotFound(t *testing.T) {
	repo := newMockRepository()
	writeRepo := newMockWriteRepository()

	repo.addActionItems("m-1", []*domain.ActionItem{})

	uc := app.NewUpdateActionItem(repo, writeRepo, nil)
	_, err := uc.Execute(context.Background(), app.UpdateActionItemInput{
		MeetingID:    "m-1",
		ActionItemID: "nonexistent",
		Text:         "New text",
	})
	if err != domain.ErrMeetingNotFound {
		t.Errorf("got error %v, want %v", err, domain.ErrMeetingNotFound)
	}
}

func TestUpdateActionItem_EmptyText(t *testing.T) {
	uc := app.NewUpdateActionItem(newMockRepository(), newMockWriteRepository(), nil)
	_, err := uc.Execute(context.Background(), app.UpdateActionItemInput{
		MeetingID:    "m-1",
		ActionItemID: "ai-1",
		Text:         "",
	})
	if err != domain.ErrInvalidActionItemText {
		t.Errorf("got error %v, want %v", err, domain.ErrInvalidActionItemText)
	}
}

func TestUpdateActionItem_EmptyMeetingID(t *testing.T) {
	uc := app.NewUpdateActionItem(newMockRepository(), newMockWriteRepository(), nil)
	_, err := uc.Execute(context.Background(), app.UpdateActionItemInput{
		MeetingID:    "",
		ActionItemID: "ai-1",
		Text:         "text",
	})
	if err != domain.ErrInvalidMeetingID {
		t.Errorf("got error %v, want %v", err, domain.ErrInvalidMeetingID)
	}
}
