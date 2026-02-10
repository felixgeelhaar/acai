package localcache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestReaderRead(t *testing.T) {
	t.Run("reads sample cache file", func(t *testing.T) {
		reader := NewReader(filepath.Join("testdata", "cache-v3-sample.json"))
		state, err := reader.Read()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(state.State.Documents) != 3 {
			t.Errorf("expected 3 documents, got %d", len(state.State.Documents))
		}

		doc, ok := state.State.Documents["meeting-001"]
		if !ok {
			t.Fatal("expected meeting-001 document")
		}
		if doc.Title != "Weekly Standup" {
			t.Errorf("expected title 'Weekly Standup', got %q", doc.Title)
		}

		if len(state.State.MeetingsMetadata) != 2 {
			t.Errorf("expected 2 metadata entries, got %d", len(state.State.MeetingsMetadata))
		}

		if len(state.State.Transcripts) != 2 {
			t.Errorf("expected 2 transcripts, got %d", len(state.State.Transcripts))
		}
	})

	t.Run("missing file returns ErrCacheFileNotFound", func(t *testing.T) {
		reader := NewReader("/nonexistent/path/cache-v3.json")
		_, err := reader.Read()
		if err == nil {
			t.Fatal("expected error for missing file")
		}
		if !errors.Is(err, ErrCacheFileNotFound) {
			t.Errorf("expected ErrCacheFileNotFound, got: %v", err)
		}
	})

	t.Run("corrupt outer JSON returns ErrCacheFileCorrupt", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "corrupt.json")
		if err := os.WriteFile(path, []byte("{invalid json"), 0o644); err != nil {
			t.Fatal(err)
		}

		reader := NewReader(path)
		_, err := reader.Read()
		if err == nil {
			t.Fatal("expected error for corrupt file")
		}
		if !errors.Is(err, ErrCacheFileCorrupt) {
			t.Errorf("expected ErrCacheFileCorrupt, got: %v", err)
		}
	})

	t.Run("empty cache field returns ErrCacheFileCorrupt", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "empty-cache.json")
		data, _ := json.Marshal(CacheFileEnvelope{Cache: ""})
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}

		reader := NewReader(path)
		_, err := reader.Read()
		if err == nil {
			t.Fatal("expected error for empty cache field")
		}
		if !errors.Is(err, ErrCacheFileCorrupt) {
			t.Errorf("expected ErrCacheFileCorrupt, got: %v", err)
		}
	})

	t.Run("corrupt inner JSON returns ErrCacheFileCorrupt", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "corrupt-inner.json")
		data, _ := json.Marshal(CacheFileEnvelope{Cache: "{not valid inner json"})
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}

		reader := NewReader(path)
		_, err := reader.Read()
		if err == nil {
			t.Fatal("expected error for corrupt inner JSON")
		}
		if !errors.Is(err, ErrCacheFileCorrupt) {
			t.Errorf("expected ErrCacheFileCorrupt, got: %v", err)
		}
	})
}

func TestReaderPath(t *testing.T) {
	reader := NewReader("/some/path/cache.json")
	if reader.Path() != "/some/path/cache.json" {
		t.Errorf("expected path '/some/path/cache.json', got %q", reader.Path())
	}
}
