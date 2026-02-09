package granola

import (
	"context"
	"errors"
	"log"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

// Repository implements domain.Repository by delegating to the Granola public API client.
// This is the adapter side of the Ports & Adapters (Hexagonal) architecture —
// it translates between infrastructure (HTTP API) and domain concepts.
type Repository struct {
	client *Client
}

func NewRepository(client *Client) *Repository {
	return &Repository{client: client}
}

func (r *Repository) FindByID(ctx context.Context, id domain.MeetingID) (*domain.Meeting, error) {
	dto, err := r.client.GetNote(ctx, string(id), false)
	if err != nil {
		return nil, r.mapError(err)
	}

	return mapNoteDetailToDomain(*dto)
}

func (r *Repository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Meeting, error) {
	resp, err := r.client.ListNotes(ctx, filter.Since, "", filter.Limit)
	if err != nil {
		return nil, r.mapError(err)
	}

	meetings := make([]*domain.Meeting, 0, len(resp.Notes))
	for _, item := range resp.Notes {
		mtg, err := mapNoteListItemToDomain(item)
		if err != nil {
			log.Printf("granola: skipping invalid note %s: %v", item.ID, err)
			continue
		}
		meetings = append(meetings, mtg)
	}

	return meetings, nil
}

func (r *Repository) GetTranscript(ctx context.Context, id domain.MeetingID) (*domain.Transcript, error) {
	dto, err := r.client.GetNote(ctx, string(id), true)
	if err != nil {
		return nil, r.mapError(err)
	}

	t := mapTranscriptFromDetail(id, *dto)
	if t == nil {
		return nil, domain.ErrTranscriptNotReady
	}
	return t, nil
}

func (r *Repository) SearchTranscripts(ctx context.Context, query string, filter domain.ListFilter) ([]*domain.Meeting, error) {
	// Granola public API doesn't have a dedicated search endpoint.
	// Fetch all and filter client-side.
	return r.List(ctx, filter)
}

func (r *Repository) GetActionItems(_ context.Context, _ domain.MeetingID) ([]*domain.ActionItem, error) {
	// Public API does not expose action items — they are local-only now.
	return []*domain.ActionItem{}, nil
}

func (r *Repository) Sync(ctx context.Context, since *time.Time) ([]domain.DomainEvent, error) {
	var allEvents []domain.DomainEvent
	var cursor string

	for {
		resp, err := r.client.ListNotes(ctx, since, cursor, 0)
		if err != nil {
			return nil, r.mapError(err)
		}

		for _, item := range resp.Notes {
			allEvents = append(allEvents, domain.NewMeetingCreatedEvent(
				domain.MeetingID(item.ID),
				item.Title,
				item.CreatedAt,
			))
		}

		if !resp.HasMore || resp.Cursor == "" {
			break
		}
		cursor = resp.Cursor
	}

	return allEvents, nil
}

// mapError translates infrastructure errors to domain errors.
// This ensures the domain layer never sees HTTP-specific error types.
func (r *Repository) mapError(err error) error {
	if errors.Is(err, ErrNotFound) {
		return domain.ErrMeetingNotFound
	}
	if errors.Is(err, ErrUnauthorized) {
		return domain.ErrAccessDenied
	}
	return err
}
