//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
	"github.com/felixgeelhaar/acai/internal/infrastructure/granola"
	mcpiface "github.com/felixgeelhaar/acai/internal/interfaces/mcp"
)

// --- Read Path Tests ---

func TestIntegration_MCP_ListMeetings(t *testing.T) {
	env := newTestEnv(t, nil)

	result, err := env.MCPServer.HandleListMeetings(context.Background(), mcpiface.ListMeetingsToolInput{})
	assertNoError(t, err)

	if len(result) != 3 {
		t.Fatalf("expected 3 meetings, got %d", len(result))
	}

	titles := map[string]bool{}
	for _, m := range result {
		titles[m.Title] = true
	}
	for _, expected := range []string{"Sprint Planning", "Retrospective", "1:1 with Manager"} {
		if !titles[expected] {
			t.Errorf("expected meeting %q in results", expected)
		}
	}

	// Verify participants for the first meeting (owner mapped as host)
	for _, m := range result {
		if m.ID == "m-1" {
			if len(m.Participants) == 0 {
				t.Error("expected participants for m-1")
			}
			found := false
			for _, p := range m.Participants {
				if p.Name == "Alice Johnson" && p.Role == "host" {
					found = true
				}
			}
			if !found {
				t.Error("expected Alice Johnson as host participant for m-1")
			}
		}
	}
}

func TestIntegration_MCP_GetMeeting(t *testing.T) {
	env := newTestEnv(t, nil)

	result, err := env.MCPServer.HandleGetMeeting(context.Background(), mcpiface.GetMeetingToolInput{ID: "m-1"})
	assertNoError(t, err)

	if result.ID != "m-1" {
		t.Errorf("expected ID m-1, got %s", result.ID)
	}
	if result.Title != "Sprint Planning" {
		t.Errorf("expected title Sprint Planning, got %s", result.Title)
	}

	// Verify markdown summary is preferred over text
	if result.Summary == nil {
		t.Fatal("expected summary")
	}
	assertContains(t, result.Summary.Content, "## Sprint Planning")

	// Verify participants: owner (Alice) + 2 attendees (Bob, Carol) = 3
	if len(result.Participants) != 3 {
		t.Errorf("expected 3 participants, got %d", len(result.Participants))
	}

	roles := map[string]string{}
	for _, p := range result.Participants {
		roles[p.Name] = p.Role
	}
	if roles["Alice Johnson"] != "host" {
		t.Error("expected Alice Johnson as host")
	}
	if roles["Bob Smith"] != "attendee" {
		t.Error("expected Bob Smith as attendee")
	}
}

