package meeting

import (
	"context"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

type CompleteActionItemInput struct {
	MeetingID    domain.MeetingID
	ActionItemID domain.ActionItemID
}

type CompleteActionItemOutput struct {
	Item *domain.ActionItem
}

type CompleteActionItem struct {
	repo       domain.Repository
	writeRepo  domain.WriteRepository
	dispatcher domain.EventDispatcher
}

func NewCompleteActionItem(repo domain.Repository, writeRepo domain.WriteRepository, dispatcher domain.EventDispatcher) *CompleteActionItem {
	return &CompleteActionItem{repo: repo, writeRepo: writeRepo, dispatcher: dispatcher}
}

func (uc *CompleteActionItem) Execute(ctx context.Context, input CompleteActionItemInput) (*CompleteActionItemOutput, error) {
	if input.MeetingID == "" {
		return nil, domain.ErrInvalidMeetingID
	}
	if input.ActionItemID == "" {
		return nil, domain.ErrInvalidActionItemID
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

	// Apply local override
	item.Complete()

	if err := uc.writeRepo.SaveActionItemState(ctx, item); err != nil {
		return nil, err
	}

	// Dispatch event
	event := domain.NewActionItemCompletedEvent(input.MeetingID, input.ActionItemID)
	if uc.dispatcher != nil {
		if err := uc.dispatcher.Dispatch(ctx, []domain.DomainEvent{event}); err != nil {
			return nil, err
		}
	}

	return &CompleteActionItemOutput{Item: item}, nil
}
