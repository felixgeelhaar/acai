package meeting

import (
	"context"

	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

type GetActionItemsInput struct {
	MeetingID domain.MeetingID
}

type GetActionItemsOutput struct {
	Items []*domain.ActionItem
}

type GetActionItems struct {
	repo domain.Repository
}

func NewGetActionItems(repo domain.Repository) *GetActionItems {
	return &GetActionItems{repo: repo}
}

func (uc *GetActionItems) Execute(ctx context.Context, input GetActionItemsInput) (*GetActionItemsOutput, error) {
	if input.MeetingID == "" {
		return nil, domain.ErrInvalidMeetingID
	}

	items, err := uc.repo.GetActionItems(ctx, input.MeetingID)
	if err != nil {
		return nil, err
	}

	return &GetActionItemsOutput{Items: items}, nil
}
