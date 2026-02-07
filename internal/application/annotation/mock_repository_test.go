package annotation_test

import (
	"context"
	"time"

	annotatn "github.com/felixgeelhaar/acai/internal/domain/annotation"
	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

// mockNoteRepository implements annotation.NoteRepository for tests.
type mockNoteRepository struct {
	notes map[annotatn.NoteID]*annotatn.AgentNote
}

func newMockNoteRepository() *mockNoteRepository {
	return &mockNoteRepository{notes: make(map[annotatn.NoteID]*annotatn.AgentNote)}
}

func (m *mockNoteRepository) Save(_ context.Context, note *annotatn.AgentNote) error {
	m.notes[note.ID()] = note
	return nil
}

func (m *mockNoteRepository) FindByID(_ context.Context, id annotatn.NoteID) (*annotatn.AgentNote, error) {
	note, ok := m.notes[id]
	if !ok {
		return nil, annotatn.ErrNoteNotFound
	}
	return note, nil
}

func (m *mockNoteRepository) ListByMeeting(_ context.Context, meetingID string) ([]*annotatn.AgentNote, error) {
	var result []*annotatn.AgentNote
	for _, note := range m.notes {
		if note.MeetingID() == meetingID {
			result = append(result, note)
		}
	}
	if result == nil {
		result = []*annotatn.AgentNote{}
	}
	return result, nil
}

func (m *mockNoteRepository) Delete(_ context.Context, id annotatn.NoteID) error {
	if _, ok := m.notes[id]; !ok {
		return annotatn.ErrNoteNotFound
	}
	delete(m.notes, id)
	return nil
}

// mockMeetingRepository implements domain.Repository for verifying meeting existence.
type mockMeetingRepository struct {
	meetings map[domain.MeetingID]*domain.Meeting
}

func newMockMeetingRepository() *mockMeetingRepository {
	return &mockMeetingRepository{meetings: make(map[domain.MeetingID]*domain.Meeting)}
}

func (m *mockMeetingRepository) addMeeting(mtg *domain.Meeting) {
	m.meetings[mtg.ID()] = mtg
}

func (m *mockMeetingRepository) FindByID(_ context.Context, id domain.MeetingID) (*domain.Meeting, error) {
	mtg, ok := m.meetings[id]
	if !ok {
		return nil, domain.ErrMeetingNotFound
	}
	return mtg, nil
}

func (m *mockMeetingRepository) List(_ context.Context, _ domain.ListFilter) ([]*domain.Meeting, error) {
	return nil, nil
}

func (m *mockMeetingRepository) GetTranscript(_ context.Context, _ domain.MeetingID) (*domain.Transcript, error) {
	return nil, nil
}

func (m *mockMeetingRepository) SearchTranscripts(_ context.Context, _ string, _ domain.ListFilter) ([]*domain.Meeting, error) {
	return nil, nil
}

func (m *mockMeetingRepository) GetActionItems(_ context.Context, _ domain.MeetingID) ([]*domain.ActionItem, error) {
	return nil, nil
}

func (m *mockMeetingRepository) Sync(_ context.Context, _ *time.Time) ([]domain.DomainEvent, error) {
	return nil, nil
}

// mockDispatcher captures dispatched events.
type mockDispatcher struct {
	events []domain.DomainEvent
}

func (m *mockDispatcher) Dispatch(_ context.Context, events []domain.DomainEvent) error {
	m.events = append(m.events, events...)
	return nil
}
