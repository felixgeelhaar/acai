package mcp_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	meetingapp "github.com/felixgeelhaar/granola-mcp/internal/application/meeting"
	workspaceapp "github.com/felixgeelhaar/granola-mcp/internal/application/workspace"
	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
	"github.com/felixgeelhaar/granola-mcp/internal/domain/workspace"
	mcpiface "github.com/felixgeelhaar/granola-mcp/internal/interfaces/mcp"
)

func TestServer_HandleListMeetings(t *testing.T) {
	repo := newMockRepo()
	repo.addMeeting(mustMeeting(t, "m-1", "Sprint Planning"))
	repo.addMeeting(mustMeeting(t, "m-2", "Retrospective"))

	srv := newTestServer(repo)

	results, err := srv.HandleListMeetings(context.Background(), mcpiface.ListMeetingsToolInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("got %d results, want 2", len(results))
	}
}

func TestServer_HandleGetMeeting(t *testing.T) {
	repo := newMockRepo()
	m := mustMeeting(t, "m-1", "Sprint Planning")
	m.AttachSummary(domain.NewSummary("m-1", "Summary here", domain.SummaryAuto))
	m.ClearDomainEvents()
	repo.addMeeting(m)

	srv := newTestServer(repo)

	result, err := srv.HandleGetMeeting(context.Background(), mcpiface.GetMeetingToolInput{ID: "m-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "m-1" {
		t.Errorf("got id %q", result.ID)
	}
	if result.Summary == nil {
		t.Fatal("expected summary")
	}
	if result.Summary.Content != "Summary here" {
		t.Errorf("got summary %q", result.Summary.Content)
	}
}

func TestServer_HandleGetMeeting_NotFound(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)

	_, err := srv.HandleGetMeeting(context.Background(), mcpiface.GetMeetingToolInput{ID: "nonexistent"})
	if err != domain.ErrMeetingNotFound {
		t.Errorf("got error %v, want %v", err, domain.ErrMeetingNotFound)
	}
}

