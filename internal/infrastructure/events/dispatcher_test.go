package events_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
	"github.com/felixgeelhaar/granola-mcp/internal/infrastructure/events"
)

type mockNotifier struct {
	updatedURIs    []string
	listChangedCnt int
	err            error
}

func (m *mockNotifier) NotifyResourceUpdated(uri string) error {
	m.updatedURIs = append(m.updatedURIs, uri)
	return m.err
}

func (m *mockNotifier) NotifyResourceListChanged() error {
	m.listChangedCnt++
	return m.err
}

func TestDispatcher_MeetingCreated_NotifiesResourceAndList(t *testing.T) {
	n := &mockNotifier{}
	d := events.NewDispatcher(n)

	ev := domain.NewMeetingCreatedEvent("m-1", "Sprint", time.Now().UTC())
	err := d.Dispatch(context.Background(), []domain.DomainEvent{ev})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(n.updatedURIs) != 1 || n.updatedURIs[0] != "meeting://m-1" {
		t.Errorf("expected [meeting://m-1], got %v", n.updatedURIs)
	}
	if n.listChangedCnt != 1 {
		t.Errorf("expected 1 list changed, got %d", n.listChangedCnt)
	}
}

func TestDispatcher_TranscriptUpdated_NotifiesTranscriptResource(t *testing.T) {
	n := &mockNotifier{}
	d := events.NewDispatcher(n)

	ev := domain.NewTranscriptUpdatedEvent("m-1", 42)
	err := d.Dispatch(context.Background(), []domain.DomainEvent{ev})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(n.updatedURIs) != 1 || n.updatedURIs[0] != "transcript://m-1" {
		t.Errorf("expected [transcript://m-1], got %v", n.updatedURIs)
	}
	if n.listChangedCnt != 0 {
		t.Errorf("expected 0 list changed, got %d", n.listChangedCnt)
	}
}

func TestDispatcher_SummaryUpdated_NotifiesMeetingResource(t *testing.T) {
	n := &mockNotifier{}
	d := events.NewDispatcher(n)

	ev := domain.NewSummaryUpdatedEvent("m-1", domain.SummaryAuto)
	err := d.Dispatch(context.Background(), []domain.DomainEvent{ev})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(n.updatedURIs) != 1 || n.updatedURIs[0] != "meeting://m-1" {
		t.Errorf("expected [meeting://m-1], got %v", n.updatedURIs)
	}
	if n.listChangedCnt != 0 {
		t.Errorf("expected 0 list changed, got %d", n.listChangedCnt)
	}
}

func TestDispatcher_MultipleEvents_NotifiesAll(t *testing.T) {
	n := &mockNotifier{}
	d := events.NewDispatcher(n)

	evts := []domain.DomainEvent{
		domain.NewMeetingCreatedEvent("m-1", "Sprint", time.Now().UTC()),
		domain.NewTranscriptUpdatedEvent("m-1", 10),
		domain.NewSummaryUpdatedEvent("m-2", domain.SummaryEdited),
	}
	err := d.Dispatch(context.Background(), evts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// MeetingCreated → meeting://m-1, TranscriptUpdated → transcript://m-1, SummaryUpdated → meeting://m-2
	if len(n.updatedURIs) != 3 {
		t.Fatalf("expected 3 updated URIs, got %d: %v", len(n.updatedURIs), n.updatedURIs)
	}
	if n.updatedURIs[0] != "meeting://m-1" {
		t.Errorf("expected meeting://m-1, got %s", n.updatedURIs[0])
	}
	if n.updatedURIs[1] != "transcript://m-1" {
		t.Errorf("expected transcript://m-1, got %s", n.updatedURIs[1])
	}
	if n.updatedURIs[2] != "meeting://m-2" {
		t.Errorf("expected meeting://m-2, got %s", n.updatedURIs[2])
	}
	if n.listChangedCnt != 1 {
		t.Errorf("expected 1 list changed, got %d", n.listChangedCnt)
	}
}

func TestDispatcher_EmptyEvents_NoOp(t *testing.T) {
	n := &mockNotifier{}
	d := events.NewDispatcher(n)

	err := d.Dispatch(context.Background(), []domain.DomainEvent{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(n.updatedURIs) != 0 {
		t.Errorf("expected no notifications, got %v", n.updatedURIs)
	}
}

// unknownEvent is a DomainEvent type unknown to the dispatcher.
type unknownEvent struct {
	occurred time.Time
}

func (e unknownEvent) EventName() string     { return "unknown.event" }
func (e unknownEvent) OccurredAt() time.Time { return e.occurred }

func TestDispatcher_UnknownEvent_LogsAndContinues(t *testing.T) {
	n := &mockNotifier{}
	d := events.NewDispatcher(n)

	evts := []domain.DomainEvent{
		unknownEvent{occurred: time.Now().UTC()},
		domain.NewMeetingCreatedEvent("m-1", "Sprint", time.Now().UTC()),
	}
	err := d.Dispatch(context.Background(), evts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still process the MeetingCreated after the unknown event
	if len(n.updatedURIs) != 1 {
		t.Errorf("expected 1 notification after unknown event, got %d", len(n.updatedURIs))
	}
}

func TestDispatcher_NotifierError_ContinuesDispatching(t *testing.T) {
	n := &mockNotifier{err: errors.New("notify failed")}
	d := events.NewDispatcher(n)

	evts := []domain.DomainEvent{
		domain.NewMeetingCreatedEvent("m-1", "Sprint", time.Now().UTC()),
		domain.NewTranscriptUpdatedEvent("m-2", 5),
	}
	err := d.Dispatch(context.Background(), evts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have attempted both notifications despite errors
	if len(n.updatedURIs) != 2 {
		t.Errorf("expected 2 notification attempts, got %d", len(n.updatedURIs))
	}
}

func TestDispatcher_NilNotifier_NoOp(t *testing.T) {
	d := events.NewDispatcher(nil)

	ev := domain.NewMeetingCreatedEvent("m-1", "Sprint", time.Now().UTC())
	err := d.Dispatch(context.Background(), []domain.DomainEvent{ev})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
