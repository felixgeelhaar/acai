// Package workspace contains application use cases for the workspace bounded context.
package workspace

import (
	"context"

	"github.com/felixgeelhaar/granola-mcp/internal/domain/workspace"
)

type ListWorkspacesInput struct{}

type ListWorkspacesOutput struct {
	Workspaces []*workspace.Workspace
}

type ListWorkspaces struct {
	repo workspace.Repository
}

func NewListWorkspaces(repo workspace.Repository) *ListWorkspaces {
	return &ListWorkspaces{repo: repo}
}

func (uc *ListWorkspaces) Execute(ctx context.Context, _ ListWorkspacesInput) (*ListWorkspacesOutput, error) {
	workspaces, err := uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	return &ListWorkspacesOutput{Workspaces: workspaces}, nil
}
