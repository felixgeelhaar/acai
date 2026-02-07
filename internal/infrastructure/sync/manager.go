// Package sync provides background meeting synchronization.
package sync

import (
	"context"
	"log"
	"sync"
	"time"

	meetingapp "github.com/felixgeelhaar/acai/internal/application/meeting"
	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

// Manager runs periodic meeting sync in the background.
type Manager struct {
	syncUC     *meetingapp.SyncMeetings
	dispatcher domain.EventDispatcher
	interval   time.Duration

	mu            sync.Mutex
	lastSyncTime  *time.Time
	cancel        context.CancelFunc
	done          chan struct{}
}

// NewManager creates a new sync manager.
func NewManager(syncUC *meetingapp.SyncMeetings, dispatcher domain.EventDispatcher, interval time.Duration) *Manager {
	return &Manager{
		syncUC:     syncUC,
		dispatcher: dispatcher,
		interval:   interval,
	}
}

// Start launches the background sync goroutine.
// It returns immediately. Call Stop to shut down gracefully.
func (m *Manager) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.done = make(chan struct{})

	go m.run(ctx)
}

// Stop gracefully shuts down the sync manager.
func (m *Manager) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
	if m.done != nil {
		<-m.done
	}
}

func (m *Manager) run(ctx context.Context) {
	defer close(m.done)

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.tick(ctx)
		}
	}
}

func (m *Manager) tick(ctx context.Context) {
	m.mu.Lock()
	since := m.lastSyncTime
	m.mu.Unlock()

	out, err := m.syncUC.Execute(ctx, meetingapp.SyncMeetingsInput{Since: since})
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		log.Printf("sync manager: sync failed: %v", err)
		return
	}

	now := time.Now().UTC()
	m.mu.Lock()
	m.lastSyncTime = &now
	m.mu.Unlock()

	if len(out.Events) == 0 {
		return
	}

	if err := m.dispatcher.Dispatch(ctx, out.Events); err != nil {
		log.Printf("sync manager: dispatch failed: %v", err)
	}
}
