package meeting_test

import (
	"testing"
	"time"

	"github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

func TestMeetingCreated_Event(t *testing.T) {
	now := time.Now().UTC()
	event := meeting.NewMeetingCreatedEvent("m-1", "Sprint Planning", now)

	if event.EventName() != "meeting.created" {
		t.Errorf("got event name %q", event.EventName())
	}
	if event.MeetingID() != meeting.MeetingID("m-1") {
		t.Errorf("got meeting id %q", event.MeetingID())
	}
	if event.Title() != "Sprint Planning" {
		t.Errorf("got title %q", event.Title())
	}
	if event.OccurredAt().IsZero() {
		t.Error("occurred_at should not be zero")
	}
}

func TestTranscriptUpdated_Event(t *testing.T) {
	event := meeting.NewTranscriptUpdatedEvent("m-1", 42)

	if event.EventName() != "transcript.updated" {
		t.Errorf("got event name %q", event.EventName())
	}
	if event.UtteranceCount() != 42 {
		t.Errorf("got utterance count %d", event.UtteranceCount())
	}
}

func TestSummaryUpdated_Event(t *testing.T) {
	event := meeting.NewSummaryUpdatedEvent("m-1", meeting.SummaryAuto)

	if event.EventName() != "summary.updated" {
		t.Errorf("got event name %q", event.EventName())
	}
	if event.Kind() != meeting.SummaryAuto {
		t.Errorf("got kind %q", event.Kind())
	}
}
