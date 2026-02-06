package meeting

import (
	"errors"
	"time"
)

// ChunkSource identifies where a chunk's content originated.
type ChunkSource string

const (
	ChunkSourceTranscript ChunkSource = "transcript"
	ChunkSourceSummary    ChunkSource = "summary"
	ChunkSourceNote       ChunkSource = "note"
)

var (
	ErrInvalidChunkContent = errors.New("chunk content must not be empty")
	ErrInvalidChunkSource  = errors.New("chunk source must be transcript, summary, or note")
	ErrInvalidTokenCount   = errors.New("token count must be non-negative")
)

// Chunk is a value object representing a segment of meeting content
// suitable for embedding generation. It is immutable after creation.
type Chunk struct {
	meetingID  MeetingID
	chunkIndex int
	content    string
	speaker    string
	startTime  time.Time
	endTime    time.Time
	source     ChunkSource
	tokenCount int
}

// NewChunk creates a validated Chunk value object.
func NewChunk(
	meetingID MeetingID,
	chunkIndex int,
	content string,
	speaker string,
	startTime, endTime time.Time,
	source ChunkSource,
	tokenCount int,
) (Chunk, error) {
	if content == "" {
		return Chunk{}, ErrInvalidChunkContent
	}
	if !isValidChunkSource(source) {
		return Chunk{}, ErrInvalidChunkSource
	}
	if tokenCount < 0 {
		return Chunk{}, ErrInvalidTokenCount
	}
	return Chunk{
		meetingID:  meetingID,
		chunkIndex: chunkIndex,
		content:    content,
		speaker:    speaker,
		startTime:  startTime,
		endTime:    endTime,
		source:     source,
		tokenCount: tokenCount,
	}, nil
}

func isValidChunkSource(s ChunkSource) bool {
	switch s {
	case ChunkSourceTranscript, ChunkSourceSummary, ChunkSourceNote:
		return true
	default:
		return false
	}
}

func (c Chunk) MeetingID() MeetingID    { return c.meetingID }
func (c Chunk) ChunkIndex() int         { return c.chunkIndex }
func (c Chunk) Content() string          { return c.content }
func (c Chunk) Speaker() string          { return c.speaker }
func (c Chunk) StartTime() time.Time     { return c.startTime }
func (c Chunk) EndTime() time.Time       { return c.endTime }
func (c Chunk) Source() ChunkSource      { return c.source }
func (c Chunk) TokenCount() int          { return c.tokenCount }
