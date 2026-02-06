package meeting

import "errors"

var (
	ErrInvalidMeetingID     = errors.New("meeting id must not be empty")
	ErrInvalidTitle         = errors.New("meeting title must not be empty")
	ErrInvalidDatetime      = errors.New("meeting datetime must not be zero")
	ErrInvalidActionItemID  = errors.New("action item id must not be empty")
	ErrInvalidActionItemText = errors.New("action item text must not be empty")
	ErrMeetingNotFound      = errors.New("meeting not found")
	ErrTranscriptNotReady   = errors.New("transcript not yet available")
	ErrAccessDenied         = errors.New("access denied to meeting")
	ErrInvalidFilter        = errors.New("invalid filter parameters")
)
