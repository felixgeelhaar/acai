package meeting_test

import (
	"context"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

// mockRepository is a test double implementing domain.Repository.
// It allows tests to control returned data and verify interactions.
type mockRepository struct {
	meetings    map[domain.MeetingID]*domain.Meeting
	transcripts map[domain.MeetingID]*domain.Transcript
	actionItems map[domain.MeetingID][]*domain.ActionItem

	findByIDCalled       bool
	listCalled           bool
	getTranscriptCalled  bool
	searchCalled         bool
	getActionItemsCalled bool
	syncCalled           bool

	listFilter *domain.ListFilter
	syncSince  *time.Time
	syncEvents []domain.DomainEvent
	syncErr    error
	listErr    error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		meetings:    make(map[domain.MeetingID]*domain.Meeting),
		transcripts: make(map[domain.MeetingID]*domain.Transcript),
		actionItems: make(map[domain.MeetingID][]*domain.ActionItem),
	}
}

func (m *mockRepository) FindByID(_ context.Context, id domain.MeetingID) (*domain.Meeting, error) {
	m.findByIDCalled = true
	mtg, ok := m.meetings[id]
	if !ok {
		return nil, domain.ErrMeetingNotFound
	}
	return mtg, nil
}

func (m *mockRepository) List(_ context.Context, filter domain.ListFilter) ([]*domain.Meeting, error) {
	m.listCalled = true
	m.listFilter = &filter

	if m.listErr != nil {
		return nil, m.listErr
	}

	result := make([]*domain.Meeting, 0, len(m.meetings))
	for _, mtg := range m.meetings {
		result = append(result, mtg)
	}

	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (m *mockRepository) GetTranscript(_ context.Context, id domain.MeetingID) (*domain.Transcript, error) {
	m.getTranscriptCalled = true
	t, ok := m.transcripts[id]
	if !ok {
		return nil, domain.ErrTranscriptNotReady
	}
	return t, nil
}

func (m *mockRepository) SearchTranscripts(_ context.Context, query string, filter domain.ListFilter) ([]*domain.Meeting, error) {
	m.searchCalled = true
	result := make([]*domain.Meeting, 0)
	for _, mtg := range m.meetings {
		result = append(result, mtg)
	}
	return result, nil
}

func (m *mockRepository) GetActionItems(_ context.Context, id domain.MeetingID) ([]*domain.ActionItem, error) {
	m.getActionItemsCalled = true
	items, ok := m.actionItems[id]
	if !ok {
		return []*domain.ActionItem{}, nil
	}
	return items, nil
}

func (m *mockRepository) Sync(_ context.Context, since *time.Time) ([]domain.DomainEvent, error) {
	m.syncCalled = true
	m.syncSince = since
	if m.syncErr != nil {
		return nil, m.syncErr
	}
	return m.syncEvents, nil
}

func (m *mockRepository) addMeeting(mtg *domain.Meeting) {
	m.meetings[mtg.ID()] = mtg
}

func (m *mockRepository) addTranscript(id domain.MeetingID, t *domain.Transcript) {
	m.transcripts[id] = t
}

func (m *mockRepository) addActionItems(id domain.MeetingID, items []*domain.ActionItem) {
	m.actionItems[id] = items
}

// mockWriteRepository implements domain.WriteRepository for tests.
type mockWriteRepository struct {
	items map[domain.ActionItemID]*domain.ActionItem
}

func newMockWriteRepository() *mockWriteRepository {
	return &mockWriteRepository{items: make(map[domain.ActionItemID]*domain.ActionItem)}
}

func (m *mockWriteRepository) SaveActionItemState(_ context.Context, item *domain.ActionItem) error {
	m.items[item.ID()] = item
	return nil
}

func (m *mockWriteRepository) GetLocalActionItemState(_ context.Context, id domain.ActionItemID) (*domain.ActionItem, error) {
	item, ok := m.items[id]
	if !ok {
		return nil, domain.ErrMeetingNotFound
	}
	return item, nil
}

// mockDispatcher captures dispatched events.
type mockDispatcher struct {
	events []domain.DomainEvent
}

func (m *mockDispatcher) Dispatch(_ context.Context, events []domain.DomainEvent) error {
	m.events = append(m.events, events...)
	return nil
}
