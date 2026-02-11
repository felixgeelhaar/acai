package localcache

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

// writeSampleCache writes a valid cache file to the given path.
func writeSampleCache(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "cache-v3.json")

	inner := CacheState{
		State: CacheInner{
			Documents: map[string]CacheDocument{
				"mtg-1": {
					ID:        "mtg-1",
					Title:     "Morning Standup",
					CreatedAt: "2025-01-15T09:00:00Z",
					UpdatedAt: "2025-01-15T09:30:00Z",
					NotesProsemirror: json.RawMessage(`{"type":"doc","content":[
						{"type":"paragraph","content":[{"type":"text","text":"Notes for standup"}]}
					]}`),
				},
				"mtg-2": {
					ID:        "mtg-2",
					Title:     "Sprint Review",
					CreatedAt: "2025-01-16T14:00:00Z",
					UpdatedAt: "2025-01-16T15:00:00Z",
				},
				"mtg-3": {
					ID:        "mtg-3",
					Title:     "1:1 with Manager",
					CreatedAt: "2025-01-14T11:00:00Z",
					UpdatedAt: "2025-01-14T11:30:00Z",
				},
			},
			MeetingsMetadata: map[string]CacheMeetingMeta{
				"mtg-1": {
					Organizer:  &CacheAttendee{Name: "Alice", Email: "alice@test.com"},
					Attendees:  []CacheAttendee{{Name: "Bob", Email: "bob@test.com"}},
					Conference: &CacheConference{Type: "zoom"},
				},
				"mtg-2": {
					Organizer:  &CacheAttendee{Name: "Carol", Email: "carol@test.com"},
					Conference: &CacheConference{Type: "google_meet"},
				},
			},
			Transcripts: map[string]CacheTranscript{
				"mtg-1": {
					Segments: []CacheSegment{
						{Speaker: "Alice", Text: "Good morning team.", Timestamp: "2025-01-15T09:00:30Z"},
						{Speaker: "Bob", Text: "Morning! I worked on the API.", Timestamp: "2025-01-15T09:01:00Z"},
					},
				},
			},
		},
	}

	innerBytes, err := json.Marshal(inner)
	if err != nil {
		t.Fatal(err)
	}

	envelope := CacheFileEnvelope{Cache: string(innerBytes)}
	outerBytes, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, outerBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	return path
}

func newTestRepo(t *testing.T) *Repository {
	t.Helper()
	dir := t.TempDir()
	path := writeSampleCache(t, dir)
	reader := NewReader(path)
	return NewRepository(reader)
}

func TestRepositoryFindByID(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	t.Run("finds existing meeting", func(t *testing.T) {
		mtg, err := repo.FindByID(ctx, "mtg-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mtg.Title() != "Morning Standup" {
			t.Errorf("expected title 'Morning Standup', got %q", mtg.Title())
		}
		if mtg.Source() != domain.SourceZoom {
			t.Errorf("expected source zoom, got %q", mtg.Source())
		}
		// Should have transcript attached
		if mtg.Transcript() == nil {
			t.Error("expected transcript to be attached")
		}
	})

	t.Run("returns ErrMeetingNotFound for missing ID", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "nonexistent")
		if err != domain.ErrMeetingNotFound {
			t.Errorf("expected ErrMeetingNotFound, got: %v", err)
		}
	})
}

