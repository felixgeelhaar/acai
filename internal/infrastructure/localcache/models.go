package localcache

import "encoding/json"

// Cache file is double-JSON encoded:
// Outer JSON has a "cache" field containing a JSON string.
// That inner JSON string decodes to a CacheState with documents, metadata, and transcripts.

// CacheFileEnvelope is the outermost structure of cache-v3.json.
// The Cache field is itself a JSON-encoded string.
type CacheFileEnvelope struct {
	Cache string `json:"cache"`
}

// CacheState is the decoded inner JSON from the cache envelope.
type CacheState struct {
	State CacheInner `json:"state"`
}

// CacheInner holds the three main collections in the cache.
type CacheInner struct {
	Documents        map[string]CacheDocument    `json:"documents"`
	MeetingsMetadata map[string]CacheMeetingMeta `json:"meetingsMetadata"`
	Transcripts      map[string]CacheTranscript  `json:"transcripts"`
}

// CacheDocument represents a single meeting/note document in the cache.
type CacheDocument struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	CreatedAt        string          `json:"created_at"`
	UpdatedAt        string          `json:"updated_at"`
	LastViewedPanel  json.RawMessage `json:"last_viewed_panel"`
	NotesProsemirror json.RawMessage `json:"notes_prosemirror"`
}

// CacheMeetingMeta holds metadata for a meeting such as attendees and conference info.
type CacheMeetingMeta struct {
	Attendees  []CacheAttendee  `json:"attendees"`
	Organizer  *CacheAttendee   `json:"organizer"`
	Conference *CacheConference `json:"conference"`
}

// CacheAttendee represents a meeting participant.
type CacheAttendee struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CacheConference holds the conferencing platform info.
type CacheConference struct {
	Type string `json:"type"`
}

// CacheTranscript contains all transcript segments for a meeting.
// The Granola cache stores transcripts in two formats:
//   - Array directly: [{segment}, {segment}, ...]
//   - Object with segments field: {"segments": [{segment}, ...]}
//
// UnmarshalJSON handles both.
type CacheTranscript struct {
	Segments []CacheSegment `json:"segments"`
}

func (t *CacheTranscript) UnmarshalJSON(data []byte) error {
	// Try array format first (real Granola cache)
	var segments []CacheSegment
	if err := json.Unmarshal(data, &segments); err == nil {
		t.Segments = segments
		return nil
	}

	// Fall back to object format {"segments": [...]}
	type alias CacheTranscript
	var obj alias
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	t.Segments = obj.Segments
	return nil
}

// CacheSegment represents a single spoken segment in a transcript.
type CacheSegment struct {
	Speaker   string `json:"speaker"`
	Text      string `json:"text"`
	Source    string `json:"source"`
	Timestamp string `json:"timestamp"`
}
