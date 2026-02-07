package workspace_test

import (
	"context"
	"errors"
	"testing"

	workspaceapp "github.com/felixgeelhaar/acai/internal/application/workspace"
	"github.com/felixgeelhaar/acai/internal/domain/workspace"
)

func TestGetWorkspace_Found(t *testing.T) {
	repo := &mockWorkspaceRepo{
		workspaces: []*workspace.Workspace{
			mustWorkspace(t, "ws-1", "Engineering", "engineering"),
		},
	}
	uc := workspaceapp.NewGetWorkspace(repo)

	out, err := uc.Execute(context.Background(), workspaceapp.GetWorkspaceInput{
		ID: "ws-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out.Workspace.ID()) != "ws-1" {
		t.Errorf("expected ws-1, got %s", out.Workspace.ID())
	}
}

func TestGetWorkspace_NotFound(t *testing.T) {
	repo := &mockWorkspaceRepo{
		workspaces: []*workspace.Workspace{
			mustWorkspace(t, "ws-1", "Engineering", "engineering"),
		},
	}
	uc := workspaceapp.NewGetWorkspace(repo)

	_, err := uc.Execute(context.Background(), workspaceapp.GetWorkspaceInput{
		ID: "nonexistent",
	})
	if !errors.Is(err, workspace.ErrWorkspaceNotFound) {
		t.Errorf("expected ErrWorkspaceNotFound, got %v", err)
	}
}

func TestGetWorkspace_RepositoryError(t *testing.T) {
	repo := &mockWorkspaceRepo{err: errors.New("api error")}
	uc := workspaceapp.NewGetWorkspace(repo)

	_, err := uc.Execute(context.Background(), workspaceapp.GetWorkspaceInput{
		ID: "ws-1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
