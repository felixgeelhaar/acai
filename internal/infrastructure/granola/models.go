// Package granola implements the anti-corruption layer (DDD) between
// the Granola public API and our domain model. External API DTOs are
// defined here and mapped to domain types via the mapper.
package granola

import "time"

// --- Public API Response DTOs ---
// These types mirror the Granola public API (public-api.granola.ai) response shapes.
// They are NEVER exposed outside the infrastructure layer.

type NoteListResponse struct {
	Notes   []NoteListItem `json:"notes"`
	HasMore bool           `json:"hasMore"`
	Cursor  string         `json:"cursor"`
}

type NoteListItem struct {
	ID        string    `json:"id"`
	Object    string    `json:"object"`
	Title     string    `json:"title"`
	Owner     UserDTO   `json:"owner"`
	CreatedAt time.Time `json:"created_at"`
}

type NoteDetailResponse struct {
	ID               string              `json:"id"`
	Object           string              `json:"object"`
	Title            string              `json:"title"`
	Owner            UserDTO             `json:"owner"`
	CreatedAt        time.Time           `json:"created_at"`
	CalendarEvent    *CalendarEventDTO   `json:"calendar_event,omitempty"`
	Attendees        []UserDTO           `json:"attendees,omitempty"`
	FolderMembership []FolderDTO         `json:"folder_membership,omitempty"`
	SummaryText      string              `json:"summary_text,omitempty"`
	SummaryMarkdown  *string             `json:"summary_markdown,omitempty"`
	Transcript       []TranscriptItemDTO `json:"transcript,omitempty"`
}

type UserDTO struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CalendarEventDTO struct {
	Title     string    `json:"title"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type TranscriptItemDTO struct {
	Speaker   string    `json:"speaker"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

type FolderDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
