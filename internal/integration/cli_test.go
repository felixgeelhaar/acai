//go:build integration

package integration_test

import (
	"encoding/json"
	"strings"
	"testing"
)

// --- List Command Tests ---

func TestIntegration_CLI_ListMeetings_Table(t *testing.T) {
	env := newTestEnv(t, nil)

	err := env.runCLI("list", "meetings")
	assertNoError(t, err)

	output := env.outputString()
	assertContains(t, output, "ID")
	assertContains(t, output, "TITLE")
	assertContains(t, output, "Sprint Planning")
	assertContains(t, output, "Retrospective")
	assertContains(t, output, "1:1 with Manager")
}

func TestIntegration_CLI_ListMeetings_JSON(t *testing.T) {
	env := newTestEnv(t, nil)

	err := env.runCLI("list", "meetings", "--format", "json")
	assertNoError(t, err)

	output := env.outputString()
	var meetings []interface{}
	if err := json.Unmarshal([]byte(output), &meetings); err != nil {
		t.Fatalf("expected valid JSON array, got error: %v\noutput: %s", err, truncate(output, 300))
	}
	if len(meetings) != 3 {
		t.Errorf("expected 3 meetings in JSON, got %d", len(meetings))
	}
}

// --- Export Command Tests ---

func TestIntegration_CLI_ExportMeeting_Markdown(t *testing.T) {
	env := newTestEnv(t, nil)

	err := env.runCLI("export", "meeting", "m-1", "--format", "md")
	assertNoError(t, err)

	output := env.outputString()
	assertContains(t, output, "# Sprint Planning")
	assertContains(t, output, "Participants")
	assertContains(t, output, "Alice Johnson")
}

func TestIntegration_CLI_ExportMeeting_JSON(t *testing.T) {
	env := newTestEnv(t, nil)

	err := env.runCLI("export", "meeting", "m-1", "--format", "json")
	assertNoError(t, err)

	output := env.outputString()
	assertContains(t, output, "m-1")
	assertContains(t, output, "Sprint Planning")
}

func TestIntegration_CLI_ExportMeeting_DefaultFormat(t *testing.T) {
	env := newTestEnv(t, nil)

	// Default format is "table" which maps to markdown for export
	err := env.runCLI("export", "meeting", "m-1")
	assertNoError(t, err)

	output := env.outputString()
	assertContains(t, output, "# Sprint Planning")
}

func TestIntegration_CLI_ExportMeeting_NotFound(t *testing.T) {
	env := newTestEnv(t, nil)

	err := env.runCLI("export", "meeting", "nonexistent")
	assertError(t, err)
}

// --- Sync Command Tests ---

func TestIntegration_CLI_Sync(t *testing.T) {
	env := newTestEnv(t, nil)

	err := env.runCLI("sync")
	assertNoError(t, err)

	output := env.outputString()
	assertContains(t, output, "Synced")
	assertContains(t, output, "meeting event(s)")
}

func TestIntegration_CLI_Sync_WithSince(t *testing.T) {
	env := newTestEnv(t, nil)

	err := env.runCLI("sync", "--since", "2025-01-01")
	assertNoError(t, err)

	output := env.outputString()
	assertContains(t, output, "Synced")
}

func TestIntegration_CLI_Sync_WithSinceRFC3339(t *testing.T) {
	env := newTestEnv(t, nil)

	err := env.runCLI("sync", "--since", "2025-01-01T00:00:00Z")
	assertNoError(t, err)

	output := env.outputString()
	assertContains(t, output, "Synced")
}

// --- Note Command Tests ---

func TestIntegration_CLI_NoteAdd_ThenList(t *testing.T) {
	env := newTestEnv(t, nil)

	// Add a note
	err := env.runCLI("note", "add", "m-1", "CLI integration test note")
	assertNoError(t, err)

	output := env.outputString()
	assertContains(t, output, "Note")
	assertContains(t, output, "added to meeting m-1")

	// List notes
	env.resetOutput()
	err = env.runCLI("note", "list", "m-1")
	assertNoError(t, err)

	output = env.outputString()
	assertContains(t, output, "CLI integration test note")
}

func TestIntegration_CLI_NoteAdd_ThenDelete(t *testing.T) {
	env := newTestEnv(t, nil)

	// Add a note via CLI — output is "Note <id> added to meeting m-1"
	err := env.runCLI("note", "add", "m-1", "To be deleted via CLI")
	assertNoError(t, err)

	output := env.outputString()
	assertContains(t, output, "added to meeting m-1")

	// Extract the note ID from the output: "Note note-XXXX added to meeting m-1"
	parts := strings.Fields(output)
	if len(parts) < 2 {
		t.Fatalf("unexpected output format: %s", output)
	}
	noteID := parts[1] // "note-XXXX"

	// Delete the note
	env.resetOutput()
	err = env.runCLI("note", "delete", noteID)
	assertNoError(t, err)

	output = env.outputString()
	assertContains(t, output, "deleted")

	// Verify it's gone by listing
	env.resetOutput()
	err = env.runCLI("note", "list", "m-1")
	assertNoError(t, err)

	output = env.outputString()
	assertNotContains(t, output, "To be deleted via CLI")
}

func TestIntegration_CLI_NoteAdd_WithAuthor(t *testing.T) {
	env := newTestEnv(t, nil)

	err := env.runCLI("note", "add", "--author", "custom-agent", "m-1", "Note with custom author")
	assertNoError(t, err)

	output := env.outputString()
	assertContains(t, output, "added to meeting m-1")

	// List and verify author
	env.resetOutput()
	err = env.runCLI("note", "list", "m-1")
	assertNoError(t, err)

	output = env.outputString()
	assertContains(t, output, "custom-agent")
}

func TestIntegration_CLI_NoteList_JSON(t *testing.T) {
	env := newTestEnv(t, nil)

	// Add a note first
	err := env.runCLI("note", "add", "m-2", "JSON format test note")
	assertNoError(t, err)

	// List in JSON format
	env.resetOutput()
	err = env.runCLI("note", "list", "m-2", "--format", "json")
	assertNoError(t, err)

	output := env.outputString()
	var notes []interface{}
	if err := json.Unmarshal([]byte(output), &notes); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}
	if len(notes) != 1 {
		t.Errorf("expected 1 note in JSON, got %d", len(notes))
	}
}

// --- Version Command Tests ---

func TestIntegration_CLI_Version(t *testing.T) {
	env := newTestEnv(t, nil)

	err := env.runCLI("version")
	assertNoError(t, err)

	output := env.outputString()
	assertContains(t, output, "acai")
}

// --- Error Cases ---

func TestIntegration_CLI_ListMeetings_WithLimit(t *testing.T) {
	env := newTestEnv(t, nil)

	err := env.runCLI("list", "meetings", "--limit", "1")
	assertNoError(t, err)

	output := env.outputString()
	// At minimum, the table headers should be present
	assertContains(t, output, "ID")
	assertContains(t, output, "TITLE")
}

func TestIntegration_CLI_ExportMeeting_MissingArg(t *testing.T) {
	env := newTestEnv(t, nil)

	err := env.runCLI("export", "meeting")
	assertError(t, err) // requires exactly 1 arg
}

func TestIntegration_CLI_NoteAdd_MeetingNotFound(t *testing.T) {
	env := newTestEnv(t, nil)

	err := env.runCLI("note", "add", "nonexistent-meeting", "This should fail")
	assertError(t, err)
}

func TestIntegration_CLI_Sync_InvalidSince(t *testing.T) {
	env := newTestEnv(t, nil)

	err := env.runCLI("sync", "--since", "not-a-date")
	assertError(t, err)
}
