package meeting_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	app "github.com/felixgeelhaar/granola-mcp/internal/application/meeting"
	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

func TestGetMeetingStats_EmptyRepository(t *testing.T) {
	repo := newMockRepository()
	uc := app.NewGetMeetingStats(repo)

	out, err := uc.Execute(context.Background(), app.GetMeetingStatsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.TotalMeetings != 0 {
		t.Errorf("got total %d, want 0", out.TotalMeetings)
	}
	if len(out.MeetingFrequency) != 0 {
		t.Errorf("got %d frequency entries, want 0", len(out.MeetingFrequency))
	}
	if len(out.PlatformDistribution) != 0 {
		t.Errorf("got %d platform entries, want 0", len(out.PlatformDistribution))
	}
	if len(out.TopParticipants) != 0 {
		t.Errorf("got %d participants, want 0", len(out.TopParticipants))
	}
	if out.ActionItems.Total != 0 {
		t.Errorf("got %d action items, want 0", out.ActionItems.Total)
	}
	if len(out.DayOfWeekHeatmap) != 0 {
		t.Errorf("got %d heatmap entries, want 0", len(out.DayOfWeekHeatmap))
	}
	if len(out.SpeakerTalkTime) != 0 {
		t.Errorf("got %d speaker entries, want 0", len(out.SpeakerTalkTime))
	}
	if out.SummaryCoverage.WithSummary != 0 {
		t.Errorf("got %d with summary, want 0", out.SummaryCoverage.WithSummary)
	}
	if out.GeneratedAt.IsZero() {
		t.Error("expected GeneratedAt to be set")
	}
}

func TestGetMeetingStats_SingleMeeting(t *testing.T) {
	repo := newMockRepository()
	now := time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC)

	m, err := domain.New("m-1", "Sprint Planning", now, domain.SourceZoom, []domain.Participant{
		domain.NewParticipant("Alice", "alice@test.com", domain.RoleHost),
		domain.NewParticipant("Bob", "bob@test.com", domain.RoleAttendee),
	})
	if err != nil {
		t.Fatal(err)
	}
	m.AttachSummary(domain.NewSummary("m-1", "Sprint planning summary", domain.SummaryAuto))
	item, _ := domain.NewActionItem("ai-1", "m-1", "Alice", "Review PR", nil)
	m.AddActionItem(item)
	m.ClearDomainEvents()
	repo.addMeeting(m)

	uc := app.NewGetMeetingStats(repo)
	out, err := uc.Execute(context.Background(), app.GetMeetingStatsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.TotalMeetings != 1 {
		t.Errorf("got total %d, want 1", out.TotalMeetings)
	}
	if len(out.MeetingFrequency) != 1 {
		t.Errorf("got %d frequency entries, want 1", len(out.MeetingFrequency))
	}
	if len(out.PlatformDistribution) != 1 {
		t.Errorf("got %d platform entries, want 1", len(out.PlatformDistribution))
	}
	if out.PlatformDistribution[0].Source != "zoom" {
		t.Errorf("got source %q, want zoom", out.PlatformDistribution[0].Source)
	}
	if len(out.TopParticipants) != 2 {
		t.Errorf("got %d participants, want 2", len(out.TopParticipants))
	}
	if out.ActionItems.Total != 1 {
		t.Errorf("got %d action items, want 1", out.ActionItems.Total)
	}
	if out.SummaryCoverage.WithSummary != 1 {
		t.Errorf("got %d with summary, want 1", out.SummaryCoverage.WithSummary)
	}
	if out.SummaryCoverage.WithoutSummary != 0 {
		t.Errorf("got %d without summary, want 0", out.SummaryCoverage.WithoutSummary)
	}
}

func TestGetMeetingStats_FrequencyComputation(t *testing.T) {
	repo := newMockRepository()

	// Two meetings on different dates
	m1, _ := domain.New("m-1", "Meeting 1", time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC), domain.SourceZoom, nil)
	m1.ClearDomainEvents()
	m2, _ := domain.New("m-2", "Meeting 2", time.Date(2025, 6, 1, 14, 0, 0, 0, time.UTC), domain.SourceZoom, nil)
	m2.ClearDomainEvents()
	m3, _ := domain.New("m-3", "Meeting 3", time.Date(2025, 6, 2, 10, 0, 0, 0, time.UTC), domain.SourceZoom, nil)
	m3.ClearDomainEvents()
	repo.addMeeting(m1)
	repo.addMeeting(m2)
	repo.addMeeting(m3)

	uc := app.NewGetMeetingStats(repo)
	out, err := uc.Execute(context.Background(), app.GetMeetingStatsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out.MeetingFrequency) != 2 {
		t.Errorf("got %d frequency entries, want 2", len(out.MeetingFrequency))
	}

	// Find the entry for June 1
	found := false
	for _, f := range out.MeetingFrequency {
		if f.Date == "2025-06-01" && f.Count == 2 {
			found = true
		}
	}
	if !found {
		t.Error("expected frequency entry for 2025-06-01 with count 2")
	}
}

