package workspace_test

import (
	"context"
	"errors"
	"testing"

	workspaceapp "github.com/felixgeelhaar/granola-mcp/internal/application/workspace"
	"github.com/felixgeelhaar/granola-mcp/internal/domain/workspace"
)

type mockWorkspaceRepo struct {
	workspaces []*workspace.Workspace
	err        error
}

func (m *mockWorkspaceRepo) List(_ context.Context) ([]*workspace.Workspace, error) {
	return m.workspaces, m.err
}

func (m *mockWorkspaceRepo) FindByID(_ context.Context, id workspace.WorkspaceID) (*workspace.Workspace, error) {
	for _, ws := range m.workspaces {
		if ws.ID() == id {
			return ws, nil
		}
	}
	return nil, workspace.ErrWorkspaceNotFound
}

func mustWorkspace(t *testing.T, id, name, slug string) *workspace.Workspace {
	t.Helper()
	ws, err := workspace.New(workspace.WorkspaceID(id), name, slug)
	if err != nil {
		t.Fatal(err)
	}
	return ws
}

func TestListWorkspaces_ReturnsAll(t *testing.T) {
	repo := &mockWorkspaceRepo{
		workspaces: []*workspace.Workspace{
			mustWorkspace(t, "ws-1", "Engineering", "engineering"),
			mustWorkspace(t, "ws-2", "Design", "design"),
		},
	}
	uc := workspaceapp.NewListWorkspaces(repo)

	out, err := uc.Execute(context.Background(), workspaceapp.ListWorkspacesInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Workspaces) != 2 {
		t.Errorf("expected 2, got %d", len(out.Workspaces))
	}
}

func TestListWorkspaces_Empty(t *testing.T) {
	repo := &mockWorkspaceRepo{workspaces: []*workspace.Workspace{}}
	uc := workspaceapp.NewListWorkspaces(repo)

	out, err := uc.Execute(context.Background(), workspaceapp.ListWorkspacesInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Workspaces) != 0 {
		t.Errorf("expected 0, got %d", len(out.Workspaces))
	}
}

func TestListWorkspaces_RepositoryError(t *testing.T) {
	repo := &mockWorkspaceRepo{err: errors.New("api error")}
	uc := workspaceapp.NewListWorkspaces(repo)

	_, err := uc.Execute(context.Background(), workspaceapp.ListWorkspacesInput{})
	if err == nil {
		t.Fatal("expected error")
	}
}
