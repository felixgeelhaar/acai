package events_test

import (
	"errors"
	"testing"

	"github.com/felixgeelhaar/acai/internal/infrastructure/events"
)

type mockSession struct {
	updatedURIs    []string
	listChangedCnt int
	err            error
}

func (m *mockSession) NotifyResourceUpdated(uri string) error {
	m.updatedURIs = append(m.updatedURIs, uri)
	return m.err
}

func (m *mockSession) NotifyResourceListChanged() error {
	m.listChangedCnt++
	return m.err
}

func TestMCPNotifier_NotifyResourceUpdated_CallsSessions(t *testing.T) {
	n := events.NewMCPNotifier()
	s1 := &mockSession{}
	s2 := &mockSession{}
	n.AddSession("s1", s1)
	n.AddSession("s2", s2)

	if err := n.NotifyResourceUpdated("meeting://m-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(s1.updatedURIs) != 1 || s1.updatedURIs[0] != "meeting://m-1" {
		t.Errorf("s1 got %v", s1.updatedURIs)
	}
	if len(s2.updatedURIs) != 1 || s2.updatedURIs[0] != "meeting://m-1" {
		t.Errorf("s2 got %v", s2.updatedURIs)
	}
}

func TestMCPNotifier_NotifyResourceListChanged_CallsSessions(t *testing.T) {
	n := events.NewMCPNotifier()
	s1 := &mockSession{}
	n.AddSession("s1", s1)

	if err := n.NotifyResourceListChanged(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s1.listChangedCnt != 1 {
		t.Errorf("expected 1 list changed, got %d", s1.listChangedCnt)
	}
}

func TestMCPNotifier_NoSessions_NoOp(t *testing.T) {
	n := events.NewMCPNotifier()

	if err := n.NotifyResourceUpdated("meeting://m-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := n.NotifyResourceListChanged(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMCPNotifier_SessionError_ContinuesOthers(t *testing.T) {
	n := events.NewMCPNotifier()
	failing := &mockSession{err: errors.New("session error")}
	healthy := &mockSession{}
	n.AddSession("failing", failing)
	n.AddSession("healthy", healthy)

	if err := n.NotifyResourceUpdated("meeting://m-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both should have been attempted
	if len(failing.updatedURIs) != 1 {
		t.Errorf("failing session: expected 1 attempt, got %d", len(failing.updatedURIs))
	}
	if len(healthy.updatedURIs) != 1 {
		t.Errorf("healthy session: expected 1 notification, got %d", len(healthy.updatedURIs))
	}
}

func TestMCPNotifier_RemoveSession(t *testing.T) {
	n := events.NewMCPNotifier()
	s := &mockSession{}
	n.AddSession("s1", s)
	n.RemoveSession("s1")

	if err := n.NotifyResourceUpdated("meeting://m-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(s.updatedURIs) != 0 {
		t.Errorf("removed session should not receive notifications, got %v", s.updatedURIs)
	}
}
