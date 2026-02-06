package annotation

import "time"

// DomainEvent is the contract for annotation domain events.
// Reuses the same pattern as meeting.DomainEvent but is defined
// separately because this is an independent bounded context.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// NoteAdded is raised when an agent note is created for a meeting.
type NoteAdded struct {
	noteID    string
	meetingID string
	author    string
	occurred  time.Time
}

func NewNoteAddedEvent(noteID, meetingID, author string) NoteAdded {
	return NoteAdded{
		noteID:    noteID,
		meetingID: meetingID,
		author:    author,
		occurred:  time.Now().UTC(),
	}
}

func (e NoteAdded) EventName() string     { return "note.added" }
func (e NoteAdded) OccurredAt() time.Time { return e.occurred }
func (e NoteAdded) NoteID() string        { return e.noteID }
func (e NoteAdded) MeetingID() string     { return e.meetingID }
func (e NoteAdded) Author() string        { return e.author }

// NoteDeleted is raised when an agent note is removed.
type NoteDeleted struct {
	noteID    string
	meetingID string
	occurred  time.Time
}

func NewNoteDeletedEvent(noteID, meetingID string) NoteDeleted {
	return NoteDeleted{
		noteID:    noteID,
		meetingID: meetingID,
		occurred:  time.Now().UTC(),
	}
}

func (e NoteDeleted) EventName() string     { return "note.deleted" }
func (e NoteDeleted) OccurredAt() time.Time { return e.occurred }
func (e NoteDeleted) NoteID() string        { return e.noteID }
func (e NoteDeleted) MeetingID() string     { return e.meetingID }
