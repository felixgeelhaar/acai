package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/felixgeelhaar/granola-mcp/internal/interfaces/cli"
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
