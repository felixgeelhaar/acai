package workspace_test

import (
	"testing"

	"github.com/felixgeelhaar/acai/internal/domain/workspace"
)

func TestNew_Valid(t *testing.T) {
	ws, err := workspace.New("ws-1", "My Workspace", "my-workspace")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.ID() != "ws-1" {
		t.Errorf("got id %q", ws.ID())
	}
	if ws.Name() != "My Workspace" {
		t.Errorf("got name %q", ws.Name())
	}
	if ws.Slug() != "my-workspace" {
		t.Errorf("got slug %q", ws.Slug())
	}
}

func TestNew_RejectsEmptyID(t *testing.T) {
	_, err := workspace.New("", "Name", "slug")
	if err != workspace.ErrInvalidWorkspaceID {
		t.Errorf("got error %v, want %v", err, workspace.ErrInvalidWorkspaceID)
	}
}
