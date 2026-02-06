package meeting

import (
	"context"
	"sort"
	"strings"
	"time"

	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

// GetMeetingStatsInput specifies optional date bounds for statistics.
type GetMeetingStatsInput struct {
	Since *time.Time
	Until *time.Time
}

// GetMeetingStatsOutput contains all aggregated meeting statistics.
type GetMeetingStatsOutput struct {
	GeneratedAt          time.Time              `json:"generated_at"`
	TotalMeetings        int                    `json:"total_meetings"`
	DateRange            DateRange              `json:"date_range"`
	MeetingFrequency     []FrequencyEntry       `json:"meeting_frequency"`
	PlatformDistribution []PlatformEntry        `json:"platform_distribution"`
	TopParticipants      []ParticipantStatsEntry `json:"top_participants"`
	ActionItems          ActionItemStats        `json:"action_items"`
	DayOfWeekHeatmap     []HeatmapEntry         `json:"day_of_week_heatmap"`
	SpeakerTalkTime      []SpeakerEntry         `json:"speaker_talk_time"`
	SummaryCoverage      SummaryCoverageStats   `json:"summary_coverage"`
}

type DateRange struct {
	Earliest string `json:"earliest"`
	Latest   string `json:"latest"`
}

type FrequencyEntry struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type PlatformEntry struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}

type ParticipantStatsEntry struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	MeetingCount int    `json:"meeting_count"`
}

type ActionItemStats struct {
	Total          int     `json:"total"`
	Completed      int     `json:"completed"`
	CompletionRate float64 `json:"completion_rate"`
}

type HeatmapEntry struct {
	Day   int `json:"day"`
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

type SpeakerEntry struct {
	Speaker        string `json:"speaker"`
	WordCount      int    `json:"word_count"`
	UtteranceCount int    `json:"utterance_count"`
}

type SummaryCoverageStats struct {
	WithSummary    int     `json:"with_summary"`
	WithoutSummary int     `json:"without_summary"`
	CoverageRate   float64 `json:"coverage_rate"`
}

// GetMeetingStats aggregates meeting data into statistics.
type GetMeetingStats struct {
	repo domain.Repository
}

// NewGetMeetingStats creates a new GetMeetingStats use case.
func NewGetMeetingStats(repo domain.Repository) *GetMeetingStats {
	return &GetMeetingStats{repo: repo}
}

const (
	maxMeetingsForStats = 10000
	maxParticipants     = 15
	maxSpeakers         = 15
)

// Execute computes meeting statistics from repository data.
func (uc *GetMeetingStats) Execute(ctx context.Context, input GetMeetingStatsInput) (*GetMeetingStatsOutput, error) {
	filter := domain.ListFilter{
		Since: input.Since,
		Until: input.Until,
		Limit: maxMeetingsForStats,
	}

	meetings, err := uc.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	out := &GetMeetingStatsOutput{
		GeneratedAt:          time.Now().UTC(),
		TotalMeetings:        len(meetings),
		MeetingFrequency:     make([]FrequencyEntry, 0),
		PlatformDistribution: make([]PlatformEntry, 0),
		TopParticipants:      make([]ParticipantStatsEntry, 0),
		DayOfWeekHeatmap:     make([]HeatmapEntry, 0),
		SpeakerTalkTime:      make([]SpeakerEntry, 0),
	}

	if len(meetings) == 0 {
		return out, nil
	}

	out.DateRange = computeDateRange(meetings)
	out.MeetingFrequency = computeFrequency(meetings)
	out.PlatformDistribution = computePlatformDistribution(meetings)
	out.TopParticipants = computeTopParticipants(meetings)
	out.ActionItems = computeActionItemStats(meetings)
	out.DayOfWeekHeatmap = computeHeatmap(meetings)
	out.SummaryCoverage = computeSummaryCoverage(meetings)
	out.SpeakerTalkTime = uc.computeSpeakerTalkTime(ctx, meetings)

	return out, nil
}

func computeDateRange(meetings []*domain.Meeting) DateRange {
	earliest := meetings[0].Datetime()
	latest := meetings[0].Datetime()
	for _, m := range meetings[1:] {
		if m.Datetime().Before(earliest) {
			earliest = m.Datetime()
		}
		if m.Datetime().After(latest) {
			latest = m.Datetime()
		}
	}
	return DateRange{
		Earliest: earliest.Format("2006-01-02"),
		Latest:   latest.Format("2006-01-02"),
	}
}

func computeFrequency(meetings []*domain.Meeting) []FrequencyEntry {
	counts := make(map[string]int)
	for _, m := range meetings {
		date := m.Datetime().Format("2006-01-02")
		counts[date]++
	}

	entries := make([]FrequencyEntry, 0, len(counts))
	for date, count := range counts {
		entries = append(entries, FrequencyEntry{Date: date, Count: count})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Date < entries[j].Date
	})
	return entries
}

