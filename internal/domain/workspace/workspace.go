// Package workspace contains the Workspace bounded context.
// A workspace represents a Granola organization that owns meetings.
package workspace

import (
	"context"
	"errors"
)

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrInvalidWorkspaceID = errors.New("workspace id must not be empty")
)

// WorkspaceID is a strongly-typed identifier for workspaces.
type WorkspaceID string

// Workspace is the aggregate root for workspace data.
type Workspace struct {
	id   WorkspaceID
	name string
	slug string
}

func New(id WorkspaceID, name, slug string) (*Workspace, error) {
	if id == "" {
		return nil, ErrInvalidWorkspaceID
	}
	return &Workspace{
		id:   id,
		name: name,
		slug: slug,
	}, nil
}

func (w *Workspace) ID() WorkspaceID { return w.id }
func (w *Workspace) Name() string    { return w.name }
func (w *Workspace) Slug() string    { return w.slug }

// Repository is the port for workspace persistence.
type Repository interface {
	List(ctx context.Context) ([]*Workspace, error)
	FindByID(ctx context.Context, id WorkspaceID) (*Workspace, error)
}
