package localcache

import (
	"context"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

// Repository implements domain.Repository by reading from the Granola desktop
// app's local cache file. All filtering is done in-memory.
type Repository struct {
	reader     *Reader
	mu         sync.RWMutex
	state      *CacheState
	prevDocIDs map[string]string // doc ID → updatedAt for Sync change detection
}

// NewRepository creates a Repository backed by the given Reader.
func NewRepository(reader *Reader) *Repository {
	return &Repository{
		reader:     reader,
		prevDocIDs: make(map[string]string),
	}
}

// ensureLoaded lazily loads (or reloads) the cache state.
func (r *Repository) ensureLoaded() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state != nil {
		return nil
	}
	return r.loadLocked()
}

// loadLocked reads the cache file into memory. Caller must hold r.mu write lock.
func (r *Repository) loadLocked() error {
	state, err := r.reader.Read()
	if err != nil {
		return err
	}
	r.state = state
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id domain.MeetingID) (*domain.Meeting, error) {
	if err := r.ensureLoaded(); err != nil {
		return nil, r.mapError(err)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	doc, ok := r.state.State.Documents[string(id)]
	if !ok {
		return nil, domain.ErrMeetingNotFound
	}

	meta := r.metaFor(string(id))
	mtg, err := mapDocumentToDomain(doc, meta)
	if err != nil {
		return nil, err
	}

	// Attach transcript if available
	if transcript, ok := r.state.State.Transcripts[string(id)]; ok {
		t := mapTranscriptToDomain(string(id), transcript)
		if t != nil {
			mtg.AttachTranscript(*t)
			mtg.ClearDomainEvents()
		}
	}

	return mtg, nil
}

func (r *Repository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Meeting, error) {
	if err := r.ensureLoaded(); err != nil {
		return nil, r.mapError(err)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var meetings []*domain.Meeting
	for id, doc := range r.state.State.Documents {
		meta := r.metaFor(id)
		mtg, err := mapDocumentToDomain(doc, meta)
		if err != nil {
			log.Printf("localcache: skipping invalid document %s: %v", id, err)
			continue
		}

		if !r.matchesFilter(mtg, filter) {
			continue
		}

		meetings = append(meetings, mtg)
	}

	// Sort by Datetime descending (newest first)
	// Note: domain.New() sets CreatedAt to time.Now(), so we sort by Datetime
	// which holds the actual meeting timestamp from the cache.
	sort.Slice(meetings, func(i, j int) bool {
		return meetings[i].Datetime().After(meetings[j].Datetime())
	})

	// Apply offset
	if filter.Offset > 0 && filter.Offset < len(meetings) {
		meetings = meetings[filter.Offset:]
	} else if filter.Offset >= len(meetings) {
		return []*domain.Meeting{}, nil
	}

	// Apply limit
	if filter.Limit > 0 && filter.Limit < len(meetings) {
		meetings = meetings[:filter.Limit]
	}

	return meetings, nil
}

func (r *Repository) GetTranscript(ctx context.Context, id domain.MeetingID) (*domain.Transcript, error) {
	if err := r.ensureLoaded(); err != nil {
		return nil, r.mapError(err)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Verify the meeting exists
	if _, ok := r.state.State.Documents[string(id)]; !ok {
		return nil, domain.ErrMeetingNotFound
	}

	transcript, ok := r.state.State.Transcripts[string(id)]
	if !ok {
		return nil, domain.ErrTranscriptNotReady
	}

	t := mapTranscriptToDomain(string(id), transcript)
	if t == nil {
		return nil, domain.ErrTranscriptNotReady
	}

	return t, nil
}

func (r *Repository) SearchTranscripts(ctx context.Context, query string, filter domain.ListFilter) ([]*domain.Meeting, error) {
	if err := r.ensureLoaded(); err != nil {
		return nil, r.mapError(err)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	queryLower := strings.ToLower(query)
	var meetings []*domain.Meeting

	for id, doc := range r.state.State.Documents {
		meta := r.metaFor(id)
		mtg, err := mapDocumentToDomain(doc, meta)
		if err != nil {
			continue
		}

		if !r.matchesFilter(mtg, filter) {
			continue
		}

		// Match against title
		if strings.Contains(strings.ToLower(doc.Title), queryLower) {
			meetings = append(meetings, mtg)
			continue
		}

		// Match against transcript text
		if transcript, ok := r.state.State.Transcripts[id]; ok {
			if transcriptContains(transcript, queryLower) {
				meetings = append(meetings, mtg)
				continue
			}
		}

		// Match against notes (ProseMirror)
		if len(doc.NotesProsemirror) > 0 {
			plainText := prosemirrorToPlainText(doc.NotesProsemirror)
			if strings.Contains(strings.ToLower(plainText), queryLower) {
				meetings = append(meetings, mtg)
				continue
			}
		}
	}

	sort.Slice(meetings, func(i, j int) bool {
		return meetings[i].Datetime().After(meetings[j].Datetime())
	})

	if filter.Offset > 0 && filter.Offset < len(meetings) {
		meetings = meetings[filter.Offset:]
	} else if filter.Offset >= len(meetings) {
		return []*domain.Meeting{}, nil
	}

	if filter.Limit > 0 && filter.Limit < len(meetings) {
		meetings = meetings[:filter.Limit]
	}

	return meetings, nil
}

func (r *Repository) GetActionItems(_ context.Context, _ domain.MeetingID) ([]*domain.ActionItem, error) {
	// Action items are not stored in the local cache file.
	return []*domain.ActionItem{}, nil
}

func (r *Repository) Sync(ctx context.Context, since *time.Time) ([]domain.DomainEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Force reload from disk
	if err := r.loadLocked(); err != nil {
		return nil, r.mapError(err)
	}

	var events []domain.DomainEvent
	currentDocIDs := make(map[string]string, len(r.state.State.Documents))

	for id, doc := range r.state.State.Documents {
		currentDocIDs[id] = doc.UpdatedAt

		createdAt, err := time.Parse(cacheTimestampLayout, doc.CreatedAt)
		if err != nil {
			createdAt = time.Now().UTC()
		}

		// Skip documents older than since
		if since != nil && createdAt.Before(*since) {
			// Check if the document was updated since the last sync
			prevUpdatedAt, existed := r.prevDocIDs[id]
			if !existed || prevUpdatedAt != doc.UpdatedAt {
				events = append(events, domain.NewSummaryUpdatedEvent(
					domain.MeetingID(id),
					domain.SummaryAuto,
				))
			}
			continue
		}

		// New document or not previously seen
		if _, existed := r.prevDocIDs[id]; !existed {
			events = append(events, domain.NewMeetingCreatedEvent(
				domain.MeetingID(id),
				doc.Title,
				createdAt,
			))
		}
	}

	r.prevDocIDs = currentDocIDs
	return events, nil
}

// matchesFilter checks if a meeting passes the given filter criteria.
func (r *Repository) matchesFilter(mtg *domain.Meeting, filter domain.ListFilter) bool {
	if filter.Since != nil && mtg.Datetime().Before(*filter.Since) {
		return false
	}
	if filter.Until != nil && mtg.Datetime().After(*filter.Until) {
		return false
	}
	if filter.Source != nil && mtg.Source() != *filter.Source {
		return false
	}
	if filter.Participant != nil {
		participantLower := strings.ToLower(*filter.Participant)
		found := false
		for _, p := range mtg.Participants() {
			if strings.Contains(strings.ToLower(p.Name()), participantLower) ||
				strings.Contains(strings.ToLower(p.Email()), participantLower) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if filter.Query != nil {
		queryLower := strings.ToLower(*filter.Query)
		if !strings.Contains(strings.ToLower(mtg.Title()), queryLower) {
			return false
		}
	}
	return true
}

// metaFor returns the metadata for a document ID, or nil if not found.
func (r *Repository) metaFor(id string) *CacheMeetingMeta {
	if meta, ok := r.state.State.MeetingsMetadata[id]; ok {
		return &meta
	}
	return nil
}

// transcriptContains checks if any segment text contains the query.
func transcriptContains(t CacheTranscript, queryLower string) bool {
	for _, seg := range t.Segments {
		if strings.Contains(strings.ToLower(seg.Text), queryLower) {
			return true
		}
	}
	return false
}

// mapError translates infrastructure errors to domain errors.
func (r *Repository) mapError(err error) error {
	if err == nil {
		return nil
	}
	// File not found maps to meeting not found at the repository level
	// since the entire data source is unavailable
	return err
}
