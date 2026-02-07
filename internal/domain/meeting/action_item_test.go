package meeting_test

import (
	"testing"
	"time"

	"github.com/felixgeelhaar/acai/internal/domain/meeting"
)

func TestNewActionItem_Valid(t *testing.T) {
	due := time.Now().Add(24 * time.Hour).UTC()
	item, err := meeting.NewActionItem("ai-1", "m-1", "Alice", "Write report", &due)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if item.ID() != meeting.ActionItemID("ai-1") {
		t.Errorf("got id %q", item.ID())
	}
	if item.MeetingID() != meeting.MeetingID("m-1") {
		t.Errorf("got meeting id %q", item.MeetingID())
	}
	if item.Owner() != "Alice" {
		t.Errorf("got owner %q", item.Owner())
	}
	if item.Text() != "Write report" {
		t.Errorf("got text %q", item.Text())
	}
	if item.DueDate() == nil || !item.DueDate().Equal(due) {
		t.Error("due date mismatch")
	}
	if item.IsCompleted() {
		t.Error("new action item should not be completed")
	}
}

func TestNewActionItem_NilDueDate(t *testing.T) {
	item, err := meeting.NewActionItem("ai-1", "m-1", "Alice", "Write report", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.DueDate() != nil {
		t.Error("due date should be nil")
	}
}

func TestNewActionItem_RejectsEmptyID(t *testing.T) {
	_, err := meeting.NewActionItem("", "m-1", "Alice", "Text", nil)
	if err != meeting.ErrInvalidActionItemID {
		t.Errorf("got error %v, want %v", err, meeting.ErrInvalidActionItemID)
	}
}

func TestNewActionItem_RejectsEmptyText(t *testing.T) {
	_, err := meeting.NewActionItem("ai-1", "m-1", "Alice", "", nil)
	if err != meeting.ErrInvalidActionItemText {
		t.Errorf("got error %v, want %v", err, meeting.ErrInvalidActionItemText)
	}
}

func TestActionItem_Complete(t *testing.T) {
	item, _ := meeting.NewActionItem("ai-1", "m-1", "Alice", "Write report", nil)
	item.Complete()

	if !item.IsCompleted() {
		t.Error("action item should be completed after Complete()")
	}
}

func TestActionItem_CompleteIdempotent(t *testing.T) {
	item, _ := meeting.NewActionItem("ai-1", "m-1", "Alice", "Write report", nil)
	item.Complete()
	item.Complete() // second call should not panic

	if !item.IsCompleted() {
		t.Error("action item should still be completed")
	}
}

func TestActionItem_Uncomplete(t *testing.T) {
	item, _ := meeting.NewActionItem("ai-1", "m-1", "Alice", "Write report", nil)
	item.Complete()
	item.Uncomplete()

	if item.IsCompleted() {
		t.Error("action item should not be completed after Uncomplete()")
	}
}

func TestActionItem_UncompleteIdempotent(t *testing.T) {
	item, _ := meeting.NewActionItem("ai-1", "m-1", "Alice", "Write report", nil)
	item.Uncomplete() // already not completed

	if item.IsCompleted() {
		t.Error("action item should not be completed")
	}
}

func TestActionItem_UpdateText(t *testing.T) {
	item, _ := meeting.NewActionItem("ai-1", "m-1", "Alice", "Write report", nil)
	err := item.UpdateText("Updated report")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Text() != "Updated report" {
		t.Errorf("got text %q, want %q", item.Text(), "Updated report")
	}
}

func TestActionItem_UpdateText_RejectsEmpty(t *testing.T) {
	item, _ := meeting.NewActionItem("ai-1", "m-1", "Alice", "Write report", nil)
	err := item.UpdateText("")
	if err != meeting.ErrInvalidActionItemText {
		t.Errorf("got error %v, want %v", err, meeting.ErrInvalidActionItemText)
	}
	if item.Text() != "Write report" {
		t.Error("text should be unchanged after rejected update")
	}
}
