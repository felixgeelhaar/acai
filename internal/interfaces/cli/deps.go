// Package cli implements the CLI interface layer using cobra.
// Each command maps to an application use case.
package cli

import (
	"io"

	meetingapp "github.com/felixgeelhaar/granola-mcp/internal/application/meeting"
	authapp "github.com/felixgeelhaar/granola-mcp/internal/application/auth"
	exportapp "github.com/felixgeelhaar/granola-mcp/internal/application/export"
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
	MCPServer         *mcpiface.Server
	Out               io.Writer
}