func TestIntegration_MCP_GetMeeting_NotFound(t *testing.T) {
	env := newTestEnv(t, nil)

	_, err := env.MCPServer.HandleGetMeeting(context.Background(), mcpiface.GetMeetingToolInput{ID: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent meeting")
	}
	if !errors.Is(err, domain.ErrMeetingNotFound) {
		t.Errorf("expected ErrMeetingNotFound, got: %v", err)
	}
}

func TestIntegration_MCP_GetTranscript(t *testing.T) {
	env := newTestEnv(t, nil)

	result, err := env.MCPServer.HandleGetTranscript(context.Background(), mcpiface.GetTranscriptToolInput{MeetingID: "m-1"})
	assertNoError(t, err)

	if result.MeetingID != "m-1" {
		t.Errorf("expected meeting_id m-1, got %s", result.MeetingID)
	}
	if len(result.Utterances) != 3 {
		t.Fatalf("expected 3 utterances, got %d", len(result.Utterances))
	}

	speakers := make([]string, len(result.Utterances))
	for i, u := range result.Utterances {
		speakers[i] = u.Speaker
	}
	if speakers[0] != "Alice Johnson" || speakers[1] != "Bob Smith" || speakers[2] != "Carol Davis" {
		t.Errorf("unexpected speakers: %v", speakers)
	}
}

func TestIntegration_MCP_GetTranscript_Empty(t *testing.T) {
	env := newTestEnv(t, nil)

	// m-3 has no transcript
	_, err := env.MCPServer.HandleGetTranscript(context.Background(), mcpiface.GetTranscriptToolInput{MeetingID: "m-3"})
	if err == nil {
		t.Fatal("expected error for meeting without transcript")
	}
	if !errors.Is(err, domain.ErrTranscriptNotReady) {
		t.Errorf("expected ErrTranscriptNotReady, got: %v", err)
	}
}

func TestIntegration_MCP_GetActionItems_Empty(t *testing.T) {
	env := newTestEnv(t, nil)

	result, err := env.MCPServer.HandleGetActionItems(context.Background(), mcpiface.GetActionItemsToolInput{MeetingID: "m-1"})
	assertNoError(t, err)

	// Public API returns empty action items (they're local-only)
	if len(result) != 0 {
		t.Errorf("expected 0 action items, got %d", len(result))
	}
}

func TestIntegration_MCP_MeetingStats(t *testing.T) {
	env := newTestEnv(t, nil)

	result, err := env.MCPServer.HandleMeetingStats(context.Background(), mcpiface.MeetingStatsToolInput{})
	assertNoError(t, err)

	if result.TotalMeetings != 3 {
		t.Errorf("expected 3 total meetings, got %d", result.TotalMeetings)
	}

	// All meetings have source "other" (public API doesn't expose source)
	if len(result.PlatformDistribution) == 0 {
		t.Error("expected platform distribution entries")
	}

	// Summary coverage: List uses mapNoteListItemToDomain which doesn't include summaries,
	// so stats sees all meetings without summaries (summaries only come via FindByID).
	if result.SummaryCoverage.WithSummary != 0 {
		t.Errorf("expected 0 with summary (list doesn't include summaries), got %d", result.SummaryCoverage.WithSummary)
	}
	if result.SummaryCoverage.WithoutSummary != 3 {
		t.Errorf("expected 3 without summary, got %d", result.SummaryCoverage.WithoutSummary)
	}
}

func TestIntegration_MCP_HandleToolJSON_Dispatch(t *testing.T) {
	env := newTestEnv(t, nil)

	tests := []struct {
		tool  string
		input interface{}
	}{
		{"list_meetings", mcpiface.ListMeetingsToolInput{}},
		{"get_meeting", mcpiface.GetMeetingToolInput{ID: "m-1"}},
		{"get_transcript", mcpiface.GetTranscriptToolInput{MeetingID: "m-1"}},
		{"get_action_items", mcpiface.GetActionItemsToolInput{MeetingID: "m-1"}},
		{"meeting_stats", mcpiface.MeetingStatsToolInput{}},
		{"list_notes", mcpiface.ListNotesToolInput{MeetingID: "m-1"}},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			result := env.mustCallToolJSON(t, tt.tool, tt.input)
			if len(result) == 0 {
				t.Error("expected non-empty result")
			}
		})
	}

	// Test unknown tool
	_, err := env.callToolJSON(t, "nonexistent_tool", struct{}{})
	if err == nil {
		t.Error("expected error for unknown tool")
	}
	assertContains(t, err.Error(), "unknown tool")
}

// --- Write Path Tests ---

func TestIntegration_MCP_AddNote_ThenList(t *testing.T) {
	env := newTestEnv(t, nil)

	// Add a note
	addResult, err := env.MCPServer.HandleAddNote(context.Background(), mcpiface.AddNoteToolInput{
		MeetingID: "m-1",
		Author:    "test-agent",
		Content:   "Important follow-up item",
	})
	assertNoError(t, err)

	if addResult.MeetingID != "m-1" {
		t.Errorf("expected meeting_id m-1, got %s", addResult.MeetingID)
	}
	if addResult.Author != "test-agent" {
		t.Errorf("expected author test-agent, got %s", addResult.Author)
	}
	if addResult.Content != "Important follow-up item" {
		t.Errorf("expected content 'Important follow-up item', got %s", addResult.Content)
	}

	// List notes for the meeting
	listResult, err := env.MCPServer.HandleListNotes(context.Background(), mcpiface.ListNotesToolInput{
		MeetingID: "m-1",
	})
	assertNoError(t, err)

	if len(listResult) != 1 {
		t.Fatalf("expected 1 note, got %d", len(listResult))
	}
	if listResult[0].Content != "Important follow-up item" {
		t.Errorf("expected note content 'Important follow-up item', got %s", listResult[0].Content)
	}
}

