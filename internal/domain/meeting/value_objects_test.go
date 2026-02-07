package meeting_test

import (
	"testing"
	"time"

	"github.com/felixgeelhaar/acai/internal/domain/meeting"
)

// --- Participant Value Object ---

func TestParticipant_Immutability(t *testing.T) {
	p := meeting.NewParticipant("Alice", "alice@example.com", meeting.RoleHost)

	if p.Name() != "Alice" {
		t.Errorf("got name %q, want %q", p.Name(), "Alice")
	}
	if p.Email() != "alice@example.com" {
		t.Errorf("got email %q, want %q", p.Email(), "alice@example.com")
	}
	if p.Role() != meeting.RoleHost {
		t.Errorf("got role %q, want %q", p.Role(), meeting.RoleHost)
	}
}

func TestParticipant_EqualityByValue(t *testing.T) {
	p1 := meeting.NewParticipant("Alice", "alice@example.com", meeting.RoleHost)
	p2 := meeting.NewParticipant("Alice", "alice@example.com", meeting.RoleHost)
	p3 := meeting.NewParticipant("Bob", "bob@example.com", meeting.RoleAttendee)

	if !p1.Equals(p2) {
		t.Error("identical participants should be equal")
	}
	if p1.Equals(p3) {
		t.Error("different participants should not be equal")
	}
}

// --- Utterance Value Object ---

func TestUtterance_Fields(t *testing.T) {
	ts := time.Now().UTC()
	u := meeting.NewUtterance("Alice", "Hello", ts, 0.95)

	if u.Speaker() != "Alice" {
		t.Errorf("got speaker %q", u.Speaker())
	}
	if u.Text() != "Hello" {
		t.Errorf("got text %q", u.Text())
	}
	if u.Timestamp() != ts {
		t.Error("timestamp mismatch")
	}
	if u.Confidence() != 0.95 {
		t.Errorf("got confidence %f", u.Confidence())
	}
}

// --- Transcript Value Object ---

func TestTranscript_Immutability(t *testing.T) {
	utterances := []meeting.Utterance{
		meeting.NewUtterance("Alice", "Hello", time.Now().UTC(), 0.9),
		meeting.NewUtterance("Bob", "Hi", time.Now().UTC(), 0.85),
	}
	tr := meeting.NewTranscript("m-1", utterances)

	if tr.MeetingID() != "m-1" {
		t.Errorf("got meeting id %q", tr.MeetingID())
	}
	if len(tr.Utterances()) != 2 {
		t.Errorf("got %d utterances", len(tr.Utterances()))
	}

	// Mutating returned slice should not affect transcript
	returned := tr.Utterances()
	returned[0] = meeting.NewUtterance("Hacker", "pwned", time.Now().UTC(), 0)
	if tr.Utterances()[0].Speaker() == "Hacker" {
		t.Error("transcript utterances should be defensively copied")
	}
}

// --- Summary Value Object ---

func TestSummary_Fields(t *testing.T) {
	s := meeting.NewSummary("m-1", "Summary content", meeting.SummaryAuto)

	if s.MeetingID() != "m-1" {
		t.Errorf("got meeting id %q", s.MeetingID())
	}
	if s.Content() != "Summary content" {
		t.Errorf("got content %q", s.Content())
	}
	if s.Kind() != meeting.SummaryAuto {
		t.Errorf("got kind %q", s.Kind())
	}
}

func TestSummary_EqualityByValue(t *testing.T) {
	s1 := meeting.NewSummary("m-1", "Content", meeting.SummaryAuto)
	s2 := meeting.NewSummary("m-1", "Content", meeting.SummaryAuto)
	s3 := meeting.NewSummary("m-1", "Different", meeting.SummaryEdited)

	if !s1.Equals(s2) {
		t.Error("identical summaries should be equal")
	}
	if s1.Equals(s3) {
		t.Error("different summaries should not be equal")
	}
}

// --- Metadata Value Object ---

func TestMetadata_DefensiveCopy(t *testing.T) {
	tags := []string{"sprint"}
	links := []string{"https://example.com"}
	refs := map[string]string{"jira": "SPRINT-1"}

	meta := meeting.NewMetadata(tags, links, refs)

	// Mutate originals
	tags[0] = "mutated"
	links[0] = "mutated"
	refs["jira"] = "mutated"

	if meta.Tags()[0] == "mutated" {
		t.Error("metadata tags should be defensively copied on construction")
	}
	if meta.Links()[0] == "mutated" {
		t.Error("metadata links should be defensively copied on construction")
	}
	if meta.ExternalRefs()["jira"] == "mutated" {
		t.Error("metadata refs should be defensively copied on construction")
	}

	// Mutate returned values
	meta.Tags()[0] = "mutated"
	if meta.Tags()[0] == "mutated" {
		t.Error("metadata tags should be defensively copied on access")
	}
}

func TestMetadata_Empty(t *testing.T) {
	meta := meeting.NewMetadata(nil, nil, nil)

	if meta.Tags() == nil {
		t.Error("tags should be empty slice, not nil")
	}
	if meta.Links() == nil {
		t.Error("links should be empty slice, not nil")
	}
	if meta.ExternalRefs() == nil {
		t.Error("refs should be empty map, not nil")
	}
}

// --- Source ---

func TestSource_ValidValues(t *testing.T) {
	sources := []meeting.Source{
		meeting.SourceZoom,
		meeting.SourceMeet,
		meeting.SourceTeams,
		meeting.SourceOther,
	}

	for _, s := range sources {
		if s.String() == "" {
			t.Errorf("source %v should have a string representation", s)
		}
	}
}
