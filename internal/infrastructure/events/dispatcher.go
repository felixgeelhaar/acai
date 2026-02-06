// Package events implements event dispatching infrastructure.
// It maps domain events to MCP resource URIs and notifies subscribers.
package events

import (
	"context"
	"fmt"
	"log"

	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

// Dispatcher maps domain events to resource URIs and calls the notifier.
// Implements domain.EventDispatcher.
type Dispatcher struct {
	notifier domain.EventNotifier
}

// NewDispatcher creates a new event dispatcher.
// If notifier is nil, dispatch is a no-op.
func NewDispatcher(notifier domain.EventNotifier) *Dispatcher {
	return &Dispatcher{notifier: notifier}
}

// Dispatch maps each domain event to resource URI(s) and sends notifications.
// Unknown event types are logged but do not cause failure.
// Notifier errors are logged but dispatching continues for remaining events.
func (d *Dispatcher) Dispatch(_ context.Context, events []domain.DomainEvent) error {
	if d.notifier == nil {
		return nil
	}

	for _, event := range events {
		d.dispatchEvent(event)
	}
	return nil
}

func (d *Dispatcher) dispatchEvent(event domain.DomainEvent) {
	switch e := event.(type) {
	case domain.MeetingCreated:
		uri := fmt.Sprintf("meeting://%s", e.MeetingID())
		if err := d.notifier.NotifyResourceUpdated(uri); err != nil {
			log.Printf("event dispatch: notify resource updated %q: %v", uri, err)
		}
		if err := d.notifier.NotifyResourceListChanged(); err != nil {
			log.Printf("event dispatch: notify resource list changed: %v", err)
		}

	case domain.TranscriptUpdated:
		uri := fmt.Sprintf("transcript://%s", e.MeetingID())
		if err := d.notifier.NotifyResourceUpdated(uri); err != nil {
			log.Printf("event dispatch: notify resource updated %q: %v", uri, err)
		}

	case domain.SummaryUpdated:
		uri := fmt.Sprintf("meeting://%s", e.MeetingID())
		if err := d.notifier.NotifyResourceUpdated(uri); err != nil {
			log.Printf("event dispatch: notify resource updated %q: %v", uri, err)
		}

	case domain.ActionItemCompleted:
		uri := fmt.Sprintf("meeting://%s", e.MeetingID())
		if err := d.notifier.NotifyResourceUpdated(uri); err != nil {
			log.Printf("event dispatch: notify resource updated %q: %v", uri, err)
		}

	case domain.ActionItemUpdated:
		uri := fmt.Sprintf("meeting://%s", e.MeetingID())
		if err := d.notifier.NotifyResourceUpdated(uri); err != nil {
			log.Printf("event dispatch: notify resource updated %q: %v", uri, err)
		}

	default:
		// Annotation events and other unknown types — log but don't fail.
		// Annotation events (note.added, note.deleted) trigger note resource updates
		// via the note://{meeting_id} URI pattern, handled at the interface level.
		log.Printf("event dispatch: unknown event type %q", event.EventName())
	}
}
