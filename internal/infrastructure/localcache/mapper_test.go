package localcache

import (
	"encoding/json"
	"testing"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

func TestMapDocumentToDomain(t *testing.T) {
	t.Run("full mapping with metadata", func(t *testing.T) {
		doc := CacheDocument{
			ID:        "test-id",
			Title:     "Test Meeting",
			CreatedAt: "2025-01-15T10:00:00Z",
			UpdatedAt: "2025-01-15T10:30:00Z",
			NotesProsemirror: json.RawMessage(`{"type":"doc","content":[
				{"type":"paragraph","content":[{"type":"text","text":"Meeting notes here"}]}
			]}`),
		}
		meta := &CacheMeetingMeta{
			Organizer: &CacheAttendee{Name: "Alice", Email: "alice@example.com"},
			Attendees: []CacheAttendee{
				{Name: "Alice", Email: "alice@example.com"},
				{Name: "Bob", Email: "bob@example.com"},
			},
			Conference: &CacheConference{Type: "zoom"},
		}

		mtg, err := mapDocumentToDomain(doc, meta)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mtg.ID() != "test-id" {
			t.Errorf("expected ID 'test-id', got %q", mtg.ID())
		}
		if mtg.Title() != "Test Meeting" {
			t.Errorf("expected title 'Test Meeting', got %q", mtg.Title())
		}
		if mtg.Source() != domain.SourceZoom {
			t.Errorf("expected source zoom, got %q", mtg.Source())
		}

		participants := mtg.Participants()
		if len(participants) != 2 {
			t.Fatalf("expected 2 participants, got %d", len(participants))
		}
		if participants[0].Role() != domain.RoleHost {
			t.Errorf("expected first participant to be host, got %q", participants[0].Role())
		}
		if participants[1].Role() != domain.RoleAttendee {
			t.Errorf("expected second participant to be attendee, got %q", participants[1].Role())
		}

		if mtg.Summary() == nil {
			t.Fatal("expected summary to be set")
		}
		if mtg.Summary().Content() != "Meeting notes here" {
			t.Errorf("expected summary 'Meeting notes here', got %q", mtg.Summary().Content())
		}

		// Reconstitution should not produce domain events
		if len(mtg.DomainEvents()) != 0 {
			t.Errorf("expected no domain events, got %d", len(mtg.DomainEvents()))
		}
	})

	t.Run("no metadata", func(t *testing.T) {
		doc := CacheDocument{
			ID:        "no-meta-id",
			Title:     "No Meta Meeting",
			CreatedAt: "2025-01-15T10:00:00Z",
			UpdatedAt: "2025-01-15T10:00:00Z",
		}

		mtg, err := mapDocumentToDomain(doc, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mtg.Source() != domain.SourceOther {
			t.Errorf("expected source 'other', got %q", mtg.Source())
		}
		if len(mtg.Participants()) != 0 {
			t.Errorf("expected 0 participants, got %d", len(mtg.Participants()))
		}
	})

	t.Run("invalid timestamp uses current time", func(t *testing.T) {
		doc := CacheDocument{
			ID:        "bad-time-id",
			Title:     "Bad Time",
			CreatedAt: "not-a-date",
			UpdatedAt: "not-a-date",
		}

		mtg, err := mapDocumentToDomain(doc, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mtg.Datetime().IsZero() {
			t.Error("expected non-zero datetime despite bad timestamp")
		}
	})

	t.Run("empty title fails validation", func(t *testing.T) {
		doc := CacheDocument{
			ID:        "empty-title",
			Title:     "",
			CreatedAt: "2025-01-15T10:00:00Z",
			UpdatedAt: "2025-01-15T10:00:00Z",
		}

		_, err := mapDocumentToDomain(doc, nil)
		if err == nil {
			t.Fatal("expected error for empty title")
		}
	})
}

func TestMapTranscriptToDomain(t *testing.T) {
	t.Run("maps segments to utterances", func(t *testing.T) {
		transcript := CacheTranscript{
			Segments: []CacheSegment{
				{Speaker: "Alice", Text: "Hello", Timestamp: "2025-01-15T10:00:30Z"},
				{Speaker: "Bob", Text: "Hi there", Timestamp: "2025-01-15T10:01:00Z"},
			},
		}

		result := mapTranscriptToDomain("meeting-id", transcript)
		if result == nil {
			t.Fatal("expected non-nil transcript")
		}

		utterances := result.Utterances()
		if len(utterances) != 2 {
			t.Fatalf("expected 2 utterances, got %d", len(utterances))
		}
		if utterances[0].Speaker() != "Alice" {
			t.Errorf("expected speaker 'Alice', got %q", utterances[0].Speaker())
		}
		if utterances[0].Confidence() != 0 {
			t.Errorf("expected confidence 0, got %f", utterances[0].Confidence())
		}
	})

	t.Run("empty segments returns nil", func(t *testing.T) {
		transcript := CacheTranscript{Segments: []CacheSegment{}}
		result := mapTranscriptToDomain("meeting-id", transcript)
		if result != nil {
			t.Error("expected nil for empty segments")
		}
	})

	t.Run("bad timestamp uses zero time", func(t *testing.T) {
		transcript := CacheTranscript{
			Segments: []CacheSegment{
				{Speaker: "Alice", Text: "Hello", Timestamp: "invalid"},
			},
		}

		result := mapTranscriptToDomain("meeting-id", transcript)
		if result == nil {
			t.Fatal("expected non-nil transcript")
		}

		utterances := result.Utterances()
		if !utterances[0].Timestamp().IsZero() {
			t.Errorf("expected zero time for bad timestamp, got %v", utterances[0].Timestamp())
		}
	})
}

func TestMapConferenceToSource(t *testing.T) {
	tests := []struct {
		conf *CacheConference
		want domain.Source
	}{
		{nil, domain.SourceOther},
		{&CacheConference{Type: "zoom"}, domain.SourceZoom},
		{&CacheConference{Type: "google_meet"}, domain.SourceMeet},
		{&CacheConference{Type: "teams"}, domain.SourceTeams},
		{&CacheConference{Type: "unknown"}, domain.SourceOther},
		{&CacheConference{Type: ""}, domain.SourceOther},
	}

	for _, tt := range tests {
		name := "nil"
		if tt.conf != nil {
			name = tt.conf.Type
			if name == "" {
				name = "empty"
			}
		}
		t.Run(name, func(t *testing.T) {
			got := mapConferenceToSource(tt.conf)
			if got != tt.want {
				t.Errorf("mapConferenceToSource() = %q, want %q", got, tt.want)
			}
		})
	}
}
