package meeting

// Source represents the meeting platform origin.
type Source string

const (
	SourceZoom  Source = "zoom"
	SourceMeet  Source = "google_meet"
	SourceTeams Source = "teams"
	SourceOther Source = "other"
)

func (s Source) String() string {
	return string(s)
}
