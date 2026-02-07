package events

import (
	"log"
	"sync"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

// SessionNotifier can send notifications to a specific MCP session.
type SessionNotifier interface {
	NotifyResourceUpdated(uri string) error
	NotifyResourceListChanged() error
}

// MCPNotifier implements domain.EventNotifier by broadcasting
// to all registered MCP sessions.
type MCPNotifier struct {
	mu       sync.RWMutex
	sessions map[string]SessionNotifier
}

// NewMCPNotifier creates a new MCP notifier.
func NewMCPNotifier() *MCPNotifier {
	return &MCPNotifier{
		sessions: make(map[string]SessionNotifier),
	}
}

// AddSession registers a session for notifications.
func (n *MCPNotifier) AddSession(id string, session SessionNotifier) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.sessions[id] = session
}

// RemoveSession unregisters a session.
func (n *MCPNotifier) RemoveSession(id string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.sessions, id)
}

// NotifyResourceUpdated broadcasts a resource update to all sessions.
func (n *MCPNotifier) NotifyResourceUpdated(uri string) error {
	n.mu.RLock()
	sessions := make(map[string]SessionNotifier, len(n.sessions))
	for id, s := range n.sessions {
		sessions[id] = s
	}
	n.mu.RUnlock()

	for id, s := range sessions {
		if err := s.NotifyResourceUpdated(uri); err != nil {
			log.Printf("notify session %s resource updated %q: %v", id, uri, err)
		}
	}
	return nil
}

// NotifyResourceListChanged broadcasts a resource list change to all sessions.
func (n *MCPNotifier) NotifyResourceListChanged() error {
	n.mu.RLock()
	sessions := make(map[string]SessionNotifier, len(n.sessions))
	for id, s := range n.sessions {
		sessions[id] = s
	}
	n.mu.RUnlock()

	for id, s := range sessions {
		if err := s.NotifyResourceListChanged(); err != nil {
			log.Printf("notify session %s resource list changed: %v", id, err)
		}
	}
	return nil
}

// compile-time check
var _ domain.EventNotifier = (*MCPNotifier)(nil)
