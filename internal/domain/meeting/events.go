package meeting

import "time"

// DomainEvent is the contract for all domain events raised by aggregates.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// MeetingCreated is raised when a new meeting aggregate is created.
type MeetingCreated struct {
	meetingID MeetingID
	title     string
	datetime  time.Time
	occurred  time.Time
}

func NewMeetingCreatedEvent(meetingID MeetingID, title string, datetime time.Time) MeetingCreated {
	return MeetingCreated{
		meetingID: meetingID,
		title:     title,
		datetime:  datetime,
		occurred:  time.Now().UTC(),
	}
}

func (e MeetingCreated) EventName() string     { return "meeting.created" }
func (e MeetingCreated) OccurredAt() time.Time { return e.occurred }
func (e MeetingCreated) MeetingID() MeetingID  { return e.meetingID }
func (e MeetingCreated) Title() string         { return e.title }

// TranscriptUpdated is raised when a transcript is attached or modified.
type TranscriptUpdated struct {
	meetingID      MeetingID
	utteranceCount int
	occurred       time.Time
}

func NewTranscriptUpdatedEvent(meetingID MeetingID, utteranceCount int) TranscriptUpdated {
	return TranscriptUpdated{
		meetingID:      meetingID,
		utteranceCount: utteranceCount,
		occurred:       time.Now().UTC(),
	}
}

func (e TranscriptUpdated) EventName() string     { return "transcript.updated" }
func (e TranscriptUpdated) OccurredAt() time.Time { return e.occurred }
func (e TranscriptUpdated) MeetingID() MeetingID  { return e.meetingID }
func (e TranscriptUpdated) UtteranceCount() int   { return e.utteranceCount }

// SummaryUpdated is raised when a summary is attached or modified.
type SummaryUpdated struct {
	meetingID MeetingID
	kind      SummaryKind
	occurred  time.Time
}

func NewSummaryUpdatedEvent(meetingID MeetingID, kind SummaryKind) SummaryUpdated {
	return SummaryUpdated{
		meetingID: meetingID,
		kind:      kind,
		occurred:  time.Now().UTC(),
	}
}

func (e SummaryUpdated) EventName() string     { return "summary.updated" }
func (e SummaryUpdated) OccurredAt() time.Time { return e.occurred }
func (e SummaryUpdated) MeetingID() MeetingID  { return e.meetingID }
func (e SummaryUpdated) Kind() SummaryKind     { return e.kind }
