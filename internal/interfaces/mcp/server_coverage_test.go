package mcp_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
	mcpiface "github.com/felixgeelhaar/granola-mcp/internal/interfaces/mcp"
)

func TestServer_NameAndVersion(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)

	if srv.Name() != "granola-mcp" {
		t.Errorf("got name %q", srv.Name())
	}
	if srv.Version() != "test" {
		t.Errorf("got version %q", srv.Version())
	}
}

func TestServer_HandleSearchTranscripts(t *testing.T) {
	repo := newMockRepo()
	repo.addMeeting(mustMeeting(t, "m-1", "Sprint Planning"))

	srv := newTestServer(repo)

	results, err := srv.HandleSearchTranscripts(context.Background(), mcpiface.SearchTranscriptsToolInput{
		Query: "sprint",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results", len(results))
	}
}

func TestServer_HandleListMeetings_WithFilters(t *testing.T) {
	repo := newMockRepo()
	repo.addMeeting(mustMeeting(t, "m-1", "Meeting"))

	srv := newTestServer(repo)
	since := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	until := time.Now().UTC().Format(time.RFC3339)
	limit := 5
	offset := 0
	source := "zoom"

	results, err := srv.HandleListMeetings(context.Background(), mcpiface.ListMeetingsToolInput{
		Since:  &since,
		Until:  &until,
		Limit:  &limit,
		Offset: &offset,
		Source: &source,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results", len(results))
	}
}

func TestServer_HandleListMeetings_InvalidSinceDate(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)

	bad := "not-a-date"
	_, err := srv.HandleListMeetings(context.Background(), mcpiface.ListMeetingsToolInput{
		Since: &bad,
	})
	if err == nil {
		t.Fatal("expected error for invalid since date")
	}
}

func TestServer_HandleListMeetings_InvalidUntilDate(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)

	bad := "not-a-date"
	_, err := srv.HandleListMeetings(context.Background(), mcpiface.ListMeetingsToolInput{
		Until: &bad,
	})
	if err == nil {
		t.Fatal("expected error for invalid until date")
	}
}

func TestServer_HandleGetActionItems_WithDueDate(t *testing.T) {
	repo := newMockRepo()
	due := time.Now().Add(48 * time.Hour).UTC()
	item, _ := domain.NewActionItem("ai-1", "m-1", "Alice", "Write report", &due)
	repo.addActionItems("m-1", []*domain.ActionItem{item})

	srv := newTestServer(repo)
	results, err := srv.HandleGetActionItems(context.Background(), mcpiface.GetActionItemsToolInput{
		MeetingID: "m-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].DueDate == nil {
		t.Error("expected due date")
	}
}

func TestServer_HandleToolJSON_GetMeeting(t *testing.T) {
	repo := newMockRepo()
	repo.addMeeting(mustMeeting(t, "m-1", "Test"))

	srv := newTestServer(repo)

	raw, err := srv.HandleToolJSON(context.Background(), "get_meeting", json.RawMessage(`{"id":"m-1"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result mcpiface.MeetingDetailResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if result.ID != "m-1" {
		t.Errorf("got id %q", result.ID)
	}
}

func TestServer_HandleToolJSON_GetTranscript(t *testing.T) {
	repo := newMockRepo()
	transcript := domain.NewTranscript("m-1", []domain.Utterance{
		domain.NewUtterance("Alice", "Hello", time.Now().UTC(), 0.9),
	})
	repo.addTranscript("m-1", &transcript)

	srv := newTestServer(repo)

	raw, err := srv.HandleToolJSON(context.Background(), "get_transcript", json.RawMessage(`{"meeting_id":"m-1"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result mcpiface.TranscriptResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(result.Utterances) != 1 {
		t.Errorf("got %d utterances", len(result.Utterances))
	}
}

func TestServer_HandleToolJSON_SearchTranscripts(t *testing.T) {
	repo := newMockRepo()
	repo.addMeeting(mustMeeting(t, "m-1", "Meeting"))

	srv := newTestServer(repo)

	raw, err := srv.HandleToolJSON(context.Background(), "search_transcripts", json.RawMessage(`{"query":"meeting"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var results []mcpiface.MeetingResult
	if err := json.Unmarshal(raw, &results); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
}

func TestServer_HandleToolJSON_GetActionItems(t *testing.T) {
	repo := newMockRepo()
	item, _ := domain.NewActionItem("ai-1", "m-1", "Alice", "Write", nil)
	repo.addActionItems("m-1", []*domain.ActionItem{item})

	srv := newTestServer(repo)

	raw, err := srv.HandleToolJSON(context.Background(), "get_action_items", json.RawMessage(`{"meeting_id":"m-1"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var results []mcpiface.ActionItemResult
	if err := json.Unmarshal(raw, &results); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results", len(results))
	}
}

func TestServer_HandleToolJSON_InvalidInput(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)

	_, err := srv.HandleToolJSON(context.Background(), "get_meeting", json.RawMessage(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestServer_Inner(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)

	if srv.Inner() == nil {
		t.Error("Inner() should not return nil")
	}
}

func TestServer_HandleSearchTranscripts_WithSince(t *testing.T) {
	repo := newMockRepo()
	repo.addMeeting(mustMeeting(t, "m-1", "Planning"))

	srv := newTestServer(repo)
	since := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	limit := 10

	results, err := srv.HandleSearchTranscripts(context.Background(), mcpiface.SearchTranscriptsToolInput{
		Query: "planning",
		Since: &since,
		Limit: &limit,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results", len(results))
	}
}

func TestServer_HandleSearchTranscripts_InvalidSince(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)

	bad := "not-a-date"
	_, err := srv.HandleSearchTranscripts(context.Background(), mcpiface.SearchTranscriptsToolInput{
		Query: "test",
		Since: &bad,
	})
	if err == nil {
		t.Fatal("expected error for invalid since date")
	}
}

func TestServer_HandleGetTranscript_NotFound(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)

	_, err := srv.HandleGetTranscript(context.Background(), mcpiface.GetTranscriptToolInput{MeetingID: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for missing transcript")
	}
}

func TestServer_HandleToolJSON_InvalidJSON_AllTools(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)

	tools := []string{"list_meetings", "get_meeting", "get_transcript", "search_transcripts", "get_action_items", "meeting_stats"}
	for _, tool := range tools {
		_, err := srv.HandleToolJSON(context.Background(), tool, json.RawMessage(`{invalid`))
		if err == nil {
			t.Errorf("expected error for invalid JSON on tool %q", tool)
		}
	}
}

func TestServer_HandleGetMeeting_WithParticipants(t *testing.T) {
	repo := newMockRepo()
	m, _ := domain.New("m-1", "Sprint Planning", time.Now().UTC(), domain.SourceZoom, []domain.Participant{
		domain.NewParticipant("Alice", "alice@test.com", domain.RoleHost),
		domain.NewParticipant("Bob", "bob@test.com", domain.RoleAttendee),
	})
	m.ClearDomainEvents()
	repo.addMeeting(m)

	srv := newTestServer(repo)

	result, err := srv.HandleGetMeeting(context.Background(), mcpiface.GetMeetingToolInput{ID: "m-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Participants) != 2 {
		t.Errorf("got %d participants", len(result.Participants))
	}
	if result.Participants[0].Role != "host" {
		t.Errorf("got role %q", result.Participants[0].Role)
	}
}
