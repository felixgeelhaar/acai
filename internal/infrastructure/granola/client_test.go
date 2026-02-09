package granola_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/felixgeelhaar/acai/internal/infrastructure/granola"
)

func TestClient_ListNotes(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/notes" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing or wrong auth header")
		}

		resp := granola.NoteListResponse{
			Notes: []granola.NoteListItem{
				{
					ID:        "m-1",
					Title:     "Sprint Planning",
					Owner:     granola.UserDTO{Name: "Alice", Email: "alice@test.com"},
					CreatedAt: now,
				},
			},
			HasMore: false,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "test-token")
	resp, err := client.ListNotes(context.Background(), nil, "", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Notes) != 1 {
		t.Fatalf("got %d notes, want 1", len(resp.Notes))
	}
	if resp.Notes[0].ID != "m-1" {
		t.Errorf("got id %q", resp.Notes[0].ID)
	}
}

func TestClient_ListNotes_WithCursor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cursor := r.URL.Query().Get("cursor")
		if cursor == "page2" {
			_ = json.NewEncoder(w).Encode(granola.NoteListResponse{
				Notes:   []granola.NoteListItem{{ID: "m-2", Title: "Page 2"}},
				HasMore: false,
			})
			return
		}
		_ = json.NewEncoder(w).Encode(granola.NoteListResponse{
			Notes:   []granola.NoteListItem{{ID: "m-1", Title: "Page 1"}},
			HasMore: true,
			Cursor:  "page2",
		})
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "token")

	resp, err := client.ListNotes(context.Background(), nil, "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.HasMore {
		t.Error("expected HasMore=true")
	}
	if resp.Cursor != "page2" {
		t.Errorf("got cursor %q", resp.Cursor)
	}
}

func TestClient_GetNote(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/notes/m-1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		resp := granola.NoteDetailResponse{
			ID:          "m-1",
			Title:       "Sprint Planning",
			Owner:       granola.UserDTO{Name: "Alice", Email: "alice@test.com"},
			CreatedAt:   now,
			SummaryText: "We planned the sprint.",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "token")
	resp, err := client.GetNote(context.Background(), "m-1", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "m-1" {
		t.Errorf("got id %q", resp.ID)
	}
	if resp.SummaryText != "We planned the sprint." {
		t.Errorf("got summary %q", resp.SummaryText)
	}
}

func TestClient_GetNote_WithTranscript(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("include") != "transcript" {
			t.Error("expected include=transcript query param")
		}
		resp := granola.NoteDetailResponse{
			ID:        "m-1",
			Title:     "Sprint Planning",
			Owner:     granola.UserDTO{Name: "Alice", Email: "alice@test.com"},
			CreatedAt: now,
			Transcript: []granola.TranscriptItemDTO{
				{Speaker: "Alice", Text: "Hello", Timestamp: now},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "token")
	resp, err := client.GetNote(context.Background(), "m-1", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Transcript) != 1 {
		t.Errorf("got %d transcript items", len(resp.Transcript))
	}
}

func TestClient_GetNote_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "test-token")
	_, err := client.GetNote(context.Background(), "nonexistent", false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "bad-token")
	_, err := client.ListNotes(context.Background(), nil, "", 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestClient_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "test-token")
	_, err := client.ListNotes(context.Background(), nil, "", 0)
	if err == nil {
		t.Fatal("expected error")
	}
}