func computePlatformDistribution(meetings []*domain.Meeting) []PlatformEntry {
	counts := make(map[string]int)
	for _, m := range meetings {
		counts[string(m.Source())]++
	}

	entries := make([]PlatformEntry, 0, len(counts))
	for source, count := range counts {
		entries = append(entries, PlatformEntry{Source: source, Count: count})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Count > entries[j].Count
	})
	return entries
}

func computeTopParticipants(meetings []*domain.Meeting) []ParticipantStatsEntry {
	type participantKey struct {
		name  string
		email string
	}
	counts := make(map[participantKey]int)
	for _, m := range meetings {
		for _, p := range m.Participants() {
			key := participantKey{name: p.Name(), email: p.Email()}
			counts[key]++
		}
	}

	entries := make([]ParticipantStatsEntry, 0, len(counts))
	for key, count := range counts {
		entries = append(entries, ParticipantStatsEntry{
			Name:         key.name,
			Email:        key.email,
			MeetingCount: count,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].MeetingCount > entries[j].MeetingCount
	})
	if len(entries) > maxParticipants {
		entries = entries[:maxParticipants]
	}
	return entries
}

func computeActionItemStats(meetings []*domain.Meeting) ActionItemStats {
	var total, completed int
	for _, m := range meetings {
		for _, item := range m.ActionItems() {
			total++
			if item.IsCompleted() {
				completed++
			}
		}
	}
	var rate float64
	if total > 0 {
		rate = float64(completed) / float64(total)
	}
	return ActionItemStats{
		Total:          total,
		Completed:      completed,
		CompletionRate: rate,
	}
}

func computeHeatmap(meetings []*domain.Meeting) []HeatmapEntry {
	type dayHour struct {
		day  int
		hour int
	}
	counts := make(map[dayHour]int)
	for _, m := range meetings {
		dt := m.Datetime()
		key := dayHour{day: int(dt.Weekday()), hour: dt.Hour()}
		counts[key]++
	}

	entries := make([]HeatmapEntry, 0, len(counts))
	for key, count := range counts {
		entries = append(entries, HeatmapEntry{
			Day:   key.day,
			Hour:  key.hour,
			Count: count,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Day != entries[j].Day {
			return entries[i].Day < entries[j].Day
		}
		return entries[i].Hour < entries[j].Hour
	})
	return entries
}

func computeSummaryCoverage(meetings []*domain.Meeting) SummaryCoverageStats {
	var withSummary, withoutSummary int
	for _, m := range meetings {
		if m.Summary() != nil {
			withSummary++
		} else {
			withoutSummary++
		}
	}
	total := withSummary + withoutSummary
	var rate float64
	if total > 0 {
		rate = float64(withSummary) / float64(total)
	}
	return SummaryCoverageStats{
		WithSummary:    withSummary,
		WithoutSummary: withoutSummary,
		CoverageRate:   rate,
	}
}

func (uc *GetMeetingStats) computeSpeakerTalkTime(ctx context.Context, meetings []*domain.Meeting) []SpeakerEntry {
	type speakerStats struct {
		wordCount      int
		utteranceCount int
	}
	stats := make(map[string]*speakerStats)

	for _, m := range meetings {
		transcript, err := uc.repo.GetTranscript(ctx, m.ID())
		if err != nil {
			continue
		}
		for _, u := range transcript.Utterances() {
			s, ok := stats[u.Speaker()]
			if !ok {
				s = &speakerStats{}
				stats[u.Speaker()] = s
			}
			s.wordCount += len(strings.Fields(u.Text()))
			s.utteranceCount++
		}
	}

	entries := make([]SpeakerEntry, 0, len(stats))
	for speaker, s := range stats {
		entries = append(entries, SpeakerEntry{
			Speaker:        speaker,
			WordCount:      s.wordCount,
			UtteranceCount: s.utteranceCount,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].WordCount > entries[j].WordCount
	})
	if len(entries) > maxSpeakers {
		entries = entries[:maxSpeakers]
	}
	return entries
}
