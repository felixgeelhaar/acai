package granola_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
	"github.com/felixgeelhaar/acai/internal/infrastructure/granola"
)

func TestRepository_FindByID(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(granola.NoteDetailResponse{
			ID:    "m-1",
			Title: "Sprint Planning",
			Owner: granola.UserDTO{Name: "Alice", Email: "alice@test.com"},
			CreatedAt: now,
			Attendees: []granola.UserDTO{
				{Name: "Alice", Email: "alice@test.com"},
			},
		})
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "token")
	repo := granola.NewRepository(client)

	mtg, err := repo.FindByID(context.Background(), "m-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mtg.ID() != "m-1" {
		t.Errorf("got id %q", mtg.ID())
	}
	if mtg.Title() != "Sprint Planning" {
		t.Errorf("got title %q", mtg.Title())
	}
	// Owner deduped from attendees
	if len(mtg.Participants()) != 1 {
		t.Errorf("got %d participants", len(mtg.Participants()))
	}
}

func TestRepository_FindByID_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "token")
	repo := granola.NewRepository(client)

	_, err := repo.FindByID(context.Background(), "nonexistent")
	if err != domain.ErrMeetingNotFound {
		t.Errorf("got error %v, want %v", err, domain.ErrMeetingNotFound)
	}
}

func TestRepository_List(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(granola.NoteListResponse{
			Notes: []granola.NoteListItem{
				{ID: "m-1", Title: "Meeting 1", CreatedAt: now},
				{ID: "m-2", Title: "Meeting 2", CreatedAt: now},
			},
		})
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "token")
	repo := granola.NewRepository(client)

	meetings, err := repo.List(context.Background(), domain.ListFilter{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(meetings) != 2 {
		t.Errorf("got %d meetings, want 2", len(meetings))
	}
}

func TestRepository_GetTranscript(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(granola.NoteDetailResponse{
			ID:    "m-1",
			Title: "Meeting",
			Owner: granola.UserDTO{Name: "Alice", Email: "alice@test.com"},
			CreatedAt: now,
			Transcript: []granola.TranscriptItemDTO{
				{Speaker: "Alice", Text: "Hello", Timestamp: now},
			},
		})
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "token")
	repo := granola.NewRepository(client)

	transcript, err := repo.GetTranscript(context.Background(), "m-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(transcript.Utterances()) != 1 {
		t.Errorf("got %d utterances", len(transcript.Utterances()))
	}
}

func TestRepository_GetTranscript_NotReady(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(granola.NoteDetailResponse{
			ID:        "m-1",
			Title:     "Meeting",
			Owner:     granola.UserDTO{Name: "Alice", Email: "alice@test.com"},
			CreatedAt: now,
			// No transcript
		})
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "token")
	repo := granola.NewRepository(client)

	_, err := repo.GetTranscript(context.Background(), "m-1")
	if err != domain.ErrTranscriptNotReady {
		t.Errorf("got error %v, want %v", err, domain.ErrTranscriptNotReady)
	}
}

func TestRepository_GetActionItems_ReturnsEmpty(t *testing.T) {
	// Public API doesn't expose action items — should always return empty
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make any HTTP call")
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "token")
	repo := granola.NewRepository(client)

	items, err := repo.GetActionItems(context.Background(), "m-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}

func TestRepository_Sync(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(granola.NoteListResponse{
			Notes: []granola.NoteListItem{
				{ID: "m-new", Title: "New Meeting", CreatedAt: now},
			},
			HasMore: false,
		})
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "token")
	repo := granola.NewRepository(client)

	since := now.Add(-1 * time.Hour)
	events, err := repo.Sync(context.Background(), &since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("got %d events, want 1", len(events))
	}
	if events[0].EventName() != "meeting.created" {
		t.Errorf("got event %q", events[0].EventName())
	}
}

func TestRepository_Sync_CursorPagination(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		cursor := r.URL.Query().Get("cursor")
		if cursor == "page2" {
			_ = json.NewEncoder(w).Encode(granola.NoteListResponse{
				Notes: []granola.NoteListItem{
					{ID: "m-2", Title: "Meeting 2", CreatedAt: now},
				},
				HasMore: false,
			})
			return
		}
		_ = json.NewEncoder(w).Encode(granola.NoteListResponse{
			Notes: []granola.NoteListItem{
				{ID: "m-1", Title: "Meeting 1", CreatedAt: now},
			},
			HasMore: true,
			Cursor:  "page2",
		})
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "token")
	repo := granola.NewRepository(client)

	events, err := repo.Sync(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("got %d events, want 2", len(events))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestRepository_FindByID_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "bad")
	repo := granola.NewRepository(client)

	_, err := repo.FindByID(context.Background(), "m-1")
	if err != domain.ErrAccessDenied {
		t.Errorf("got error %v, want %v", err, domain.ErrAccessDenied)
	}
}
