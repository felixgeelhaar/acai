package annotation

import (
	"context"

	annotatn "github.com/felixgeelhaar/granola-mcp/internal/domain/annotation"
)

type ListNotesInput struct {
	MeetingID string
}

type ListNotesOutput struct {
	Notes []*annotatn.AgentNote
}

type ListNotes struct {
	noteRepo annotatn.NoteRepository
}

func NewListNotes(noteRepo annotatn.NoteRepository) *ListNotes {
	return &ListNotes{noteRepo: noteRepo}
}

func (uc *ListNotes) Execute(ctx context.Context, input ListNotesInput) (*ListNotesOutput, error) {
	if input.MeetingID == "" {
		return nil, annotatn.ErrInvalidMeetingID
	}

	notes, err := uc.noteRepo.ListByMeeting(ctx, input.MeetingID)
	if err != nil {
		return nil, err
	}

	return &ListNotesOutput{Notes: notes}, nil
}