func TestIntegration_MCP_AddNote_MeetingNotFound(t *testing.T) {
	env := newTestEnv(t, nil)

	_, err := env.MCPServer.HandleAddNote(context.Background(), mcpiface.AddNoteToolInput{
		MeetingID: "nonexistent",
		Author:    "test-agent",
		Content:   "This should fail",
	})
	assertError(t, err)
}

func TestIntegration_MCP_DeleteNote(t *testing.T) {
	env := newTestEnv(t, nil)

	// Add a note
	addResult, err := env.MCPServer.HandleAddNote(context.Background(), mcpiface.AddNoteToolInput{
		MeetingID: "m-1",
		Author:    "test-agent",
		Content:   "To be deleted",
	})
	assertNoError(t, err)

	// Delete it
	_, err = env.MCPServer.HandleDeleteNote(context.Background(), mcpiface.DeleteNoteToolInput{
		NoteID: addResult.ID,
	})
	assertNoError(t, err)

	// List should be empty
	listResult, err := env.MCPServer.HandleListNotes(context.Background(), mcpiface.ListNotesToolInput{
		MeetingID: "m-1",
	})
	assertNoError(t, err)
	if len(listResult) != 0 {
		t.Errorf("expected 0 notes after delete, got %d", len(listResult))
	}
}

func TestIntegration_MCP_AddNote_ViaToolJSON(t *testing.T) {
	env := newTestEnv(t, nil)

	// Add note via HandleToolJSON
	addInput := mcpiface.AddNoteToolInput{
		MeetingID: "m-1",
		Author:    "json-agent",
		Content:   "Note via tool JSON",
	}
	addRaw := env.mustCallToolJSON(t, "add_note", addInput)

	var addResult mcpiface.NoteResult
	unmarshalResult(t, addRaw, &addResult)

	if addResult.Content != "Note via tool JSON" {
		t.Errorf("expected content 'Note via tool JSON', got %s", addResult.Content)
	}

	// List via HandleToolJSON
	listRaw := env.mustCallToolJSON(t, "list_notes", mcpiface.ListNotesToolInput{MeetingID: "m-1"})

	var listResult []mcpiface.NoteResult
	unmarshalResult(t, listRaw, &listResult)

	if len(listResult) != 1 {
		t.Fatalf("expected 1 note, got %d", len(listResult))
	}

	// Delete via HandleToolJSON
	env.mustCallToolJSON(t, "delete_note", mcpiface.DeleteNoteToolInput{NoteID: addResult.ID})

	// Verify empty
	listRaw2 := env.mustCallToolJSON(t, "list_notes", mcpiface.ListNotesToolInput{MeetingID: "m-1"})
	var listResult2 []mcpiface.NoteResult
	unmarshalResult(t, listRaw2, &listResult2)
	if len(listResult2) != 0 {
		t.Errorf("expected 0 notes after delete, got %d", len(listResult2))
	}
}

func TestIntegration_MCP_SearchTranscripts(t *testing.T) {
	env := newTestEnv(t, nil)

	// SearchTranscripts falls back to List (no search API)
	result, err := env.MCPServer.HandleSearchTranscripts(context.Background(), mcpiface.SearchTranscriptsToolInput{
		Query: "sprint",
	})
	assertNoError(t, err)

	// Should return all meetings since search falls back to list
	if len(result) != 3 {
		t.Errorf("expected 3 meetings from search fallback, got %d", len(result))
	}
}

