package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	annotationapp "github.com/felixgeelhaar/acai/internal/application/annotation"
	authapp "github.com/felixgeelhaar/acai/internal/application/auth"
	embeddingapp "github.com/felixgeelhaar/acai/internal/application/embedding"
	exportapp "github.com/felixgeelhaar/acai/internal/application/export"
	meetingapp "github.com/felixgeelhaar/acai/internal/application/meeting"
	workspaceapp "github.com/felixgeelhaar/acai/internal/application/workspace"
	"github.com/felixgeelhaar/acai/internal/domain/annotation"
	domainauth "github.com/felixgeelhaar/acai/internal/domain/auth"
	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
	"github.com/felixgeelhaar/acai/internal/domain/workspace"
	"github.com/felixgeelhaar/acai/internal/interfaces/cli"
	mcpiface "github.com/felixgeelhaar/acai/internal/interfaces/mcp"
)

func TestRootCmd_HasExpectedSubcommands(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	expected := []string{"auth", "sync", "list", "export", "serve", "workspace", "note", "action", "version"}
	for _, name := range expected {
		found := false
		for _, cmd := range root.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing subcommand %q", name)
		}
	}
}

func TestVersionCmd(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "acai") {
		t.Errorf("version output should contain 'acai': %q", buf.String())
	}
}

func TestAuthLoginCmd(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"auth", "login"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := deps.Out.(*bytes.Buffer).String()
	if !strings.Contains(output, "Authenticated successfully") {
		t.Errorf("expected success message, got: %q", output)
	}
}

func TestAuthStatusCmd(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"auth", "status"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSyncCmd(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"sync"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := deps.Out.(*bytes.Buffer).String()
	if !strings.Contains(output, "Synced") {
		t.Errorf("expected sync output, got: %q", output)
	}
}

func TestExportMeetingCmd(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"export", "meeting", "m-1"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := deps.Out.(*bytes.Buffer).String()
	if !strings.Contains(output, "Sprint Planning") {
		t.Errorf("expected meeting title in output, got: %q", output)
	}
}

func TestExportMeetingCmd_MissingArg(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"export", "meeting"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for missing meeting ID")
	}
}