func TestServer_HandleGetTranscript(t *testing.T) {
	repo := newMockRepo()
	transcript := domain.NewTranscript("m-1", []domain.Utterance{
		domain.NewUtterance("Alice", "Hello everyone", time.Now().UTC(), 0.95),
	})
	repo.addTranscript("m-1", &transcript)

	srv := newTestServer(repo)

	result, err := srv.HandleGetTranscript(context.Background(), mcpiface.GetTranscriptToolInput{MeetingID: "m-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Utterances) != 1 {
		t.Errorf("got %d utterances", len(result.Utterances))
	}
	if result.Utterances[0].Speaker != "Alice" {
		t.Errorf("got speaker %q", result.Utterances[0].Speaker)
	}
}

func TestServer_HandleGetActionItems(t *testing.T) {
	repo := newMockRepo()
	item, _ := domain.NewActionItem("ai-1", "m-1", "Alice", "Write report", nil)
	repo.addActionItems("m-1", []*domain.ActionItem{item})

	srv := newTestServer(repo)

	results, err := srv.HandleGetActionItems(context.Background(), mcpiface.GetActionItemsToolInput{MeetingID: "m-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d items", len(results))
	}
	if results[0].Owner != "Alice" {
		t.Errorf("got owner %q", results[0].Owner)
	}
}

func TestServer_HandleToolJSON_ListMeetings(t *testing.T) {
	repo := newMockRepo()
	repo.addMeeting(mustMeeting(t, "m-1", "Test"))

	srv := newTestServer(repo)

	raw, err := srv.HandleToolJSON(context.Background(), "list_meetings", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var results []mcpiface.MeetingResult
	if err := json.Unmarshal(raw, &results); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("got %d results", len(results))
	}
}

func TestServer_HandleToolJSON_UnknownTool(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)

	_, err := srv.HandleToolJSON(context.Background(), "unknown_tool", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}

func TestServer_HandleMeetingStats(t *testing.T) {
	repo := newMockRepo()
	repo.addMeeting(mustMeeting(t, "m-1", "Sprint Planning"))
	repo.addMeeting(mustMeeting(t, "m-2", "Retrospective"))

	srv := newTestServer(repo)

	result, err := srv.HandleMeetingStats(context.Background(), mcpiface.MeetingStatsToolInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalMeetings != 2 {
		t.Errorf("got %d meetings, want 2", result.TotalMeetings)
	}
	if result.GeneratedAt == "" {
		t.Error("expected GeneratedAt to be set")
	}
}

func TestServer_HandleMeetingStats_Empty(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)

	result, err := srv.HandleMeetingStats(context.Background(), mcpiface.MeetingStatsToolInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalMeetings != 0 {
		t.Errorf("got %d meetings, want 0", result.TotalMeetings)
	}
}

func TestServer_HandleMeetingStats_InvalidSince(t *testing.T) {
	repo := newMockRepo()
	srv := newTestServer(repo)

	bad := "not-a-date"
	_, err := srv.HandleMeetingStats(context.Background(), mcpiface.MeetingStatsToolInput{
		Since: &bad,
	})
	if err == nil {
		t.Fatal("expected error for invalid since date")
	}
}

func TestServer_HandleToolJSON_MeetingStats(t *testing.T) {
	repo := newMockRepo()
	repo.addMeeting(mustMeeting(t, "m-1", "Test"))

	srv := newTestServer(repo)

	raw, err := srv.HandleToolJSON(context.Background(), "meeting_stats", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result mcpiface.MeetingStatsResult
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if result.TotalMeetings != 1 {
		t.Errorf("got %d meetings, want 1", result.TotalMeetings)
	}
}

// --- Test Helpers ---

type mockRepo struct {
	meetings    map[domain.MeetingID]*domain.Meeting
	transcripts map[domain.MeetingID]*domain.Transcript
	actionItems map[domain.MeetingID][]*domain.ActionItem
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		meetings:    make(map[domain.MeetingID]*domain.Meeting),
		transcripts: make(map[domain.MeetingID]*domain.Transcript),
		actionItems: make(map[domain.MeetingID][]*domain.ActionItem),
	}
}

func (m *mockRepo) FindByID(_ context.Context, id domain.MeetingID) (*domain.Meeting, error) {
	mtg, ok := m.meetings[id]
	if !ok {
		return nil, domain.ErrMeetingNotFound
	}
	return mtg, nil
}

func (m *mockRepo) List(_ context.Context, _ domain.ListFilter) ([]*domain.Meeting, error) {
	result := make([]*domain.Meeting, 0, len(m.meetings))
	for _, mtg := range m.meetings {
		result = append(result, mtg)
	}
	return result, nil
}

func (m *mockRepo) GetTranscript(_ context.Context, id domain.MeetingID) (*domain.Transcript, error) {
	t, ok := m.transcripts[id]
	if !ok {
		return nil, domain.ErrTranscriptNotReady
	}
	return t, nil
}

func (m *mockRepo) SearchTranscripts(_ context.Context, _ string, _ domain.ListFilter) ([]*domain.Meeting, error) {
	result := make([]*domain.Meeting, 0)
	for _, mtg := range m.meetings {
		result = append(result, mtg)
	}
	return result, nil
}

func (m *mockRepo) GetActionItems(_ context.Context, id domain.MeetingID) ([]*domain.ActionItem, error) {
	return m.actionItems[id], nil
}

func (m *mockRepo) Sync(_ context.Context, _ *time.Time) ([]domain.DomainEvent, error) {
	return nil, nil
}

func (m *mockRepo) addMeeting(mtg *domain.Meeting) {
	m.meetings[mtg.ID()] = mtg
}

func (m *mockRepo) addTranscript(id domain.MeetingID, t *domain.Transcript) {
	m.transcripts[id] = t
}

func (m *mockRepo) addActionItems(id domain.MeetingID, items []*domain.ActionItem) {
	m.actionItems[id] = items
}

func mustMeeting(t *testing.T, id domain.MeetingID, title string) *domain.Meeting {
	t.Helper()
	m, err := domain.New(id, title, time.Now().UTC(), domain.SourceZoom, nil)
	if err != nil {
		t.Fatal(err)
	}
	m.ClearDomainEvents()
	return m
}

type mockWorkspaceRepo struct {
	workspaces []*workspace.Workspace
}

func (m *mockWorkspaceRepo) List(_ context.Context) ([]*workspace.Workspace, error) {
	return m.workspaces, nil
}

func (m *mockWorkspaceRepo) FindByID(_ context.Context, id workspace.WorkspaceID) (*workspace.Workspace, error) {
	for _, ws := range m.workspaces {
		if ws.ID() == id {
			return ws, nil
		}
	}
	return nil, workspace.ErrWorkspaceNotFound
}

func newTestServer(repo *mockRepo) *mcpiface.Server {
	wsRepo := &mockWorkspaceRepo{}
	return mcpiface.NewServer("granola-mcp", "test", mcpiface.ServerOptions{
		ListMeetings:      meetingapp.NewListMeetings(repo),
		GetMeeting:        meetingapp.NewGetMeeting(repo),
		GetTranscript:     meetingapp.NewGetTranscript(repo),
		SearchTranscripts: meetingapp.NewSearchTranscripts(repo),
		GetActionItems:    meetingapp.NewGetActionItems(repo),
		GetMeetingStats:   meetingapp.NewGetMeetingStats(repo),
		ListWorkspaces:    workspaceapp.NewListWorkspaces(wsRepo),
		GetWorkspace:      workspaceapp.NewGetWorkspace(wsRepo),
	})
}

func newTestServerWithWorkspaces(repo *mockRepo, wsRepo *mockWorkspaceRepo) *mcpiface.Server {
	return mcpiface.NewServer("granola-mcp", "test", mcpiface.ServerOptions{
		ListMeetings:      meetingapp.NewListMeetings(repo),
		GetMeeting:        meetingapp.NewGetMeeting(repo),
		GetTranscript:     meetingapp.NewGetTranscript(repo),
		SearchTranscripts: meetingapp.NewSearchTranscripts(repo),
		GetActionItems:    meetingapp.NewGetActionItems(repo),
		GetMeetingStats:   meetingapp.NewGetMeetingStats(repo),
		ListWorkspaces:    workspaceapp.NewListWorkspaces(wsRepo),
		GetWorkspace:      workspaceapp.NewGetWorkspace(wsRepo),
	})
}
