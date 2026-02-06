package meeting_test

import (
	"context"
	"testing"
	"time"

	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

// mockDispatcher verifies EventDispatcher interface is implementable.
type mockDispatcher struct {
	dispatched []domain.DomainEvent
}

func (m *mockDispatcher) Dispatch(_ context.Context, events []domain.DomainEvent) error {
	m.dispatched = append(m.dispatched, events...)
	return nil
}

// mockNotifier verifies EventNotifier interface is implementable.
type mockNotifier struct {
	updatedURIs    []string
	listChangedCnt int
}

func (m *mockNotifier) NotifyResourceUpdated(uri string) error {
	m.updatedURIs = append(m.updatedURIs, uri)
	return nil
}

func (m *mockNotifier) NotifyResourceListChanged() error {
	m.listChangedCnt++
	return nil
}

func TestEventDispatcher_InterfaceContract(t *testing.T) {
	var d domain.EventDispatcher = &mockDispatcher{}

	event := domain.NewMeetingCreatedEvent("m-1", "Test", time.Now().UTC())
	err := d.Dispatch(context.Background(), []domain.DomainEvent{event})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := d.(*mockDispatcher)
	if len(md.dispatched) != 1 {
		t.Errorf("got %d dispatched events, want 1", len(md.dispatched))
	}
}

func TestEventNotifier_InterfaceContract(t *testing.T) {
	var n domain.EventNotifier = &mockNotifier{}

	if err := n.NotifyResourceUpdated("meeting://m-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := n.NotifyResourceListChanged(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mn := n.(*mockNotifier)
	if len(mn.updatedURIs) != 1 || mn.updatedURIs[0] != "meeting://m-1" {
		t.Errorf("unexpected updatedURIs: %v", mn.updatedURIs)
	}
	if mn.listChangedCnt != 1 {
		t.Errorf("expected 1 list changed call, got %d", mn.listChangedCnt)
	}
}