func TestWorkspaceListCmd(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	root.SetArgs([]string{"workspace", "list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := deps.Out.(*bytes.Buffer).String()
	if !strings.Contains(output, "No workspaces found") {
		t.Errorf("expected empty workspace message, got: %q", output)
	}
}

func TestWorkspaceListCmd_WithWorkspaces(t *testing.T) {
	deps := testDeps(t)

	// Replace workspace repo with one that returns data
	ws1, _ := workspace.New("ws-1", "Engineering", "engineering")
	ws2, _ := workspace.New("ws-2", "Design", "design")
	wsRepo := &mockWorkspaceRepo{workspaces: []*workspace.Workspace{ws1, ws2}}
	deps.ListWorkspaces = workspaceapp.NewListWorkspaces(wsRepo)

	root := cli.NewRootCmd(deps)
	root.SetArgs([]string{"workspace", "list"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := deps.Out.(*bytes.Buffer).String()
	if !strings.Contains(output, "Engineering") {
		t.Errorf("expected workspace name, got: %q", output)
	}
	if !strings.Contains(output, "design") {
		t.Errorf("expected workspace slug, got: %q", output)
	}
}

// --- Test Helpers ---

type mockAuthService struct{}

func (m *mockAuthService) Login(_ context.Context, method domainauth.AuthMethod) (*domainauth.Credential, error) {
	token := domainauth.NewToken("test", "test", time.Now().Add(1*time.Hour).UTC())
	return domainauth.NewCredential(method, token, "test-ws"), nil
}
func (m *mockAuthService) Status(_ context.Context) (*domainauth.Credential, error) {
	return nil, domainauth.ErrNotAuthenticated
}
func (m *mockAuthService) Logout(_ context.Context) error { return nil }

type mockMeetingRepo struct{}

func (m *mockMeetingRepo) FindByID(_ context.Context, id domain.MeetingID) (*domain.Meeting, error) {
	mtg, _ := domain.New(id, "Sprint Planning", time.Now().UTC(), domain.SourceZoom, nil)
	mtg.ClearDomainEvents()
	return mtg, nil
}
func (m *mockMeetingRepo) List(_ context.Context, _ domain.ListFilter) ([]*domain.Meeting, error) {
	return []*domain.Meeting{}, nil
}
func (m *mockMeetingRepo) GetTranscript(_ context.Context, _ domain.MeetingID) (*domain.Transcript, error) {
	return nil, domain.ErrTranscriptNotReady
}
func (m *mockMeetingRepo) SearchTranscripts(_ context.Context, _ string, _ domain.ListFilter) ([]*domain.Meeting, error) {
	return nil, nil
}
func (m *mockMeetingRepo) GetActionItems(_ context.Context, _ domain.MeetingID) ([]*domain.ActionItem, error) {
	ai, _ := domain.NewActionItem("ai-1", "m-1", "Alice", "Review PR", nil)
	return []*domain.ActionItem{ai}, nil
}
func (m *mockMeetingRepo) Sync(_ context.Context, _ *time.Time) ([]domain.DomainEvent, error) {
	return []domain.DomainEvent{}, nil
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

type mockNoteRepo struct {
	notes []*annotation.AgentNote
}

func (m *mockNoteRepo) Save(_ context.Context, note *annotation.AgentNote) error {
	m.notes = append(m.notes, note)
	return nil
}
func (m *mockNoteRepo) FindByID(_ context.Context, id annotation.NoteID) (*annotation.AgentNote, error) {
	for _, n := range m.notes {
		if n.ID() == id {
			return n, nil
		}
	}
	return nil, annotation.ErrNoteNotFound
}
func (m *mockNoteRepo) ListByMeeting(_ context.Context, meetingID string) ([]*annotation.AgentNote, error) {
	var result []*annotation.AgentNote
	for _, n := range m.notes {
		if n.MeetingID() == meetingID {
			result = append(result, n)
		}
	}
	return result, nil
}
func (m *mockNoteRepo) Delete(_ context.Context, id annotation.NoteID) error {
	for i, n := range m.notes {
		if n.ID() == id {
			m.notes = append(m.notes[:i], m.notes[i+1:]...)
			return nil
		}
	}
	return annotation.ErrNoteNotFound
}

type mockWriteRepo struct{}

func (m *mockWriteRepo) SaveActionItemState(_ context.Context, _ *domain.ActionItem) error {
	return nil
}
func (m *mockWriteRepo) GetLocalActionItemState(_ context.Context, _ domain.ActionItemID) (*domain.ActionItem, error) {
	return nil, domain.ErrMeetingNotFound
}

type mockDispatcher struct{}

func (m *mockDispatcher) Dispatch(_ context.Context, _ []domain.DomainEvent) error { return nil }

func testDeps(t *testing.T) *cli.Dependencies {
	t.Helper()
	repo := &mockMeetingRepo{}
	authSvc := &mockAuthService{}
	wsRepo := &mockWorkspaceRepo{}
	noteRepo := &mockNoteRepo{}
	writeRepo := &mockWriteRepo{}
	dispatcher := &mockDispatcher{}
	buf := new(bytes.Buffer)

	return &cli.Dependencies{
		ListMeetings:      meetingapp.NewListMeetings(repo),
		GetMeeting:        meetingapp.NewGetMeeting(repo),
		GetTranscript:     meetingapp.NewGetTranscript(repo),
		SearchTranscripts: meetingapp.NewSearchTranscripts(repo),
		GetActionItems:    meetingapp.NewGetActionItems(repo),
		SyncMeetings:      meetingapp.NewSyncMeetings(repo),
		ExportMeeting:     exportapp.NewExportMeeting(repo),
		Login:             authapp.NewLogin(authSvc),
		CheckStatus:       authapp.NewCheckStatus(authSvc),
		ListWorkspaces:    workspaceapp.NewListWorkspaces(wsRepo),
		GetWorkspace:      workspaceapp.NewGetWorkspace(wsRepo),
		AddNote:           annotationapp.NewAddNote(noteRepo, repo, dispatcher),
		ListNotes:         annotationapp.NewListNotes(noteRepo),
		DeleteNote:        annotationapp.NewDeleteNote(noteRepo, dispatcher),
		CompleteActionItem: meetingapp.NewCompleteActionItem(repo, writeRepo, dispatcher),
		UpdateActionItem:   meetingapp.NewUpdateActionItem(repo, writeRepo, dispatcher),
		ExportEmbeddings:   embeddingapp.NewExportEmbeddings(repo, noteRepo),
		MCPServer: mcpiface.NewServer("acai", "test", mcpiface.ServerOptions{
			ListMeetings:      meetingapp.NewListMeetings(repo),
			GetMeeting:        meetingapp.NewGetMeeting(repo),
			GetTranscript:     meetingapp.NewGetTranscript(repo),
			SearchTranscripts: meetingapp.NewSearchTranscripts(repo),
			GetActionItems:    meetingapp.NewGetActionItems(repo),
			GetMeetingStats:   meetingapp.NewGetMeetingStats(repo),
			ListWorkspaces:    workspaceapp.NewListWorkspaces(wsRepo),
			GetWorkspace:      workspaceapp.NewGetWorkspace(wsRepo),
			AddNote:           annotationapp.NewAddNote(noteRepo, repo, dispatcher),
			ListNotes:         annotationapp.NewListNotes(noteRepo),
			DeleteNote:        annotationapp.NewDeleteNote(noteRepo, dispatcher),
			CompleteActionItem: meetingapp.NewCompleteActionItem(repo, writeRepo, dispatcher),
			UpdateActionItem:   meetingapp.NewUpdateActionItem(repo, writeRepo, dispatcher),
			ExportEmbeddings:   embeddingapp.NewExportEmbeddings(repo, noteRepo),
		}),
		Out: buf,
	}
}