func TestGetMeetingStats_PlatformDistribution(t *testing.T) {
	repo := newMockRepository()

	m1, _ := domain.New("m-1", "Zoom Meeting", time.Now().UTC(), domain.SourceZoom, nil)
	m1.ClearDomainEvents()
	m2, _ := domain.New("m-2", "Meet Meeting", time.Now().UTC(), domain.SourceMeet, nil)
	m2.ClearDomainEvents()
	m3, _ := domain.New("m-3", "Teams Meeting", time.Now().UTC(), domain.SourceTeams, nil)
	m3.ClearDomainEvents()
	m4, _ := domain.New("m-4", "Other Meeting", time.Now().UTC(), domain.SourceOther, nil)
	m4.ClearDomainEvents()
	repo.addMeeting(m1)
	repo.addMeeting(m2)
	repo.addMeeting(m3)
	repo.addMeeting(m4)

	uc := app.NewGetMeetingStats(repo)
	out, err := uc.Execute(context.Background(), app.GetMeetingStatsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out.PlatformDistribution) != 4 {
		t.Errorf("got %d platforms, want 4", len(out.PlatformDistribution))
	}

	sources := make(map[string]int)
	for _, pd := range out.PlatformDistribution {
		sources[pd.Source] = pd.Count
	}
	for _, src := range []string{"zoom", "google_meet", "teams", "other"} {
		if sources[src] != 1 {
			t.Errorf("got %d for %s, want 1", sources[src], src)
		}
	}
}

func TestGetMeetingStats_TopParticipants(t *testing.T) {
	repo := newMockRepository()

	// Alice appears in 3 meetings, Bob in 2, Carol in 1
	for i, id := range []domain.MeetingID{"m-1", "m-2", "m-3"} {
		participants := []domain.Participant{
			domain.NewParticipant("Alice", "alice@test.com", domain.RoleAttendee),
		}
		if i < 2 {
			participants = append(participants, domain.NewParticipant("Bob", "bob@test.com", domain.RoleAttendee))
		}
		if i == 0 {
			participants = append(participants, domain.NewParticipant("Carol", "carol@test.com", domain.RoleAttendee))
		}
		m, _ := domain.New(id, "Meeting", time.Now().UTC(), domain.SourceZoom, participants)
		m.ClearDomainEvents()
		repo.addMeeting(m)
	}

	uc := app.NewGetMeetingStats(repo)
	out, err := uc.Execute(context.Background(), app.GetMeetingStatsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out.TopParticipants) != 3 {
		t.Errorf("got %d participants, want 3", len(out.TopParticipants))
	}
	// Should be sorted by meeting count descending
	if out.TopParticipants[0].Name != "Alice" {
		t.Errorf("got top participant %q, want Alice", out.TopParticipants[0].Name)
	}
	if out.TopParticipants[0].MeetingCount != 3 {
		t.Errorf("got count %d for Alice, want 3", out.TopParticipants[0].MeetingCount)
	}
}

func TestGetMeetingStats_TopParticipants_CappedAt15(t *testing.T) {
	repo := newMockRepository()

	// Create a meeting with 20 participants
	participants := make([]domain.Participant, 20)
	for i := range 20 {
		participants[i] = domain.NewParticipant(
			"Person"+string(rune('A'+i)),
			"person"+string(rune('a'+i))+"@test.com",
			domain.RoleAttendee,
		)
	}
	m, _ := domain.New("m-1", "Big Meeting", time.Now().UTC(), domain.SourceZoom, participants)
	m.ClearDomainEvents()
	repo.addMeeting(m)

	uc := app.NewGetMeetingStats(repo)
	out, err := uc.Execute(context.Background(), app.GetMeetingStatsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out.TopParticipants) > 15 {
		t.Errorf("got %d participants, want at most 15", len(out.TopParticipants))
	}
}

