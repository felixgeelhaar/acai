package meeting_test

import (
	"testing"
	"time"

	"github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

func TestNewMeeting_CreatesValidAggregate(t *testing.T) {
	now := time.Now().UTC()
	participants := []meeting.Participant{
		meeting.NewParticipant("Alice", "alice@example.com", meeting.RoleHost),
		meeting.NewParticipant("Bob", "bob@example.com", meeting.RoleAttendee),
	}

	m, err := meeting.New("m-1", "Sprint Planning", now, meeting.SourceZoom, participants)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.ID() != meeting.MeetingID("m-1") {
		t.Errorf("got id %q, want %q", m.ID(), "m-1")
	}
	if m.Title() != "Sprint Planning" {
		t.Errorf("got title %q, want %q", m.Title(), "Sprint Planning")
	}
	if m.Datetime() != now {
		t.Errorf("got datetime %v, want %v", m.Datetime(), now)
	}
	if m.Source() != meeting.SourceZoom {
		t.Errorf("got source %q, want %q", m.Source(), meeting.SourceZoom)
	}
	if len(m.Participants()) != 2 {
		t.Errorf("got %d participants, want 2", len(m.Participants()))
	}
}

func TestNewMeeting_RejectsEmptyID(t *testing.T) {
	now := time.Now().UTC()
	_, err := meeting.New("", "Title", now, meeting.SourceZoom, nil)
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
	if err != meeting.ErrInvalidMeetingID {
		t.Errorf("got error %v, want %v", err, meeting.ErrInvalidMeetingID)
	}
}

func TestNewMeeting_RejectsEmptyTitle(t *testing.T) {
	now := time.Now().UTC()
	_, err := meeting.New("m-1", "", now, meeting.SourceZoom, nil)
	if err == nil {
		t.Fatal("expected error for empty title")
	}
	if err != meeting.ErrInvalidTitle {
		t.Errorf("got error %v, want %v", err, meeting.ErrInvalidTitle)
	}
}

func TestNewMeeting_RejectsZeroDatetime(t *testing.T) {
	_, err := meeting.New("m-1", "Title", time.Time{}, meeting.SourceZoom, nil)
	if err == nil {
		t.Fatal("expected error for zero datetime")
	}
	if err != meeting.ErrInvalidDatetime {
		t.Errorf("got error %v, want %v", err, meeting.ErrInvalidDatetime)
	}
}

func TestMeeting_AttachTranscript(t *testing.T) {
	m := mustCreateMeeting(t)

	transcript := meeting.NewTranscript(m.ID(), []meeting.Utterance{
		meeting.NewUtterance("Alice", "Hello everyone", time.Now().UTC(), 0.95),
		meeting.NewUtterance("Bob", "Hi Alice", time.Now().UTC(), 0.90),
	})

	events := m.AttachTranscript(transcript)
	if m.Transcript() == nil {
		t.Fatal("transcript should be attached")
	}
	if len(m.Transcript().Utterances()) != 2 {
		t.Errorf("got %d utterances, want 2", len(m.Transcript().Utterances()))
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 domain event, got %d", len(events))
	}
	if events[0].EventName() != "transcript.updated" {
		t.Errorf("got event %q, want %q", events[0].EventName(), "transcript.updated")
	}
}

func TestMeeting_AttachSummary(t *testing.T) {
	m := mustCreateMeeting(t)

	summary := meeting.NewSummary(m.ID(), "Key decisions were made.", meeting.SummaryAuto)

	events := m.AttachSummary(summary)
	if m.Summary() == nil {
		t.Fatal("summary should be attached")
	}
	if m.Summary().Content() != "Key decisions were made." {
		t.Errorf("got content %q", m.Summary().Content())
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 domain event, got %d", len(events))
	}
	if events[0].EventName() != "summary.updated" {
		t.Errorf("got event %q, want %q", events[0].EventName(), "summary.updated")
	}
}

func TestMeeting_AddActionItem(t *testing.T) {
	m := mustCreateMeeting(t)
	due := time.Now().Add(48 * time.Hour).UTC()

	item, err := meeting.NewActionItem("ai-1", m.ID(), "Alice", "Send follow-up email", &due)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m.AddActionItem(item)
	items := m.ActionItems()
	if len(items) != 1 {
		t.Fatalf("got %d action items, want 1", len(items))
	}
	if items[0].Owner() != "Alice" {
		t.Errorf("got owner %q, want %q", items[0].Owner(), "Alice")
	}
	if items[0].IsCompleted() {
		t.Error("new action item should not be completed")
	}
}

func TestMeeting_SetMetadata(t *testing.T) {
	m := mustCreateMeeting(t)

	meta := meeting.NewMetadata(
		[]string{"sprint", "planning"},
		[]string{"https://jira.example.com/SPRINT-1"},
		map[string]string{"jira": "SPRINT-1"},
	)
	m.SetMetadata(meta)

	if len(m.Metadata().Tags()) != 2 {
		t.Errorf("got %d tags, want 2", len(m.Metadata().Tags()))
	}
	if m.Metadata().ExternalRefs()["jira"] != "SPRINT-1" {
		t.Error("external ref 'jira' not found")
	}
}

func TestMeeting_DomainEventsOnCreation(t *testing.T) {
	m := mustCreateMeeting(t)
	events := m.DomainEvents()

	if len(events) != 1 {
		t.Fatalf("expected 1 domain event on creation, got %d", len(events))
	}
	if events[0].EventName() != "meeting.created" {
		t.Errorf("got event %q, want %q", events[0].EventName(), "meeting.created")
	}
}

func TestMeeting_ClearDomainEvents(t *testing.T) {
	m := mustCreateMeeting(t)
	m.ClearDomainEvents()

	if len(m.DomainEvents()) != 0 {
		t.Errorf("events should be cleared, got %d", len(m.DomainEvents()))
	}
}

func TestMeeting_ParticipantsAreDefensivelyCopied(t *testing.T) {
	m := mustCreateMeeting(t)
	original := m.Participants()
	original[0] = meeting.NewParticipant("Mutated", "mutated@test.com", meeting.RoleAttendee)

	if m.Participants()[0].Name() == "Mutated" {
		t.Error("aggregate participants should not be mutated through returned slice")
	}
}

func mustCreateMeeting(t *testing.T) *meeting.Meeting {
	t.Helper()
	now := time.Now().UTC()
	participants := []meeting.Participant{
		meeting.NewParticipant("Alice", "alice@example.com", meeting.RoleHost),
	}
	m, err := meeting.New("m-1", "Sprint Planning", now, meeting.SourceZoom, participants)
	if err != nil {
		t.Fatalf("unexpected error creating meeting: %v", err)
	}
	return m
}
