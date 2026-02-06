package meeting

import "time"

// ActionItemID is a strongly-typed identifier for action items.
type ActionItemID string

// ActionItem is an entity with identity, belonging to a Meeting aggregate.
// Entities are compared by identity (ID), not by attribute values.
type ActionItem struct {
	id        ActionItemID
	meetingID MeetingID
	owner     string
	text      string
	dueDate   *time.Time
	completed bool
}

func NewActionItem(id ActionItemID, meetingID MeetingID, owner, text string, dueDate *time.Time) (*ActionItem, error) {
	if id == "" {
		return nil, ErrInvalidActionItemID
	}
	if text == "" {
		return nil, ErrInvalidActionItemText
	}

	var due *time.Time
	if dueDate != nil {
		d := *dueDate
		due = &d
	}

	return &ActionItem{
		id:        id,
		meetingID: meetingID,
		owner:     owner,
		text:      text,
		dueDate:   due,
		completed: false,
	}, nil
}

func (a *ActionItem) ID() ActionItemID   { return a.id }
func (a *ActionItem) MeetingID() MeetingID { return a.meetingID }
func (a *ActionItem) Owner() string       { return a.owner }
func (a *ActionItem) Text() string        { return a.text }

func (a *ActionItem) DueDate() *time.Time {
	if a.dueDate == nil {
		return nil
	}
	d := *a.dueDate
	return &d
}

func (a *ActionItem) IsCompleted() bool { return a.completed }

// Complete marks the action item as done. This is a domain behavior on the entity.
func (a *ActionItem) Complete() {
	a.completed = true
}
