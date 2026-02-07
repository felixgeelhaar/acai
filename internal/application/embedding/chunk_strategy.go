// Package embedding provides use cases for embedding export functionality.
// It chunks meeting content (transcripts, summaries, notes) into segments
// suitable for vector embedding generation by external providers.
package embedding

import (
	"strings"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

// ChunkStrategy defines how meeting content is segmented into chunks.
type ChunkStrategy interface {
	ChunkTranscript(meetingID domain.MeetingID, utterances []domain.Utterance) ([]domain.Chunk, error)
}

// BySpeakerTurn groups consecutive utterances by the same speaker into a single chunk.
type BySpeakerTurn struct{}

func (s *BySpeakerTurn) ChunkTranscript(meetingID domain.MeetingID, utterances []domain.Utterance) ([]domain.Chunk, error) {
	if len(utterances) == 0 {
		return nil, nil
	}

	var chunks []domain.Chunk
	idx := 0
	i := 0
	for i < len(utterances) {
		speaker := utterances[i].Speaker()
		startTime := utterances[i].Timestamp()
		var texts []string
		endTime := startTime

		for i < len(utterances) && utterances[i].Speaker() == speaker {
			texts = append(texts, utterances[i].Text())
			endTime = utterances[i].Timestamp()
			i++
		}

		content := strings.Join(texts, " ")
		c, err := domain.NewChunk(meetingID, idx, content, speaker, startTime, endTime, domain.ChunkSourceTranscript, estimateTokens(content))
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, c)
		idx++
	}
	return chunks, nil
}

// ByTimeWindow groups utterances within a fixed-duration window.
type ByTimeWindow struct {
	Window time.Duration
}

func (s *ByTimeWindow) ChunkTranscript(meetingID domain.MeetingID, utterances []domain.Utterance) ([]domain.Chunk, error) {
	if len(utterances) == 0 {
		return nil, nil
	}

	window := s.Window
	if window <= 0 {
		window = 30 * time.Second
	}

	var chunks []domain.Chunk
	idx := 0
	i := 0
	for i < len(utterances) {
		windowStart := utterances[i].Timestamp()
		windowEnd := windowStart.Add(window)
		var texts []string
		var speakers []string
		endTime := windowStart

		for i < len(utterances) && !utterances[i].Timestamp().After(windowEnd) {
			texts = append(texts, utterances[i].Text())
			if len(speakers) == 0 || speakers[len(speakers)-1] != utterances[i].Speaker() {
				speakers = append(speakers, utterances[i].Speaker())
			}
			endTime = utterances[i].Timestamp()
			i++
		}

		content := strings.Join(texts, " ")
		speaker := strings.Join(speakers, ", ")
		c, err := domain.NewChunk(meetingID, idx, content, speaker, windowStart, endTime, domain.ChunkSourceTranscript, estimateTokens(content))
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, c)
		idx++
	}
	return chunks, nil
}

// ByTokenLimit splits content at approximate token boundaries.
type ByTokenLimit struct {
	MaxTokens int
}

func (s *ByTokenLimit) ChunkTranscript(meetingID domain.MeetingID, utterances []domain.Utterance) ([]domain.Chunk, error) {
	if len(utterances) == 0 {
		return nil, nil
	}

	maxTokens := s.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 256
	}

	var chunks []domain.Chunk
	idx := 0
	var texts []string
	var speakers []string
	tokenCount := 0
	startTime := utterances[0].Timestamp()
	endTime := startTime

	flush := func() error {
		if len(texts) == 0 {
			return nil
		}
		content := strings.Join(texts, " ")
		speaker := strings.Join(speakers, ", ")
		c, err := domain.NewChunk(meetingID, idx, content, speaker, startTime, endTime, domain.ChunkSourceTranscript, estimateTokens(content))
		if err != nil {
			return err
		}
		chunks = append(chunks, c)
		idx++
		texts = nil
		speakers = nil
		tokenCount = 0
		return nil
	}

	for _, u := range utterances {
		uTokens := estimateTokens(u.Text())
		if tokenCount+uTokens > maxTokens && len(texts) > 0 {
			if err := flush(); err != nil {
				return nil, err
			}
			startTime = u.Timestamp()
		}
		texts = append(texts, u.Text())
		if len(speakers) == 0 || speakers[len(speakers)-1] != u.Speaker() {
			speakers = append(speakers, u.Speaker())
		}
		tokenCount += uTokens
		endTime = u.Timestamp()
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return chunks, nil
}

// estimateTokens provides a rough token count approximation (~0.75 words per token).
func estimateTokens(text string) int {
	words := len(strings.Fields(text))
	tokens := int(float64(words) / 0.75)
	if tokens == 0 && words > 0 {
		tokens = 1
	}
	return tokens
}
