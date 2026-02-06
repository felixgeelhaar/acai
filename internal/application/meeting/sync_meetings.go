package meeting

import (
	"context"
	"time"

	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

type SyncMeetingsInput struct {
	Since *time.Time
}

type SyncMeetingsOutput struct {
	Events []domain.DomainEvent
}

type SyncMeetings struct {
	repo domain.Repository
}

func NewSyncMeetings(repo domain.Repository) *SyncMeetings {
	return &SyncMeetings{repo: repo}
}

func (uc *SyncMeetings) Execute(ctx context.Context, input SyncMeetingsInput) (*SyncMeetingsOutput, error) {
	events, err := uc.repo.Sync(ctx, input.Since)
	if err != nil {
		return nil, err
	}

	return &SyncMeetingsOutput{Events: events}, nil
}
