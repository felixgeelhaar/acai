// Package cli implements the CLI interface layer using cobra.
// Each command maps to an application use case.
package cli

import (
	"io"
	"net/http"

	annotationapp "github.com/felixgeelhaar/acai/internal/application/annotation"
	authapp "github.com/felixgeelhaar/acai/internal/application/auth"
	embeddingapp "github.com/felixgeelhaar/acai/internal/application/embedding"
	exportapp "github.com/felixgeelhaar/acai/internal/application/export"
	meetingapp "github.com/felixgeelhaar/acai/internal/application/meeting"
	workspaceapp "github.com/felixgeelhaar/acai/internal/application/workspace"
	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
	mcpiface "github.com/felixgeelhaar/acai/internal/interfaces/mcp"
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

	// Write use cases (Phase 3)
	AddNote            *annotationapp.AddNote
	ListNotes          *annotationapp.ListNotes
	DeleteNote         *annotationapp.DeleteNote
	CompleteActionItem *meetingapp.CompleteActionItem
	UpdateActionItem   *meetingapp.UpdateActionItem

	// Embedding export (Phase 3)
	ExportEmbeddings *embeddingapp.ExportEmbeddings
}
