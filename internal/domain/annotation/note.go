package annotation

import "time"

// NoteID is a strongly-typed identifier for agent notes.
type NoteID string

// AgentNote is an entity representing a note added by an MCP agent to a meeting.
// Notes belong to the annotation bounded context, not the meeting aggregate.
// Cross-context reference is by meeting ID string only.
type AgentNote struct {
	id        NoteID
	meetingID string
	author    string
	content   string
	createdAt time.Time
}

// NewAgentNote constructs a valid AgentNote, enforcing creation invariants.
func NewAgentNote(id NoteID, meetingID, author, content string) (*AgentNote, error) {
	if id == "" {
		return nil, ErrInvalidNoteID
	}
	if meetingID == "" {
		return nil, ErrInvalidMeetingID
	}
	if author == "" {
		return nil, ErrInvalidAuthor
	}
	if content == "" {
		return nil, ErrInvalidNoteContent
	}

	return &AgentNote{
		id:        id,
		meetingID: meetingID,
		author:    author,
		content:   content,
		createdAt: time.Now().UTC(),
	}, nil
}

// ReconstructAgentNote reconstitutes a note from persistence without raising events.
func ReconstructAgentNote(id NoteID, meetingID, author, content string, createdAt time.Time) *AgentNote {
	return &AgentNote{
		id:        id,
		meetingID: meetingID,
		author:    author,
		content:   content,
		createdAt: createdAt,
	}
}

func (n *AgentNote) ID() NoteID       { return n.id }
func (n *AgentNote) MeetingID() string { return n.meetingID }
func (n *AgentNote) Author() string    { return n.author }
func (n *AgentNote) Content() string   { return n.content }
func (n *AgentNote) CreatedAt() time.Time { return n.createdAt }
