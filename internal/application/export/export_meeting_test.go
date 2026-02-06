package export_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/felixgeelhaar/granola-mcp/internal/application/export"
	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

type mockRepo struct {
	meetings map[domain.MeetingID]*domain.Meeting
}

func (m *mockRepo) FindByID(_ context.Context, id domain.MeetingID) (*domain.Meeting, error) {
	mtg, ok := m.meetings[id]
	if !ok {
		return nil, domain.ErrMeetingNotFound
	}
	return mtg, nil
}

func (m *mockRepo) List(_ context.Context, _ domain.ListFilter) ([]*domain.Meeting, error) {
	return nil, nil
}
func (m *mockRepo) GetTranscript(_ context.Context, _ domain.MeetingID) (*domain.Transcript, error) {
	return nil, nil
}
func (m *mockRepo) SearchTranscripts(_ context.Context, _ string, _ domain.ListFilter) ([]*domain.Meeting, error) {
	return nil, nil
}
func (m *mockRepo) GetActionItems(_ context.Context, _ domain.MeetingID) ([]*domain.ActionItem, error) {
	return nil, nil
}
func (m *mockRepo) Sync(_ context.Context, _ *time.Time) ([]domain.DomainEvent, error) {
	return nil, nil
}

func TestExportMeeting_Markdown(t *testing.T) {
	mtg, _ := domain.New("m-1", "Sprint Planning", time.Now().UTC(), domain.SourceZoom, []domain.Participant{
		domain.NewParticipant("Alice", "alice@example.com", domain.RoleHost),
	})
	mtg.AttachSummary(domain.NewSummary("m-1", "We planned the sprint.", domain.SummaryAuto))
	mtg.ClearDomainEvents()

	repo := &mockRepo{meetings: map[domain.MeetingID]*domain.Meeting{"m-1": mtg}}
	uc := export.NewExportMeeting(repo)

	out, err := uc.Execute(context.Background(), export.ExportMeetingInput{
		MeetingID: "m-1",
		Format:    export.FormatMarkdown,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.Content, "# Sprint Planning") {
		t.Error("markdown should contain title as h1")
	}
	if !strings.Contains(out.Content, "We planned the sprint.") {
		t.Error("markdown should contain summary")
	}
	if out.Format != export.FormatMarkdown {
		t.Errorf("got format %q", out.Format)
	}
}

func TestExportMeeting_JSON(t *testing.T) {
	mtg, _ := domain.New("m-1", "Meeting", time.Now().UTC(), domain.SourceZoom, nil)
	mtg.ClearDomainEvents()

	repo := &mockRepo{meetings: map[domain.MeetingID]*domain.Meeting{"m-1": mtg}}
	uc := export.NewExportMeeting(repo)

	out, err := uc.Execute(context.Background(), export.ExportMeetingInput{
		MeetingID: "m-1",
		Format:    export.FormatJSON,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.Content, `"id":"m-1"`) {
		t.Error("JSON should contain meeting id")
	}
}

func TestExportMeeting_NotFound(t *testing.T) {
	repo := &mockRepo{meetings: map[domain.MeetingID]*domain.Meeting{}}
	uc := export.NewExportMeeting(repo)

	_, err := uc.Execute(context.Background(), export.ExportMeetingInput{
		MeetingID: "nonexistent",
		Format:    export.FormatJSON,
	})
	if err != domain.ErrMeetingNotFound {
		t.Errorf("got error %v, want %v", err, domain.ErrMeetingNotFound)
	}
}

func TestExportMeeting_UnsupportedFormat(t *testing.T) {
	mtg, _ := domain.New("m-1", "Meeting", time.Now().UTC(), domain.SourceZoom, nil)
	mtg.ClearDomainEvents()

	repo := &mockRepo{meetings: map[domain.MeetingID]*domain.Meeting{"m-1": mtg}}
	uc := export.NewExportMeeting(repo)

	_, err := uc.Execute(context.Background(), export.ExportMeetingInput{
		MeetingID: "m-1",
		Format:    "xml",
	})
	if err != export.ErrUnsupportedFormat {
		t.Errorf("got error %v, want %v", err, export.ErrUnsupportedFormat)
	}
}

func TestExportMeeting_EmptyID(t *testing.T) {
	repo := &mockRepo{meetings: map[domain.MeetingID]*domain.Meeting{}}
	uc := export.NewExportMeeting(repo)

	_, err := uc.Execute(context.Background(), export.ExportMeetingInput{MeetingID: ""})
	if err != domain.ErrInvalidMeetingID {
		t.Errorf("got error %v, want %v", err, domain.ErrInvalidMeetingID)
	}
}
