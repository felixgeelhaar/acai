// Package annotation contains the Annotation bounded context.
// Agent notes are external to the Granola domain — they originate from MCP clients,
// not the meeting platform. Cross-context reference to meetings is by MeetingID string only.
package annotation

import "errors"

var (
	ErrInvalidNoteID      = errors.New("note id must not be empty")
	ErrInvalidMeetingID   = errors.New("meeting id must not be empty")
	ErrInvalidNoteContent = errors.New("note content must not be empty")
	ErrInvalidAuthor      = errors.New("note author must not be empty")
	ErrNoteNotFound       = errors.New("note not found")
)
