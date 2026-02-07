package meeting

import (
	"context"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

type ListMeetingsInput struct {
	Since       *time.Time
	Until       *time.Time
	Source      *string
	Participant *string
	Query       *string
	Limit       int
	Offset      int
}

type ListMeetingsOutput struct {
	Meetings []*domain.Meeting
	Total    int
}

type ListMeetings struct {
	repo domain.Repository
}

func NewListMeetings(repo domain.Repository) *ListMeetings {
	return &ListMeetings{repo: repo}
}

func (uc *ListMeetings) Execute(ctx context.Context, input ListMeetingsInput) (*ListMeetingsOutput, error) {
	filter := domain.ListFilter{
		Since:       input.Since,
		Until:       input.Until,
		Participant: input.Participant,
		Query:       input.Query,
		Limit:       input.Limit,
		Offset:      input.Offset,
	}

	if input.Source != nil {
		src := domain.Source(*input.Source)
		filter.Source = &src
	}

	meetings, err := uc.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &ListMeetingsOutput{
		Meetings: meetings,
		Total:    len(meetings),
	}, nil
}