func TestIntegration_MCP_GetMeeting_SummaryPreference(t *testing.T) {
	env := newTestEnv(t, nil)

	// m-2 has only text summary (no markdown)
	result, err := env.MCPServer.HandleGetMeeting(context.Background(), mcpiface.GetMeetingToolInput{ID: "m-2"})
	assertNoError(t, err)

	if result.Summary == nil {
		t.Fatal("expected summary for m-2")
	}
	assertContains(t, result.Summary.Content, "Retro summary")

	// m-3 has no summary
	result3, err := env.MCPServer.HandleGetMeeting(context.Background(), mcpiface.GetMeetingToolInput{ID: "m-3"})
	assertNoError(t, err)
	if result3.Summary != nil {
		t.Error("expected no summary for m-3")
	}
}

func TestIntegration_MCP_ExportEmbeddings(t *testing.T) {
	env := newTestEnv(t, nil)

	raw := env.mustCallToolJSON(t, "export_embeddings", mcpiface.ExportEmbeddingsToolInput{
		MeetingIDs: []string{"m-1"},
		Strategy:   "speaker_turn",
	})

	var result mcpiface.ExportEmbeddingsResult
	unmarshalResult(t, raw, &result)

	if result.ChunkCount == 0 {
		t.Error("expected at least one chunk")
	}
	if result.Format != "jsonl" {
		t.Errorf("expected format jsonl, got %s", result.Format)
	}
	if result.Content == "" {
		t.Error("expected non-empty content")
	}
}

func TestIntegration_MCP_GetMeeting_ParticipantDedup(t *testing.T) {
	// Create fixtures where owner is also in attendees (should be deduped)
	fixtures := defaultFixtures()
	// Add Alice (owner) as attendee too
	fixtures.NoteDetails["m-1"] = func() granola.NoteDetailResponse {
		d := fixtures.NoteDetails["m-1"]
		d.Attendees = append(d.Attendees, granola.UserDTO{Name: "Alice Johnson", Email: "alice@example.com"})
		return d
	}()

	env := newTestEnv(t, fixtures)

	result, err := env.MCPServer.HandleGetMeeting(context.Background(), mcpiface.GetMeetingToolInput{ID: "m-1"})
	assertNoError(t, err)

	// Alice should appear only once (as host), not duplicated
	aliceCount := 0
	for _, p := range result.Participants {
		if p.Email == "alice@example.com" {
			aliceCount++
		}
	}
	if aliceCount != 1 {
		t.Errorf("expected Alice exactly once, found %d times", aliceCount)
	}
}

func TestIntegration_MCP_ListMeetings_WithLimit(t *testing.T) {
	env := newTestEnv(t, nil)

	limit := 2
	result, err := env.MCPServer.HandleListMeetings(context.Background(), mcpiface.ListMeetingsToolInput{
		Limit: &limit,
	})
	assertNoError(t, err)

	// The API returns all 3 (fake API doesn't enforce limit), but at least the request succeeds
	if len(result) == 0 {
		t.Error("expected at least 1 meeting")
	}
}

func TestIntegration_MCP_AuthHeader_Verified(t *testing.T) {
	fixtures := defaultFixtures()
	apiServer := fakeGranolaAPI(t, fixtures)
	t.Cleanup(apiServer.Close)

	// Use wrong token
	client := granola.NewClient(apiServer.URL, apiServer.Client(), "wrong-token")
	repo := granola.NewRepository(client)

	_, err := repo.FindByID(context.Background(), "m-1")
	if err == nil {
		t.Fatal("expected error with wrong token")
	}
	if !errors.Is(err, domain.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied, got: %v", err)
	}
}

// Store raw JSON to verify dispatch roundtrip
func TestIntegration_MCP_ToolJSON_InvalidInput(t *testing.T) {
	env := newTestEnv(t, nil)

	// Pass invalid JSON
	_, err := env.MCPServer.HandleToolJSON(context.Background(), "get_meeting", json.RawMessage(`{invalid`))
	assertError(t, err)
	assertContains(t, err.Error(), "invalid input")
}
