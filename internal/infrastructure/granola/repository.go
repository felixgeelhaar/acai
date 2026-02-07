package granola

import (
	"context"
	"errors"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

// Repository implements domain.Repository by delegating to the Granola API client.
// This is the adapter side of the Ports & Adapters (Hexagonal) architecture —
// it translates between infrastructure (HTTP API) and domain concepts.
type Repository struct {
	client *Client
}

func NewRepository(client *Client) *Repository {
	return &Repository{client: client}
}

func (r *Repository) FindByID(ctx context.Context, id domain.MeetingID) (*domain.Meeting, error) {
	dto, err := r.client.GetDocument(ctx, string(id))
	if err != nil {
		return nil, r.mapError(err)
	}

	return mapDocumentToDomain(*dto)
}

func (r *Repository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Meeting, error) {
	resp, err := r.client.GetDocuments(ctx, filter.Since, filter.Limit, filter.Offset)
	if err != nil {
		return nil, r.mapError(err)
	}

	meetings := make([]*domain.Meeting, 0, len(resp.Documents))
	for _, dto := range resp.Documents {
		mtg, err := mapDocumentToDomain(dto)
		if err != nil {
			continue
		}
		meetings = append(meetings, mtg)
	}

	return meetings, nil
}

func (r *Repository) GetTranscript(ctx context.Context, id domain.MeetingID) (*domain.Transcript, error) {
	resp, err := r.client.GetTranscript(ctx, string(id))
	if err != nil {
		return nil, r.mapError(err)
	}

	return mapTranscriptToDomain(id, *resp), nil
}

func (r *Repository) SearchTranscripts(ctx context.Context, query string, filter domain.ListFilter) ([]*domain.Meeting, error) {
	// Granola API doesn't have a dedicated search endpoint.
	// For now, fetch all and filter client-side. This will be improved
	// when Granola adds search capabilities to their API.
	return r.List(ctx, filter)
}

func (r *Repository) GetActionItems(ctx context.Context, id domain.MeetingID) ([]*domain.ActionItem, error) {
	dto, err := r.client.GetDocument(ctx, string(id))
	if err != nil {
		return nil, r.mapError(err)
	}

	items := make([]*domain.ActionItem, 0, len(dto.ActionItems))
	for _, ai := range dto.ActionItems {
		item, err := mapActionItemToDomain(id, ai)
		if err != nil {
			continue
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *Repository) Sync(ctx context.Context, since *time.Time) ([]domain.DomainEvent, error) {
	resp, err := r.client.GetDocuments(ctx, since, 0, 0)
	if err != nil {
		return nil, r.mapError(err)
	}

	events := make([]domain.DomainEvent, 0, len(resp.Documents))
	for _, dto := range resp.Documents {
		events = append(events, domain.NewMeetingCreatedEvent(
			domain.MeetingID(dto.ID),
			dto.Title,
			dto.CreatedAt,
		))
	}

	return events, nil
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
