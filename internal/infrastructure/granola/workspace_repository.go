package granola

import (
	"context"

	"github.com/felixgeelhaar/acai/internal/domain/workspace"
)

// WorkspaceRepository implements workspace.Repository using the Granola API client.
type WorkspaceRepository struct {
	client *Client
}

// NewWorkspaceRepository creates a new workspace repository.
func NewWorkspaceRepository(client *Client) *WorkspaceRepository {
	return &WorkspaceRepository{client: client}
}

// List returns all workspaces from the Granola API.
func (r *WorkspaceRepository) List(ctx context.Context) ([]*workspace.Workspace, error) {
	resp, err := r.client.GetWorkspaces(ctx)
	if err != nil {
		return nil, err
	}

	workspaces := make([]*workspace.Workspace, 0, len(resp.Workspaces))
	for _, dto := range resp.Workspaces {
		ws, err := workspace.New(workspace.WorkspaceID(dto.ID), dto.Name, dto.Slug)
		if err != nil {
			continue // skip invalid workspaces
		}
		workspaces = append(workspaces, ws)
	}
	return workspaces, nil
}

// FindByID returns a specific workspace by ID.
// Since the Granola API has no get-by-id endpoint, we list all and filter.
func (r *WorkspaceRepository) FindByID(ctx context.Context, id workspace.WorkspaceID) (*workspace.Workspace, error) {
	workspaces, err := r.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, ws := range workspaces {
		if ws.ID() == id {
			return ws, nil
		}
	}
	return nil, workspace.ErrWorkspaceNotFound
}

// compile-time check
var _ workspace.Repository = (*WorkspaceRepository)(nil)
