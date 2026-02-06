package meeting

import (
	"context"
	"time"
)

// ListFilter defines criteria for querying meetings.
// This is a domain-level specification pattern.
type ListFilter struct {
	Since       *time.Time
	Until       *time.Time
	Source      *Source
	Participant *string
	Query       *string
	Limit       int
	Offset      int
}

// Repository is the port (interface) for meeting persistence (read-only).
// Defined in the domain layer, implemented in infrastructure.
// This is the DDD repository pattern — it provides collection-like access
// to aggregates while hiding persistence details.
type Repository interface {
	FindByID(ctx context.Context, id MeetingID) (*Meeting, error)
	List(ctx context.Context, filter ListFilter) ([]*Meeting, error)
	GetTranscript(ctx context.Context, id MeetingID) (*Transcript, error)
	SearchTranscripts(ctx context.Context, query string, filter ListFilter) ([]*Meeting, error)
	GetActionItems(ctx context.Context, id MeetingID) ([]*ActionItem, error)
	Sync(ctx context.Context, since *time.Time) ([]DomainEvent, error)
}

// WriteRepository is a separate port for local write operations (ISP).
// Write operations go to local SQLite — no resilience or cache decorators needed.
// This follows CQRS: reads go through the decorator chain, writes go directly to local store.
type WriteRepository interface {
	SaveActionItemState(ctx context.Context, item *ActionItem) error
	GetLocalActionItemState(ctx context.Context, id ActionItemID) (*ActionItem, error)
}
