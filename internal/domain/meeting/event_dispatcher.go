package meeting

import "context"

// EventDispatcher is the port for dispatching domain events to subscribers.
// Defined in the domain layer, implemented in infrastructure.
type EventDispatcher interface {
	Dispatch(ctx context.Context, events []DomainEvent) error
}

// EventNotifier is the port for sending resource change notifications.
// Defined in the domain layer, implemented in infrastructure.
type EventNotifier interface {
	NotifyResourceUpdated(uri string) error
	NotifyResourceListChanged() error
}
