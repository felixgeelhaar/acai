package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/felixgeelhaar/acai/internal/interfaces/cli"
)

func TestListMeetingsCmd_TableFormat(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"list", "meetings", "--format", "table"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := deps.Out.(*bytes.Buffer).String()
	if !strings.Contains(output, "ID") || !strings.Contains(output, "TITLE") {
		t.Errorf("expected table headers, got: %q", output)
	}
}

func TestListMeetingsCmd_JSONFormat(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"list", "meetings", "--format", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := deps.Out.(*bytes.Buffer).String()
	if !strings.HasPrefix(strings.TrimSpace(output), "[") {
		t.Errorf("expected JSON array, got: %q", output)
	}
}

func TestListMeetingsCmd_WithSource(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"list", "meetings", "--source", "zoom", "--limit", "5"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSyncCmd_WithSinceDate(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"sync", "--since", "2025-01-01"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := deps.Out.(*bytes.Buffer).String()
	if !strings.Contains(output, "Synced") {
		t.Errorf("expected sync output, got: %q", output)
	}
}

func TestSyncCmd_WithSinceRFC3339(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"sync", "--since", "2025-01-01T00:00:00Z"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSyncCmd_InvalidSince(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"sync", "--since", "not-a-date"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestExportMeetingCmd_JSONFormat(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"export", "meeting", "m-1", "--format", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := deps.Out.(*bytes.Buffer).String()
	if !strings.Contains(output, "m-1") {
		t.Errorf("expected meeting id in JSON output, got: %q", output)
	}
}

func TestExportMeetingCmd_MarkdownFormat(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"export", "meeting", "m-1", "--format", "md"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := deps.Out.(*bytes.Buffer).String()
	if !strings.Contains(output, "# Sprint Planning") {
		t.Errorf("expected markdown header, got: %q", output)
	}
}

// --- Note command tests ---

func TestNoteAddCmd(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"note", "add", "m-1", "Agent observation"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := deps.Out.(*bytes.Buffer).String()
	if !strings.Contains(output, "added to meeting") {
		t.Errorf("expected success message, got: %q", output)
	}
}

func TestNoteAddCmd_MissingArgs(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"note", "add", "m-1"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for missing text arg")
	}
}

func TestNoteListCmd(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"note", "list", "m-1"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := deps.Out.(*bytes.Buffer).String()
	if !strings.Contains(output, "ID") || !strings.Contains(output, "AUTHOR") {
		t.Errorf("expected table headers, got: %q", output)
	}
}

func TestNoteListCmd_JSONFormat(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"note", "list", "m-1", "--format", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNoteDeleteCmd_NotFound(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"note", "delete", "nonexistent"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for nonexistent note")
	}
}

// --- Action command tests ---

func TestActionCompleteCmd(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"action", "complete", "m-1", "ai-1"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := deps.Out.(*bytes.Buffer).String()
	if !strings.Contains(output, "completed") {
		t.Errorf("expected completed message, got: %q", output)
	}
}

func TestActionCompleteCmd_MissingArgs(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"action", "complete", "m-1"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for missing action_item_id arg")
	}
}

func TestActionUpdateCmd(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"action", "update", "m-1", "ai-1", "Updated text"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := deps.Out.(*bytes.Buffer).String()
	if !strings.Contains(output, "updated") {
		t.Errorf("expected updated message, got: %q", output)
	}
}

func TestActionUpdateCmd_MissingArgs(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"action", "update", "m-1", "ai-1"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for missing text arg")
	}
}

func TestAuthLoginCmd_APIToken(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"auth", "login", "--method", "api_token"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := deps.Out.(*bytes.Buffer).String()
	if !strings.Contains(output, "Authenticated successfully") {
		t.Errorf("expected success message, got: %q", output)
	}
}
