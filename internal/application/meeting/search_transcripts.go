package meeting

import (
	"context"
	"errors"
	"time"

	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

var ErrEmptyQuery = errors.New("search query must not be empty")

type SearchTranscriptsInput struct {
	Query string
	Since *time.Time
	Until *time.Time
	Limit int
}

type SearchTranscriptsOutput struct {
	Meetings []*domain.Meeting
	Total    int
}

type SearchTranscripts struct {
	repo domain.Repository
}

func NewSearchTranscripts(repo domain.Repository) *SearchTranscripts {
	return &SearchTranscripts{repo: repo}
}

func (uc *SearchTranscripts) Execute(ctx context.Context, input SearchTranscriptsInput) (*SearchTranscriptsOutput, error) {
	if input.Query == "" {
		return nil, ErrEmptyQuery
	}

	filter := domain.ListFilter{
		Since: input.Since,
		Until: input.Until,
		Limit: input.Limit,
	}

	meetings, err := uc.repo.SearchTranscripts(ctx, input.Query, filter)
	if err != nil {
		return nil, err
	}

	return &SearchTranscriptsOutput{
		Meetings: meetings,
		Total:    len(meetings),
	}, nil
}
