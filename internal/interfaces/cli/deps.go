// Package cli implements the CLI interface layer using cobra.
// Each command maps to an application use case.
package cli

import (
	"io"
	"net/http"

	authapp "github.com/felixgeelhaar/granola-mcp/internal/application/auth"
	exportapp "github.com/felixgeelhaar/granola-mcp/internal/application/export"
	meetingapp "github.com/felixgeelhaar/granola-mcp/internal/application/meeting"
	workspaceapp "github.com/felixgeelhaar/granola-mcp/internal/application/workspace"
	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
	mcpiface "github.com/felixgeelhaar/granola-mcp/internal/interfaces/mcp"
)

// Dependencies holds all injected use cases for the CLI.
// This is the composition root's way of providing dependencies
// to the interface layer without any service locator.
type Dependencies struct {
	ListMeetings      *meetingapp.ListMeetings
	GetMeeting        *meetingapp.GetMeeting
	GetTranscript     *meetingapp.GetTranscript
	SearchTranscripts *meetingapp.SearchTranscripts
	GetActionItems    *meetingapp.GetActionItems
	SyncMeetings      *meetingapp.SyncMeetings
	ExportMeeting     *exportapp.ExportMeeting
	Login             *authapp.Login
	CheckStatus       *authapp.CheckStatus
	ListWorkspaces    *workspaceapp.ListWorkspaces
	GetWorkspace      *workspaceapp.GetWorkspace
	EventDispatcher   domain.EventDispatcher
	WebhookHandler    http.Handler
	MCPServer         *mcpiface.Server
	Out               io.Writer
}
