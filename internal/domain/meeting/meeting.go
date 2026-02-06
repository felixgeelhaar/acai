// Package meeting contains the Meeting bounded context.
// The Meeting aggregate root is the consistency boundary for all meeting-related data.
package meeting

import "time"

// MeetingID is a strongly-typed identifier for meetings.
type MeetingID string

// Meeting is the aggregate root for meeting data.
// All mutations go through the aggregate root to enforce invariants.
// Domain events are collected and can be dispatched after persistence.
type Meeting struct {
	id           MeetingID
	title        string
	datetime     time.Time
	source       Source
	participants []Participant
	transcript   *Transcript
	summary      *Summary
	actionItems  []*ActionItem
	metadata     Metadata
	createdAt    time.Time
	updatedAt    time.Time
	events       []DomainEvent
}

// New constructs a valid Meeting aggregate, enforcing all creation invariants.
func New(id MeetingID, title string, datetime time.Time, source Source, participants []Participant) (*Meeting, error) {
	if id == "" {
		return nil, ErrInvalidMeetingID
	}
	if title == "" {
		return nil, ErrInvalidTitle
	}
	if datetime.IsZero() {
		return nil, ErrInvalidDatetime
	}

	now := time.Now().UTC()
	ps := make([]Participant, len(participants))
	copy(ps, participants)

	m := &Meeting{
		id:           id,
		title:        title,
		datetime:     datetime,
		source:       source,
		participants: ps,
		actionItems:  make([]*ActionItem, 0),
		metadata:     NewMetadata(nil, nil, nil),
		createdAt:    now,
		updatedAt:    now,
	}

	m.events = append(m.events, NewMeetingCreatedEvent(id, title, datetime))

	return m, nil
}

// --- Accessors ---

func (m *Meeting) ID() MeetingID       { return m.id }
func (m *Meeting) Title() string        { return m.title }
func (m *Meeting) Datetime() time.Time  { return m.datetime }
func (m *Meeting) Source() Source        { return m.source }
func (m *Meeting) CreatedAt() time.Time { return m.createdAt }
func (m *Meeting) UpdatedAt() time.Time { return m.updatedAt }
func (m *Meeting) Metadata() Metadata   { return m.metadata }

func (m *Meeting) Participants() []Participant {
	copied := make([]Participant, len(m.participants))
	copy(copied, m.participants)
	return copied
}

func (m *Meeting) Transcript() *Transcript {
	return m.transcript
}

func (m *Meeting) Summary() *Summary {
	return m.summary
}

func (m *Meeting) ActionItems() []*ActionItem {
	copied := make([]*ActionItem, len(m.actionItems))
	copy(copied, m.actionItems)
	return copied
}

// --- Aggregate Behaviors ---

// AttachTranscript sets the transcript on the meeting and raises a TranscriptUpdated event.
func (m *Meeting) AttachTranscript(t Transcript) []DomainEvent {
	m.transcript = &t
	m.updatedAt = time.Now().UTC()

	event := NewTranscriptUpdatedEvent(m.id, len(t.Utterances()))
	m.events = append(m.events, event)
	return []DomainEvent{event}
}

// AttachSummary sets the summary on the meeting and raises a SummaryUpdated event.
func (m *Meeting) AttachSummary(s Summary) []DomainEvent {
	m.summary = &s
	m.updatedAt = time.Now().UTC()

	event := NewSummaryUpdatedEvent(m.id, s.Kind())
	m.events = append(m.events, event)
	return []DomainEvent{event}
}

// AddActionItem appends an action item to the meeting aggregate.
func (m *Meeting) AddActionItem(item *ActionItem) {
	m.actionItems = append(m.actionItems, item)
	m.updatedAt = time.Now().UTC()
}

// SetMetadata replaces the meeting metadata.
func (m *Meeting) SetMetadata(meta Metadata) {
	m.metadata = meta
	m.updatedAt = time.Now().UTC()
}

// --- Domain Events ---

// DomainEvents returns the uncommitted domain events.
func (m *Meeting) DomainEvents() []DomainEvent {
	copied := make([]DomainEvent, len(m.events))
	copy(copied, m.events)
	return copied
}

// ClearDomainEvents removes all collected events (call after dispatching).
func (m *Meeting) ClearDomainEvents() {
	m.events = m.events[:0]
}