func TestGetMeetingStats_ActionItemCompletionRate(t *testing.T) {
	repo := newMockRepository()

	m, _ := domain.New("m-1", "Meeting", time.Now().UTC(), domain.SourceZoom, nil)
	item1, _ := domain.NewActionItem("ai-1", "m-1", "Alice", "Task 1", nil)
	item1.Complete()
	item2, _ := domain.NewActionItem("ai-2", "m-1", "Bob", "Task 2", nil)
	item3, _ := domain.NewActionItem("ai-3", "m-1", "Alice", "Task 3", nil)
	item3.Complete()
	m.AddActionItem(item1)
	m.AddActionItem(item2)
	m.AddActionItem(item3)
	m.ClearDomainEvents()
	repo.addMeeting(m)

	uc := app.NewGetMeetingStats(repo)
	out, err := uc.Execute(context.Background(), app.GetMeetingStatsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.ActionItems.Total != 3 {
		t.Errorf("got total %d, want 3", out.ActionItems.Total)
	}
	if out.ActionItems.Completed != 2 {
		t.Errorf("got completed %d, want 2", out.ActionItems.Completed)
	}
	// 2/3 ≈ 0.6667
	if out.ActionItems.CompletionRate < 0.66 || out.ActionItems.CompletionRate > 0.67 {
		t.Errorf("got completion rate %f, want ~0.6667", out.ActionItems.CompletionRate)
	}
}

func TestGetMeetingStats_DayOfWeekHeatmap(t *testing.T) {
	repo := newMockRepository()

	// Monday at 9am and 10am
	mon9 := time.Date(2025, 6, 2, 9, 0, 0, 0, time.UTC) // Monday
	mon10 := time.Date(2025, 6, 2, 10, 0, 0, 0, time.UTC)
	// Wednesday at 14
	wed14 := time.Date(2025, 6, 4, 14, 0, 0, 0, time.UTC)

	m1, _ := domain.New("m-1", "Mon 9am", mon9, domain.SourceZoom, nil)
	m1.ClearDomainEvents()
	m2, _ := domain.New("m-2", "Mon 10am", mon10, domain.SourceZoom, nil)
	m2.ClearDomainEvents()
	m3, _ := domain.New("m-3", "Wed 2pm", wed14, domain.SourceZoom, nil)
	m3.ClearDomainEvents()
	repo.addMeeting(m1)
	repo.addMeeting(m2)
	repo.addMeeting(m3)

	uc := app.NewGetMeetingStats(repo)
	out, err := uc.Execute(context.Background(), app.GetMeetingStatsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out.DayOfWeekHeatmap) != 3 {
		t.Errorf("got %d heatmap entries, want 3", len(out.DayOfWeekHeatmap))
	}

	heatmap := make(map[[2]int]int) // [day, hour] -> count
	for _, h := range out.DayOfWeekHeatmap {
		heatmap[[2]int{h.Day, h.Hour}] = h.Count
	}

	// Monday=1 (time.Weekday), 9am
	if heatmap[[2]int{1, 9}] != 1 {
		t.Errorf("got %d for Mon 9am, want 1", heatmap[[2]int{1, 9}])
	}
	if heatmap[[2]int{1, 10}] != 1 {
		t.Errorf("got %d for Mon 10am, want 1", heatmap[[2]int{1, 10}])
	}
	// Wednesday=3, 14
	if heatmap[[2]int{3, 14}] != 1 {
		t.Errorf("got %d for Wed 2pm, want 1", heatmap[[2]int{3, 14}])
	}
}

func TestGetMeetingStats_SpeakerTalkTime(t *testing.T) {
	repo := newMockRepository()

	m, _ := domain.New("m-1", "Meeting", time.Now().UTC(), domain.SourceZoom, nil)
	m.ClearDomainEvents()
	repo.addMeeting(m)

	transcript := domain.NewTranscript("m-1", []domain.Utterance{
		domain.NewUtterance("Alice", "Hello everyone welcome to the meeting", time.Now().UTC(), 0.95),
		domain.NewUtterance("Bob", "Thanks Alice", time.Now().UTC(), 0.90),
		domain.NewUtterance("Alice", "Let us discuss the sprint goals for this week", time.Now().UTC(), 0.92),
	})
	repo.addTranscript("m-1", &transcript)

	uc := app.NewGetMeetingStats(repo)
	out, err := uc.Execute(context.Background(), app.GetMeetingStatsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out.SpeakerTalkTime) != 2 {
		t.Errorf("got %d speakers, want 2", len(out.SpeakerTalkTime))
	}

	// Alice should be first (more words)
	if out.SpeakerTalkTime[0].Speaker != "Alice" {
		t.Errorf("got top speaker %q, want Alice", out.SpeakerTalkTime[0].Speaker)
	}
	if out.SpeakerTalkTime[0].UtteranceCount != 2 {
		t.Errorf("got %d utterances for Alice, want 2", out.SpeakerTalkTime[0].UtteranceCount)
	}
}

func TestGetMeetingStats_SummaryCoverage(t *testing.T) {
	repo := newMockRepository()

	m1, _ := domain.New("m-1", "With Summary", time.Now().UTC(), domain.SourceZoom, nil)
	m1.AttachSummary(domain.NewSummary("m-1", "Summary content", domain.SummaryAuto))
	m1.ClearDomainEvents()
	repo.addMeeting(m1)

	m2, _ := domain.New("m-2", "Without Summary", time.Now().UTC(), domain.SourceZoom, nil)
	m2.ClearDomainEvents()
	repo.addMeeting(m2)

	m3, _ := domain.New("m-3", "Also With Summary", time.Now().UTC(), domain.SourceZoom, nil)
	m3.AttachSummary(domain.NewSummary("m-3", "Another summary", domain.SummaryEdited))
	m3.ClearDomainEvents()
	repo.addMeeting(m3)

	uc := app.NewGetMeetingStats(repo)
	out, err := uc.Execute(context.Background(), app.GetMeetingStatsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.SummaryCoverage.WithSummary != 2 {
		t.Errorf("got %d with summary, want 2", out.SummaryCoverage.WithSummary)
	}
	if out.SummaryCoverage.WithoutSummary != 1 {
		t.Errorf("got %d without summary, want 1", out.SummaryCoverage.WithoutSummary)
	}
	// 2/3 ≈ 0.6667
	if out.SummaryCoverage.CoverageRate < 0.66 || out.SummaryCoverage.CoverageRate > 0.67 {
		t.Errorf("got coverage rate %f, want ~0.6667", out.SummaryCoverage.CoverageRate)
	}
}

func TestGetMeetingStats_WithDateFilter(t *testing.T) {
	repo := newMockRepository()

	m1, _ := domain.New("m-1", "Meeting", time.Now().UTC(), domain.SourceZoom, nil)
	m1.ClearDomainEvents()
	repo.addMeeting(m1)

	uc := app.NewGetMeetingStats(repo)
	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	_, err := uc.Execute(context.Background(), app.GetMeetingStatsInput{
		Since: &since,
		Until: &until,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.listCalled {
		t.Error("expected List to be called")
	}
	if repo.listFilter == nil {
		t.Fatal("expected filter to be set")
	}
	if repo.listFilter.Since == nil || !repo.listFilter.Since.Equal(since) {
		t.Error("expected Since filter to be passed to repository")
	}
	if repo.listFilter.Until == nil || !repo.listFilter.Until.Equal(until) {
		t.Error("expected Until filter to be passed to repository")
	}
}

func TestGetMeetingStats_RepositoryError(t *testing.T) {
	repo := newMockRepository()
	repo.listErr = errors.New("database connection failed")

	uc := app.NewGetMeetingStats(repo)
	_, err := uc.Execute(context.Background(), app.GetMeetingStatsInput{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "database connection failed") {
		t.Errorf("got error %q, want to contain 'database connection failed'", err.Error())
	}
}

func TestGetMeetingStats_DateRange(t *testing.T) {
	repo := newMockRepository()

	early := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	late := time.Date(2025, 6, 20, 14, 0, 0, 0, time.UTC)

	m1, _ := domain.New("m-1", "Early", early, domain.SourceZoom, nil)
	m1.ClearDomainEvents()
	m2, _ := domain.New("m-2", "Late", late, domain.SourceZoom, nil)
	m2.ClearDomainEvents()
	repo.addMeeting(m1)
	repo.addMeeting(m2)

	uc := app.NewGetMeetingStats(repo)
	out, err := uc.Execute(context.Background(), app.GetMeetingStatsInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.DateRange.Earliest != "2025-01-15" {
		t.Errorf("got earliest %q, want 2025-01-15", out.DateRange.Earliest)
	}
	if out.DateRange.Latest != "2025-06-20" {
		t.Errorf("got latest %q, want 2025-06-20", out.DateRange.Latest)
	}
}
