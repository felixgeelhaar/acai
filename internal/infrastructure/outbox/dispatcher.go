package outbox

import (
	"context"
	"fmt"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

// writeEventTypes are the event types that should be persisted to the outbox.
var writeEventTypes = map[string]bool{
	"note.added":            true,
	"note.deleted":          true,
	"action_item.completed": true,
	"action_item.updated":   true,
}

// Dispatcher decorates a domain.EventDispatcher, persisting write events
// to an outbox table alongside dispatching them to MCP sessions.
type Dispatcher struct {
	inner domain.EventDispatcher
	store Store
}

// NewDispatcher creates a new outbox dispatcher decorator.
func NewDispatcher(inner domain.EventDispatcher, store Store) *Dispatcher {
	return &Dispatcher{inner: inner, store: store}
}

// Dispatch forwards all events to the inner dispatcher, and additionally
// persists write-related events to the outbox for future upstream sync.
func (d *Dispatcher) Dispatch(ctx context.Context, events []domain.DomainEvent) error {
	// Always dispatch to inner (MCP sessions) first
	if err := d.inner.Dispatch(ctx, events); err != nil {
		return err
	}

	// Persist write events to outbox
	for _, event := range events {
		if writeEventTypes[event.EventName()] {
			entry := Entry{
				ID:        fmt.Sprintf("%s-%d", event.EventName(), time.Now().UnixNano()),
				EventType: event.EventName(),
				Payload:   MarshalEventPayload(event),
				CreatedAt: event.OccurredAt(),
			}
			if err := d.store.Append(entry); err != nil {
				return fmt.Errorf("outbox append %s: %w", event.EventName(), err)
			}
		}
	}
	return nil
}

var _ domain.EventDispatcher = (*Dispatcher)(nil)
