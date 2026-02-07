package granola

import (
	"testing"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

func TestMapDocumentToDomain(t *testing.T) {
	now := time.Now().UTC()
	dto := DocumentDTO{
		ID:        "m-1",
		Title:     "Sprint Planning",
		CreatedAt: now,
		Source:    "zoom",
		Participants: []ParticipantDTO{
			{Name: "Alice", Email: "alice@example.com", Role: "host"},
			{Name: "Bob", Email: "bob@example.com", Role: "attendee"},
		},
		Summary: &SummaryDTO{Content: "We planned the sprint.", Type: "auto"},
		ActionItems: []ActionItemDTO{
			{ID: "ai-1", Owner: "Alice", Text: "Write report", Done: false},
		},
		Metadata: &MetadataDTO{
			Tags:         []string{"sprint"},
			Links:        []string{"https://jira.example.com"},
			ExternalRefs: map[string]string{"jira": "SPRINT-1"},
		},
	}

	mtg, err := mapDocumentToDomain(dto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mtg.ID() != "m-1" {
		t.Errorf("got id %q", mtg.ID())
	}
	if mtg.Title() != "Sprint Planning" {
		t.Errorf("got title %q", mtg.Title())
	}
	if mtg.Source() != domain.SourceZoom {
		t.Errorf("got source %q", mtg.Source())
	}
	if len(mtg.Participants()) != 2 {
		t.Errorf("got %d participants", len(mtg.Participants()))
	}
	if mtg.Summary() == nil {
		t.Fatal("summary should be attached")
	}
	if mtg.Summary().Content() != "We planned the sprint." {
		t.Errorf("got summary %q", mtg.Summary().Content())
	}
	if len(mtg.ActionItems()) != 1 {
		t.Errorf("got %d action items", len(mtg.ActionItems()))
	}
	if mtg.Metadata().Tags()[0] != "sprint" {
		t.Error("metadata tags not mapped")
	}

	// Reconstitution should not produce domain events
	if len(mtg.DomainEvents()) != 0 {
		t.Errorf("reconstituted meeting should have 0 events, got %d", len(mtg.DomainEvents()))
	}
}

func TestMapSourceToDomain(t *testing.T) {
	tests := []struct {
		input string
		want  domain.Source
	}{
		{"zoom", domain.SourceZoom},
		{"google_meet", domain.SourceMeet},
		{"teams", domain.SourceTeams},
		{"webex", domain.SourceOther},
		{"", domain.SourceOther},
	}

	for _, tt := range tests {
		got := mapSourceToDomain(tt.input)
		if got != tt.want {
			t.Errorf("mapSourceToDomain(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMapTranscriptToDomain(t *testing.T) {
	now := time.Now().UTC()
	dto := TranscriptResponse{
		MeetingID: "m-1",
		Utterances: []UtteranceDTO{
			{Speaker: "Alice", Text: "Hello", Timestamp: now, Confidence: 0.95},
			{Speaker: "Bob", Text: "Hi", Timestamp: now.Add(time.Second), Confidence: 0.90},
		},
	}

	transcript := mapTranscriptToDomain("m-1", dto)
	if transcript.MeetingID() != "m-1" {
		t.Errorf("got meeting id %q", transcript.MeetingID())
	}
	if len(transcript.Utterances()) != 2 {
		t.Errorf("got %d utterances", len(transcript.Utterances()))
	}
	if transcript.Utterances()[0].Speaker() != "Alice" {
		t.Errorf("got speaker %q", transcript.Utterances()[0].Speaker())
	}
}

func TestMapActionItemToDomain_Completed(t *testing.T) {
	dto := ActionItemDTO{
		ID:    "ai-1",
		Owner: "Alice",
		Text:  "Write report",
		Done:  true,
	}

	item, err := mapActionItemToDomain("m-1", dto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !item.IsCompleted() {
		t.Error("action item should be completed")
	}
}

func TestMapActionItemToDomain_InvalidSkipped(t *testing.T) {
	dto := ActionItemDTO{
		ID:   "",
		Text: "Invalid",
	}

	_, err := mapActionItemToDomain("m-1", dto)
	if err == nil {
		t.Error("expected error for empty ID")
	}
}
