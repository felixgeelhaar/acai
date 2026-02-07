package meeting

import (
	"context"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

type UpdateActionItemInput struct {
	MeetingID    domain.MeetingID
	ActionItemID domain.ActionItemID
	Text         string
}

type UpdateActionItemOutput struct {
	Item *domain.ActionItem
}

type UpdateActionItem struct {
	repo       domain.Repository
	writeRepo  domain.WriteRepository
	dispatcher domain.EventDispatcher
}

func NewUpdateActionItem(repo domain.Repository, writeRepo domain.WriteRepository, dispatcher domain.EventDispatcher) *UpdateActionItem {
	return &UpdateActionItem{repo: repo, writeRepo: writeRepo, dispatcher: dispatcher}
}

func (uc *UpdateActionItem) Execute(ctx context.Context, input UpdateActionItemInput) (*UpdateActionItemOutput, error) {
	if input.MeetingID == "" {
		return nil, domain.ErrInvalidMeetingID
	}
	if input.ActionItemID == "" {
		return nil, domain.ErrInvalidActionItemID
	}
	if input.Text == "" {
		return nil, domain.ErrInvalidActionItemText
	}

	// Read the action item from the upstream repo
	items, err := uc.repo.GetActionItems(ctx, input.MeetingID)
	if err != nil {
		return nil, err
	}

	var item *domain.ActionItem
	for _, ai := range items {
		if ai.ID() == input.ActionItemID {
			item = ai
			break
		}
	}
	if item == nil {
		return nil, domain.ErrMeetingNotFound
	}

	// Apply text update
	if err := item.UpdateText(input.Text); err != nil {
		return nil, err
	}

	if err := uc.writeRepo.SaveActionItemState(ctx, item); err != nil {
		return nil, err
	}

	// Dispatch event
	event := domain.NewActionItemUpdatedEvent(input.MeetingID, input.ActionItemID, input.Text)
	if uc.dispatcher != nil {
		if err := uc.dispatcher.Dispatch(ctx, []domain.DomainEvent{event}); err != nil {
			return nil, err
		}
	}

	return &UpdateActionItemOutput{Item: item}, nil
}