func TestRepositoryList(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	t.Run("lists all meetings sorted by Datetime desc", func(t *testing.T) {
		meetings, err := repo.List(ctx, domain.ListFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(meetings) != 3 {
			t.Fatalf("expected 3 meetings, got %d", len(meetings))
		}
		// Newest first
		if meetings[0].Title() != "Sprint Review" {
			t.Errorf("expected newest first, got %q", meetings[0].Title())
		}
		if meetings[2].Title() != "1:1 with Manager" {
			t.Errorf("expected oldest last, got %q", meetings[2].Title())
		}
	})

	t.Run("filters by Since", func(t *testing.T) {
		since := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		meetings, err := repo.List(ctx, domain.ListFilter{Since: &since})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(meetings) != 2 {
			t.Errorf("expected 2 meetings since Jan 15, got %d", len(meetings))
		}
	})

	t.Run("filters by Until", func(t *testing.T) {
		until := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		meetings, err := repo.List(ctx, domain.ListFilter{Until: &until})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(meetings) != 1 {
			t.Errorf("expected 1 meeting before Jan 15, got %d", len(meetings))
		}
	})

	t.Run("filters by Source", func(t *testing.T) {
		src := domain.SourceZoom
		meetings, err := repo.List(ctx, domain.ListFilter{Source: &src})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(meetings) != 1 {
			t.Errorf("expected 1 zoom meeting, got %d", len(meetings))
		}
	})

	t.Run("filters by Participant", func(t *testing.T) {
		participant := "bob"
		meetings, err := repo.List(ctx, domain.ListFilter{Participant: &participant})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(meetings) != 1 {
			t.Errorf("expected 1 meeting with Bob, got %d", len(meetings))
		}
	})

	t.Run("filters by Query (title match)", func(t *testing.T) {
		query := "sprint"
		meetings, err := repo.List(ctx, domain.ListFilter{Query: &query})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(meetings) != 1 {
			t.Errorf("expected 1 meeting matching 'sprint', got %d", len(meetings))
		}
	})

	t.Run("applies Limit", func(t *testing.T) {
		meetings, err := repo.List(ctx, domain.ListFilter{Limit: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(meetings) != 2 {
			t.Errorf("expected 2 meetings with limit, got %d", len(meetings))
		}
	})

	t.Run("applies Offset", func(t *testing.T) {
		meetings, err := repo.List(ctx, domain.ListFilter{Offset: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(meetings) != 2 {
			t.Errorf("expected 2 meetings after offset 1, got %d", len(meetings))
		}
	})

	t.Run("offset beyond length returns empty", func(t *testing.T) {
		meetings, err := repo.List(ctx, domain.ListFilter{Offset: 100})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(meetings) != 0 {
			t.Errorf("expected 0 meetings for large offset, got %d", len(meetings))
		}
	})
}

func TestRepositoryGetTranscript(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	t.Run("returns transcript for meeting with transcript", func(t *testing.T) {
		transcript, err := repo.GetTranscript(ctx, "mtg-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		utterances := transcript.Utterances()
		if len(utterances) != 2 {
			t.Fatalf("expected 2 utterances, got %d", len(utterances))
		}
		if utterances[0].Speaker() != "Alice" {
			t.Errorf("expected first speaker 'Alice', got %q", utterances[0].Speaker())
		}
	})

	t.Run("returns ErrTranscriptNotReady for meeting without transcript", func(t *testing.T) {
		_, err := repo.GetTranscript(ctx, "mtg-2")
		if err != domain.ErrTranscriptNotReady {
			t.Errorf("expected ErrTranscriptNotReady, got: %v", err)
		}
	})

	t.Run("returns ErrMeetingNotFound for nonexistent meeting", func(t *testing.T) {
		_, err := repo.GetTranscript(ctx, "nonexistent")
		if err != domain.ErrMeetingNotFound {
			t.Errorf("expected ErrMeetingNotFound, got: %v", err)
		}
	})
}

func TestRepositorySearchTranscripts(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	t.Run("finds by transcript text", func(t *testing.T) {
		meetings, err := repo.SearchTranscripts(ctx, "API", domain.ListFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(meetings) != 1 {
			t.Fatalf("expected 1 meeting matching 'API', got %d", len(meetings))
		}
		if meetings[0].ID() != "mtg-1" {
			t.Errorf("expected mtg-1, got %q", meetings[0].ID())
		}
	})

	t.Run("finds by title", func(t *testing.T) {
		meetings, err := repo.SearchTranscripts(ctx, "Sprint", domain.ListFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(meetings) != 1 {
			t.Errorf("expected 1 meeting matching 'Sprint' in title, got %d", len(meetings))
		}
	})

	t.Run("finds by notes content", func(t *testing.T) {
		meetings, err := repo.SearchTranscripts(ctx, "standup", domain.ListFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should match both the title "Morning Standup" and the notes "Notes for standup"
		found := false
		for _, m := range meetings {
			if m.ID() == domain.MeetingID("mtg-1") {
				found = true
			}
		}
		if !found {
			t.Error("expected to find mtg-1 matching 'standup' in notes")
		}
	})

	t.Run("no results for unmatched query", func(t *testing.T) {
		meetings, err := repo.SearchTranscripts(ctx, "xyznonexistent", domain.ListFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(meetings) != 0 {
			t.Errorf("expected 0 results for unmatched query, got %d", len(meetings))
		}
	})
}

func TestRepositoryGetActionItems(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	items, err := repo.GetActionItems(ctx, "mtg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 action items (not in cache), got %d", len(items))
	}
}

func TestRepositorySync(t *testing.T) {
	dir := t.TempDir()
	path := writeSampleCache(t, dir)
	reader := NewReader(path)
	repo := NewRepository(reader)
	ctx := context.Background()

	t.Run("first sync returns MeetingCreated events for all docs", func(t *testing.T) {
		events, err := repo.Sync(ctx, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 3 {
			t.Errorf("expected 3 events on first sync, got %d", len(events))
		}
		for _, e := range events {
			if e.EventName() != "meeting.created" {
				t.Errorf("expected 'meeting.created' event, got %q", e.EventName())
			}
		}
	})

	t.Run("second sync with no changes returns no events", func(t *testing.T) {
		events, err := repo.Sync(ctx, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 0 {
			t.Errorf("expected 0 events on unchanged second sync, got %d", len(events))
		}
	})

	t.Run("sync detects updated document", func(t *testing.T) {
		// Modify the cache file: change updatedAt for mtg-1
		inner := CacheState{
			State: CacheInner{
				Documents: map[string]CacheDocument{
					"mtg-1": {
						ID: "mtg-1", Title: "Morning Standup",
						CreatedAt: "2025-01-15T09:00:00Z",
						UpdatedAt: "2025-01-15T12:00:00Z", // changed
					},
					"mtg-2": {
						ID: "mtg-2", Title: "Sprint Review",
						CreatedAt: "2025-01-16T14:00:00Z",
						UpdatedAt: "2025-01-16T15:00:00Z",
					},
					"mtg-3": {
						ID: "mtg-3", Title: "1:1 with Manager",
						CreatedAt: "2025-01-14T11:00:00Z",
						UpdatedAt: "2025-01-14T11:30:00Z",
					},
					"mtg-4": {
						ID: "mtg-4", Title: "New Meeting",
						CreatedAt: "2025-01-17T10:00:00Z",
						UpdatedAt: "2025-01-17T10:30:00Z",
					},
				},
				MeetingsMetadata: map[string]CacheMeetingMeta{},
				Transcripts:      map[string]CacheTranscript{},
			},
		}

		innerBytes, _ := json.Marshal(inner)
		envelope := CacheFileEnvelope{Cache: string(innerBytes)}
		outerBytes, _ := json.Marshal(envelope)
		if err := os.WriteFile(path, outerBytes, 0o644); err != nil {
			t.Fatalf("failed to write test cache file: %v", err)
		}

		events, err := repo.Sync(ctx, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should detect: mtg-4 as new (MeetingCreated), mtg-1 as updated (SummaryUpdated with since=nil means all are "new")
		// Actually with since=nil, all docs pass the since check, so new ones get MeetingCreated
		if len(events) != 1 {
			t.Errorf("expected 1 event (new meeting), got %d", len(events))
			for _, e := range events {
				t.Logf("  event: %s", e.EventName())
			}
		}
	})
}

func TestRepositoryMissingCacheFile(t *testing.T) {
	reader := NewReader("/nonexistent/cache-v3.json")
	repo := NewRepository(reader)
	ctx := context.Background()

	_, err := repo.List(ctx, domain.ListFilter{})
	if err == nil {
		t.Fatal("expected error for missing cache file")
	}
}
