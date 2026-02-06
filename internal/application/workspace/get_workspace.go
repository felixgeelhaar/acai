package workspace

import (
	"context"

	"github.com/felixgeelhaar/granola-mcp/internal/domain/workspace"
)

type GetWorkspaceInput struct {
	ID workspace.WorkspaceID
}

type GetWorkspaceOutput struct {
	Workspace *workspace.Workspace
}

type GetWorkspace struct {
	repo workspace.Repository
}

func NewGetWorkspace(repo workspace.Repository) *GetWorkspace {
	return &GetWorkspace{repo: repo}
}

func (uc *GetWorkspace) Execute(ctx context.Context, input GetWorkspaceInput) (*GetWorkspaceOutput, error) {
	ws, err := uc.repo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	return &GetWorkspaceOutput{Workspace: ws}, nil
}
