package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	authapp "github.com/felixgeelhaar/granola-mcp/internal/application/auth"
	meetingapp "github.com/felixgeelhaar/granola-mcp/internal/application/meeting"
	exportapp "github.com/felixgeelhaar/granola-mcp/internal/application/export"
	domainauth "github.com/felixgeelhaar/granola-mcp/internal/domain/auth"
	domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
	"github.com/felixgeelhaar/granola-mcp/internal/interfaces/cli"
	mcpiface "github.com/felixgeelhaar/granola-mcp/internal/interfaces/mcp"
)

func TestRootCmd_HasExpectedSubcommands(t *testing.T) {
	deps := testDeps(t)
	root := cli.NewRootCmd(deps)

	expected := []string{"auth", "sync", "list", "export", "serve", "version"}
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

	if !strings.Contains(buf.String(), "granola-mcp") {
		t.Errorf("version output should contain 'granola-mcp': %q", buf.String())
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
	return nil, nil
}
func (m *mockMeetingRepo) Sync(_ context.Context, _ *time.Time) ([]domain.DomainEvent, error) {
	return []domain.DomainEvent{}, nil
}

func testDeps(t *testing.T) *cli.Dependencies {
	t.Helper()
	repo := &mockMeetingRepo{}
	authSvc := &mockAuthService{}
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
		MCPServer: mcpiface.NewServer(
			"granola-mcp", "test",
			meetingapp.NewListMeetings(repo),
			meetingapp.NewGetMeeting(repo),
			meetingapp.NewGetTranscript(repo),
			meetingapp.NewSearchTranscripts(repo),
			meetingapp.NewGetActionItems(repo),
			meetingapp.NewGetMeetingStats(repo),
		),
		Out: buf,
	}
}
