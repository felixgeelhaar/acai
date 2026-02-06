// Package granola implements the anti-corruption layer (DDD) between
// the Granola REST API and our domain model. External API DTOs are
// defined here and mapped to domain types via the mapper.
package granola

import "time"

// --- API Response DTOs ---
// These types mirror the Granola REST API response shapes.
// They are NEVER exposed outside the infrastructure layer.

type DocumentListResponse struct {
	Documents []DocumentDTO `json:"documents"`
}

type DocumentDTO struct {
	ID           string           `json:"id"`
	Title        string           `json:"title"`
	CreatedAt    time.Time        `json:"created_at"`
	Source       string           `json:"source"`
	Participants []ParticipantDTO `json:"participants"`
	Summary      *SummaryDTO      `json:"summary,omitempty"`
	ActionItems  []ActionItemDTO  `json:"action_items,omitempty"`
	Metadata     *MetadataDTO     `json:"metadata,omitempty"`
}

type ParticipantDTO struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type TranscriptResponse struct {
	MeetingID  string         `json:"meeting_id"`
	Utterances []UtteranceDTO `json:"utterances"`
}

type UtteranceDTO struct {
	Speaker    string    `json:"speaker"`
	Text       string    `json:"text"`
	Timestamp  time.Time `json:"timestamp"`
	Confidence float64   `json:"confidence"`
}

type SummaryDTO struct {
	Content string `json:"content"`
	Type    string `json:"type"`
}

type ActionItemDTO struct {
	ID      string     `json:"id"`
	Owner   string     `json:"owner"`
	Text    string     `json:"text"`
	DueDate *time.Time `json:"due_date,omitempty"`
	Done    bool       `json:"done"`
}

type MetadataDTO struct {
	Tags         []string          `json:"tags,omitempty"`
	Links        []string          `json:"links,omitempty"`
	ExternalRefs map[string]string `json:"external_refs,omitempty"`
}

type WorkspaceDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type WorkspaceListResponse struct {
	Workspaces []WorkspaceDTO `json:"workspaces"`
}
