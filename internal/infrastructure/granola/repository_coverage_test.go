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

func TestRepository_SearchTranscripts(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(granola.NoteListResponse{
			Notes: []granola.NoteListItem{
				{ID: "m-1", Title: "Meeting with keyword", CreatedAt: now},
			},
		})
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "token")
	repo := granola.NewRepository(client)

	meetings, err := repo.SearchTranscripts(context.Background(), "keyword", domain.ListFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(meetings) != 1 {
		t.Errorf("got %d meetings, want 1", len(meetings))
	}
}

func TestClient_SetToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer new-token" {
			t.Error("expected updated token in header")
		}
		_ = json.NewEncoder(w).Encode(granola.NoteListResponse{})
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "old-token")
	client.SetToken("new-token")

	_, err := client.ListNotes(context.Background(), nil, "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_NewClient_NilHTTPClient(t *testing.T) {
	client := granola.NewClient("http://localhost", nil, "token")
	if client == nil {
		t.Fatal("client should not be nil")
	}
}

func TestClient_ListNotes_WithCreatedAfter(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("created_after") == "" {
			t.Error("expected created_after query param")
		}
		_ = json.NewEncoder(w).Encode(granola.NoteListResponse{
			Notes: []granola.NoteListItem{
				{ID: "m-1", Title: "Recent", CreatedAt: now},
			},
		})
	}))
	defer server.Close()

	client := granola.NewClient(server.URL, server.Client(), "token")
	since := now.Add(-24 * time.Hour)
	resp, err := client.ListNotes(context.Background(), &since, "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Notes) != 1 {
		t.Errorf("got %d notes", len(resp.Notes))
	}
}
