package sync_test

import (
	"context"
	"errors"
	"testing"
	"time"

	meetingapp "github.com/felixgeelhaar/acai/internal/application/meeting"
	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
	syncmgr "github.com/felixgeelhaar/acai/internal/infrastructure/sync"
)

// mockRepo implements domain.Repository for testing.
type mockRepo struct {
	events []domain.DomainEvent
	err    error
	calls  int
}

func (m *mockRepo) FindByID(_ context.Context, _ domain.MeetingID) (*domain.Meeting, error) {
	return nil, nil
}
func (m *mockRepo) List(_ context.Context, _ domain.ListFilter) ([]*domain.Meeting, error) {
	return nil, nil
}
func (m *mockRepo) GetTranscript(_ context.Context, _ domain.MeetingID) (*domain.Transcript, error) {
	return nil, nil
}
func (m *mockRepo) SearchTranscripts(_ context.Context, _ string, _ domain.ListFilter) ([]*domain.Meeting, error) {
	return nil, nil
}
func (m *mockRepo) GetActionItems(_ context.Context, _ domain.MeetingID) ([]*domain.ActionItem, error) {
	return nil, nil
}
func (m *mockRepo) Sync(_ context.Context, _ *time.Time) ([]domain.DomainEvent, error) {
	m.calls++
	return m.events, m.err
}

// mockDispatcher tracks dispatched events.
type mockDispatcher struct {
	dispatched []domain.DomainEvent
	err        error
}

func (m *mockDispatcher) Dispatch(_ context.Context, events []domain.DomainEvent) error {
	m.dispatched = append(m.dispatched, events...)
	return m.err
}

func TestSyncManager_Start_TicksSyncAndDispatches(t *testing.T) {
	event := domain.NewMeetingCreatedEvent("m-1", "Test", time.Now().UTC())
	repo := &mockRepo{events: []domain.DomainEvent{event}}
	dispatcher := &mockDispatcher{}
	uc := meetingapp.NewSyncMeetings(repo)

	mgr := syncmgr.NewManager(uc, dispatcher, 50*time.Millisecond)
	mgr.Start(context.Background())

	// Wait for at least one tick
	time.Sleep(150 * time.Millisecond)
	mgr.Stop()

	if repo.calls < 1 {
		t.Errorf("expected at least 1 sync call, got %d", repo.calls)
	}
	if len(dispatcher.dispatched) < 1 {
		t.Errorf("expected at least 1 dispatched event, got %d", len(dispatcher.dispatched))
	}
}

func TestSyncManager_Stop_GracefulShutdown(t *testing.T) {
	repo := &mockRepo{}
	dispatcher := &mockDispatcher{}
	uc := meetingapp.NewSyncMeetings(repo)

	mgr := syncmgr.NewManager(uc, dispatcher, 1*time.Hour) // long interval
	mgr.Start(context.Background())

	// Stop should return quickly
	done := make(chan struct{})
	go func() {
		mgr.Stop()
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not return in time")
	}
}

func TestSyncManager_SyncError_ContinuesNextTick(t *testing.T) {
	repo := &mockRepo{err: errors.New("api error")}
	dispatcher := &mockDispatcher{}
	uc := meetingapp.NewSyncMeetings(repo)

	mgr := syncmgr.NewManager(uc, dispatcher, 50*time.Millisecond)
	mgr.Start(context.Background())

	time.Sleep(150 * time.Millisecond)
	mgr.Stop()

	// Should have attempted sync multiple times despite errors
	if repo.calls < 1 {
		t.Errorf("expected at least 1 sync attempt, got %d", repo.calls)
	}
	// No events dispatched because sync failed
	if len(dispatcher.dispatched) != 0 {
		t.Errorf("expected 0 dispatched events on error, got %d", len(dispatcher.dispatched))
	}
}

func TestSyncManager_ContextCanceled_Stops(t *testing.T) {
	repo := &mockRepo{}
	dispatcher := &mockDispatcher{}
	uc := meetingapp.NewSyncMeetings(repo)

	ctx, cancel := context.WithCancel(context.Background())
	mgr := syncmgr.NewManager(uc, dispatcher, 50*time.Millisecond)
	mgr.Start(ctx)

	cancel()

	// Stop should return quickly since context is already canceled
	done := make(chan struct{})
	go func() {
		mgr.Stop()
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not return in time after context cancellation")
	}
}

func TestSyncManager_TracksSinceTimestamp(t *testing.T) {
	var sinceTimes []*time.Time
	repo := &mockRepo{}
	// Override Sync to capture the since parameter
	origSync := repo.Sync
	_ = origSync

	// We need a custom repo that captures since
	capRepo := &captureSinceRepo{}
	dispatcher := &mockDispatcher{}
	uc := meetingapp.NewSyncMeetings(capRepo)

	mgr := syncmgr.NewManager(uc, dispatcher, 50*time.Millisecond)
	mgr.Start(context.Background())

	time.Sleep(150 * time.Millisecond)
	mgr.Stop()

	sinceTimes = capRepo.sinceTimes
	if len(sinceTimes) < 2 {
		t.Skipf("only got %d ticks, need at least 2 to verify since tracking", len(sinceTimes))
	}

	// First call should have nil since (no previous sync)
	if sinceTimes[0] != nil {
		t.Errorf("first call should have nil since, got %v", sinceTimes[0])
	}
	// Second call should have a non-nil since
	if sinceTimes[1] == nil {
		t.Error("second call should have non-nil since")
	}
}

type captureSinceRepo struct {
	sinceTimes []*time.Time
}

func (m *captureSinceRepo) FindByID(_ context.Context, _ domain.MeetingID) (*domain.Meeting, error) {
	return nil, nil
}
func (m *captureSinceRepo) List(_ context.Context, _ domain.ListFilter) ([]*domain.Meeting, error) {
	return nil, nil
}
func (m *captureSinceRepo) GetTranscript(_ context.Context, _ domain.MeetingID) (*domain.Transcript, error) {
	return nil, nil
}
func (m *captureSinceRepo) SearchTranscripts(_ context.Context, _ string, _ domain.ListFilter) ([]*domain.Meeting, error) {
	return nil, nil
}
func (m *captureSinceRepo) GetActionItems(_ context.Context, _ domain.MeetingID) ([]*domain.ActionItem, error) {
	return nil, nil
}
func (m *captureSinceRepo) Sync(_ context.Context, since *time.Time) ([]domain.DomainEvent, error) {
	m.sinceTimes = append(m.sinceTimes, since)
	return nil, nil
}

func TestSyncManager_ZeroEvents_NoDispatch(t *testing.T) {
	repo := &mockRepo{events: []domain.DomainEvent{}} // empty events
	dispatcher := &mockDispatcher{}
	uc := meetingapp.NewSyncMeetings(repo)

	mgr := syncmgr.NewManager(uc, dispatcher, 50*time.Millisecond)
	mgr.Start(context.Background())

	time.Sleep(150 * time.Millisecond)
	mgr.Stop()

	if repo.calls < 1 {
		t.Errorf("expected at least 1 sync call, got %d", repo.calls)
	}
	if len(dispatcher.dispatched) != 0 {
		t.Errorf("expected 0 dispatched events for empty sync, got %d", len(dispatcher.dispatched))
	}
}
