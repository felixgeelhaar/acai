package annotation

import (
	"context"

	annotatn "github.com/felixgeelhaar/acai/internal/domain/annotation"
	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

type DeleteNoteInput struct {
	NoteID string
}

type DeleteNoteOutput struct{}

type DeleteNote struct {
	noteRepo   annotatn.NoteRepository
	dispatcher domain.EventDispatcher
}

func NewDeleteNote(noteRepo annotatn.NoteRepository, dispatcher domain.EventDispatcher) *DeleteNote {
	return &DeleteNote{noteRepo: noteRepo, dispatcher: dispatcher}
}

func (uc *DeleteNote) Execute(ctx context.Context, input DeleteNoteInput) (*DeleteNoteOutput, error) {
	if input.NoteID == "" {
		return nil, annotatn.ErrInvalidNoteID
	}

	noteID := annotatn.NoteID(input.NoteID)

	// Find the note to get meeting ID for the event
	note, err := uc.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if err := uc.noteRepo.Delete(ctx, noteID); err != nil {
		return nil, err
	}

	// Dispatch event
	event := annotatn.NewNoteDeletedEvent(string(note.ID()), note.MeetingID())
	if uc.dispatcher != nil {
		if err := uc.dispatcher.Dispatch(ctx, []domain.DomainEvent{event}); err != nil {
			return nil, err
		}
	}

	return &DeleteNoteOutput{}, nil
}
