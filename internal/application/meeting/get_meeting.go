package meeting

import (
	"context"

	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

type GetMeetingInput struct {
	ID domain.MeetingID
}

type GetMeetingOutput struct {
	Meeting *domain.Meeting
}

type GetMeeting struct {
	repo domain.Repository
}

func NewGetMeeting(repo domain.Repository) *GetMeeting {
	return &GetMeeting{repo: repo}
}

func (uc *GetMeeting) Execute(ctx context.Context, input GetMeetingInput) (*GetMeetingOutput, error) {
	if input.ID == "" {
		return nil, domain.ErrInvalidMeetingID
	}

	mtg, err := uc.repo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	return &GetMeetingOutput{Meeting: mtg}, nil
}
