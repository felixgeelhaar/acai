package outbox_test

import (
	"context"
	"testing"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
	"github.com/felixgeelhaar/acai/internal/infrastructure/outbox"
)

// mockInnerDispatcher captures dispatched events.
type mockInnerDispatcher struct {
	dispatched []domain.DomainEvent
	err        error
}

func (m *mockInnerDispatcher) Dispatch(_ context.Context, events []domain.DomainEvent) error {
	if m.err != nil {
		return m.err
	}
	m.dispatched = append(m.dispatched, events...)
	return nil
}

// mockOutboxStore captures appended entries.
type mockOutboxStore struct {
	entries []outbox.Entry
}

func (m *mockOutboxStore) Append(entry outbox.Entry) error {
	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockOutboxStore) ListPending() ([]outbox.Entry, error) { return m.entries, nil }
func (m *mockOutboxStore) MarkSynced(_ string) error            { return nil }
func (m *mockOutboxStore) MarkFailed(_ string) error            { return nil }

func TestOutboxDispatcher_PersistsWriteEvents(t *testing.T) {
	inner := &mockInnerDispatcher{}
	store := &mockOutboxStore{}
	d := outbox.NewDispatcher(inner, store)

	events := []domain.DomainEvent{
		domain.NewActionItemCompletedEvent("m-1", "ai-1"),
	}

	if err := d.Dispatch(context.Background(), events); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	// Inner should receive the event
	if len(inner.dispatched) != 1 {
		t.Errorf("inner got %d events, want 1", len(inner.dispatched))
	}

	// Outbox should have persisted the write event
	if len(store.entries) != 1 {
		t.Fatalf("outbox got %d entries, want 1", len(store.entries))
	}
	if store.entries[0].EventType != "action_item.completed" {
		t.Errorf("got event type %q", store.entries[0].EventType)
	}
}

func TestOutboxDispatcher_SkipsNonWriteEvents(t *testing.T) {
	inner := &mockInnerDispatcher{}
	store := &mockOutboxStore{}
	d := outbox.NewDispatcher(inner, store)

	events := []domain.DomainEvent{
		domain.NewMeetingCreatedEvent("m-1", "Sprint Planning", time.Now()),
	}

	if err := d.Dispatch(context.Background(), events); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	// Inner should receive the event
	if len(inner.dispatched) != 1 {
		t.Errorf("inner got %d events, want 1", len(inner.dispatched))
	}

	// Outbox should NOT persist read events
	if len(store.entries) != 0 {
		t.Errorf("outbox got %d entries, want 0 (meeting.created is not a write event)", len(store.entries))
	}
}

func TestOutboxDispatcher_MixedEvents(t *testing.T) {
	inner := &mockInnerDispatcher{}
	store := &mockOutboxStore{}
	d := outbox.NewDispatcher(inner, store)

	events := []domain.DomainEvent{
		domain.NewMeetingCreatedEvent("m-1", "Meeting", time.Now()),
		domain.NewActionItemCompletedEvent("m-1", "ai-1"),
		domain.NewActionItemUpdatedEvent("m-1", "ai-2", "Updated text"),
	}

	if err := d.Dispatch(context.Background(), events); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	if len(inner.dispatched) != 3 {
		t.Errorf("inner got %d events, want 3", len(inner.dispatched))
	}

	// Only 2 write events should be persisted
	if len(store.entries) != 2 {
		t.Fatalf("outbox got %d entries, want 2", len(store.entries))
	}
}

func TestOutboxDispatcher_InnerError_PropagatesWithoutOutbox(t *testing.T) {
	inner := &mockInnerDispatcher{err: context.DeadlineExceeded}
	store := &mockOutboxStore{}
	d := outbox.NewDispatcher(inner, store)

	events := []domain.DomainEvent{
		domain.NewActionItemCompletedEvent("m-1", "ai-1"),
	}

	err := d.Dispatch(context.Background(), events)
	if err == nil {
		t.Fatal("expected error from inner dispatcher")
	}

	// Outbox should not be called when inner fails
	if len(store.entries) != 0 {
		t.Errorf("outbox got %d entries, want 0 (inner failed)", len(store.entries))
	}
}
