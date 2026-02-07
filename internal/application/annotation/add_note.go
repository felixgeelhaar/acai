// Package annotation contains use cases for the annotation bounded context.
package annotation

import (
	"context"
	"fmt"
	"time"

	"github.com/felixgeelhaar/acai/internal/domain/annotation"
	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

type AddNoteInput struct {
	MeetingID string
	Author    string
	Content   string
}

type AddNoteOutput struct {
	Note *annotation.AgentNote
}

type AddNote struct {
	noteRepo   annotation.NoteRepository
	meetingRepo domain.Repository
	dispatcher domain.EventDispatcher
}

func NewAddNote(noteRepo annotation.NoteRepository, meetingRepo domain.Repository, dispatcher domain.EventDispatcher) *AddNote {
	return &AddNote{noteRepo: noteRepo, meetingRepo: meetingRepo, dispatcher: dispatcher}
}

func (uc *AddNote) Execute(ctx context.Context, input AddNoteInput) (*AddNoteOutput, error) {
	// Verify meeting exists
	if input.MeetingID == "" {
		return nil, annotation.ErrInvalidMeetingID
	}
	if _, err := uc.meetingRepo.FindByID(ctx, domain.MeetingID(input.MeetingID)); err != nil {
		return nil, err
	}

	// Generate note ID
	noteID := annotation.NoteID(fmt.Sprintf("note-%d", time.Now().UnixNano()))

	note, err := annotation.NewAgentNote(noteID, input.MeetingID, input.Author, input.Content)
	if err != nil {
		return nil, err
	}

	if err := uc.noteRepo.Save(ctx, note); err != nil {
		return nil, err
	}

	// Dispatch event (annotation events satisfy meeting.DomainEvent via structural typing)
	event := annotation.NewNoteAddedEvent(string(note.ID()), note.MeetingID(), note.Author())
	if uc.dispatcher != nil {
		if err := uc.dispatcher.Dispatch(ctx, []domain.DomainEvent{event}); err != nil {
			return nil, err
		}
	}

	return &AddNoteOutput{Note: note}, nil
}
