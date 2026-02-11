package annotation

import "context"

// NoteRepository is the port for agent note persistence.
// Defined in the domain layer, implemented in infrastructure (local store).
type NoteRepository interface {
	Save(ctx context.Context, note *AgentNote) error
	FindByID(ctx context.Context, id NoteID) (*AgentNote, error)
	ListByMeeting(ctx context.Context, meetingID string) ([]*AgentNote, error)
	ListAll(ctx context.Context) ([]*AgentNote, error)
	Delete(ctx context.Context, id NoteID) error
}
