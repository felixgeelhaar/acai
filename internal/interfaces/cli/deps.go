// Package cli implements the CLI interface layer using cobra.
// Each command maps to an application use case.
package cli

import (
	"errors"
	"io"

	annotationapp "github.com/felixgeelhaar/acai/internal/application/annotation"
	authapp "github.com/felixgeelhaar/acai/internal/application/auth"
	embeddingapp "github.com/felixgeelhaar/acai/internal/application/embedding"
	exportapp "github.com/felixgeelhaar/acai/internal/application/export"
	meetingapp "github.com/felixgeelhaar/acai/internal/application/meeting"
	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
	mcpiface "github.com/felixgeelhaar/acai/internal/interfaces/mcp"
)

// errLocalDBRequired is returned when a write command is invoked but the
// local database was not available at startup.
var errLocalDBRequired = errors.New("this feature requires local storage (check ~/.acai/cache directory permissions)")

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
	Logout            *authapp.Logout
	EventDispatcher   domain.EventDispatcher
	MCPServer         *mcpiface.Server
	Out               io.Writer

	// Write use cases
	AddNote            *annotationapp.AddNote
	ListNotes          *annotationapp.ListNotes
	DeleteNote         *annotationapp.DeleteNote
	CompleteActionItem *meetingapp.CompleteActionItem
	UpdateActionItem   *meetingapp.UpdateActionItem

	// Embedding export
	ExportEmbeddings *embeddingapp.ExportEmbeddings

	// Config-provided API token for auth login
	GranolaAPIToken string
}
