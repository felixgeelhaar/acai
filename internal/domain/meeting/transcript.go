package meeting

import "time"

// Utterance is an immutable value object representing a single spoken segment.
type Utterance struct {
	speaker    string
	text       string
	timestamp  time.Time
	confidence float64
}

func NewUtterance(speaker, text string, timestamp time.Time, confidence float64) Utterance {
	return Utterance{
		speaker:    speaker,
		text:       text,
		timestamp:  timestamp,
		confidence: confidence,
	}
}

func (u Utterance) Speaker() string      { return u.speaker }
func (u Utterance) Text() string         { return u.text }
func (u Utterance) Timestamp() time.Time { return u.timestamp }
func (u Utterance) Confidence() float64  { return u.confidence }

// Transcript is an immutable value object containing ordered utterances for a meeting.
type Transcript struct {
	meetingID  MeetingID
	utterances []Utterance
}

func NewTranscript(meetingID MeetingID, utterances []Utterance) Transcript {
	copied := make([]Utterance, len(utterances))
	copy(copied, utterances)
	return Transcript{
		meetingID:  meetingID,
		utterances: copied,
	}
}

func (t Transcript) MeetingID() MeetingID { return t.meetingID }

func (t Transcript) Utterances() []Utterance {
	copied := make([]Utterance, len(t.utterances))
	copy(copied, t.utterances)
	return copied
}
