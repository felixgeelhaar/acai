package meeting

// SummaryKind distinguishes between auto-generated and user-edited summaries.
type SummaryKind string

const (
	SummaryAuto   SummaryKind = "auto"
	SummaryEdited SummaryKind = "user_edited"
)

// Summary is an immutable value object for meeting summaries.
type Summary struct {
	meetingID MeetingID
	content   string
	kind      SummaryKind
}

func NewSummary(meetingID MeetingID, content string, kind SummaryKind) Summary {
	return Summary{
		meetingID: meetingID,
		content:   content,
		kind:      kind,
	}
}

func (s Summary) MeetingID() MeetingID { return s.meetingID }
func (s Summary) Content() string      { return s.content }
func (s Summary) Kind() SummaryKind    { return s.kind }

func (s Summary) Equals(other Summary) bool {
	return s.meetingID == other.meetingID && s.content == other.content && s.kind == other.kind
}
