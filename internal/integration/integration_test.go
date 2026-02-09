//go:build integration

// Package integration_test provides integration tests that exercise the full
// dependency graph: interface → application → infrastructure → mock HTTP server.
// Only the external HTTP boundary is mocked via httptest.NewServer.
package integration_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	annotationapp "github.com/felixgeelhaar/acai/internal/application/annotation"
	authapp "github.com/felixgeelhaar/acai/internal/application/auth"
	embeddingapp "github.com/felixgeelhaar/acai/internal/application/embedding"
	exportapp "github.com/felixgeelhaar/acai/internal/application/export"
	meetingapp "github.com/felixgeelhaar/acai/internal/application/meeting"
	domainauth "github.com/felixgeelhaar/acai/internal/domain/auth"
	"github.com/felixgeelhaar/acai/internal/infrastructure/events"
	"github.com/felixgeelhaar/acai/internal/infrastructure/granola"
	"github.com/felixgeelhaar/acai/internal/infrastructure/localstore"
	"github.com/felixgeelhaar/acai/internal/interfaces/cli"
	mcpiface "github.com/felixgeelhaar/acai/internal/interfaces/mcp"
)

const testToken = "test-token-abc123"

// --- API Fixtures ---

type APIFixtures struct {
	Notes       []granola.NoteListItem
	NoteDetails map[string]granola.NoteDetailResponse
	HasMore     bool
	Cursor      string
}

func (f *APIFixtures) ListNotesResponse(cursor string) granola.NoteListResponse {
	return granola.NoteListResponse{
		Notes:   f.Notes,
		HasMore: f.HasMore,
		Cursor:  f.Cursor,
	}
}

func (f *APIFixtures) GetNoteResponse(id string, includeTranscript bool) (*granola.NoteDetailResponse, bool) {
	detail, ok := f.NoteDetails[id]
	if !ok {
		return nil, false
	}
	if !includeTranscript {
		// Return a copy without transcript data
		noTranscript := detail
		noTranscript.Transcript = nil
		return &noTranscript, true
	}
	return &detail, true
}

// --- Default Fixtures ---

func defaultFixtures() *APIFixtures {
	now := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	mdSummary := "## Sprint Planning\n\nWe discussed the following items..."

	return &APIFixtures{
		Notes: []granola.NoteListItem{
			{
				ID:     "m-1",
				Object: "note",
				Title:  "Sprint Planning",
				Owner:  granola.UserDTO{Name: "Alice Johnson", Email: "alice@example.com"},
				CreatedAt: now,
			},
			{
				ID:     "m-2",
				Object: "note",
				Title:  "Retrospective",
				Owner:  granola.UserDTO{Name: "Bob Smith", Email: "bob@example.com"},
				CreatedAt: now.Add(-24 * time.Hour),
			},
			{
				ID:     "m-3",
				Object: "note",
				Title:  "1:1 with Manager",
				Owner:  granola.UserDTO{Name: "Carol Davis", Email: "carol@example.com"},
				CreatedAt: now.Add(-48 * time.Hour),
			},
		},
		NoteDetails: map[string]granola.NoteDetailResponse{
			"m-1": {
				ID:        "m-1",
				Object:    "note",
				Title:     "Sprint Planning",
				Owner:     granola.UserDTO{Name: "Alice Johnson", Email: "alice@example.com"},
				CreatedAt: now,
				Attendees: []granola.UserDTO{
					{Name: "Bob Smith", Email: "bob@example.com"},
					{Name: "Carol Davis", Email: "carol@example.com"},
				},
				SummaryText:     "We discussed the following items",
				SummaryMarkdown: &mdSummary,
				Transcript: []granola.TranscriptItemDTO{
					{Speaker: "Alice Johnson", Text: "Let's start the sprint planning.", Timestamp: now},
					{Speaker: "Bob Smith", Text: "I'll take the login feature.", Timestamp: now.Add(30 * time.Second)},
					{Speaker: "Carol Davis", Text: "I'll work on the API.", Timestamp: now.Add(60 * time.Second)},
				},
			},
			"m-2": {
				ID:          "m-2",
				Object:      "note",
				Title:       "Retrospective",
				Owner:       granola.UserDTO{Name: "Bob Smith", Email: "bob@example.com"},
				CreatedAt:   now.Add(-24 * time.Hour),
				SummaryText: "Retro summary: went well, could improve",
			},
			"m-3": {
				ID:        "m-3",
				Object:    "note",
				Title:     "1:1 with Manager",
				Owner:     granola.UserDTO{Name: "Carol Davis", Email: "carol@example.com"},
				CreatedAt: now.Add(-48 * time.Hour),
			},
		},
	}
}

// --- Fake Granola API Server ---

func fakeGranolaAPI(t *testing.T, fixtures *APIFixtures) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+testToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/notes":
			cursor := r.URL.Query().Get("cursor")
			resp := fixtures.ListNotesResponse(cursor)
			json.NewEncoder(w).Encode(resp)

		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/notes/"):
			id := strings.TrimPrefix(r.URL.Path, "/v1/notes/")
			includeTranscript := r.URL.Query().Get("include") == "transcript"
			detail, ok := fixtures.GetNoteResponse(id, includeTranscript)
			if !ok {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(detail)

		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
}

// --- Test Auth Service (minimal mock) ---

type testAuthService struct{}

func (s *testAuthService) Login(_ context.Context, params domainauth.LoginParams) (*domainauth.Credential, error) {
	token := domainauth.NewToken(params.APIToken, "", time.Now().Add(365*24*time.Hour))
	return domainauth.NewCredential(params.Method, token, "test-workspace"), nil
}

func (s *testAuthService) Status(_ context.Context) (*domainauth.Credential, error) {
	return nil, domainauth.ErrNotAuthenticated
}

func (s *testAuthService) Logout(_ context.Context) error {
	return nil
}

// --- Test Environment ---

type TestEnv struct {
	MCPServer  *mcpiface.Server
	CLIDeps    *cli.Dependencies
	Fixtures   *APIFixtures
	APIServer  *httptest.Server
	DB         *sql.DB
	Output     *bytes.Buffer
}

func newTestEnv(t *testing.T, fixtures *APIFixtures) *TestEnv {
	t.Helper()

	if fixtures == nil {
		fixtures = defaultFixtures()
	}

	// 1. Fake HTTP server
	apiServer := fakeGranolaAPI(t, fixtures)
	t.Cleanup(apiServer.Close)

	// 2. Granola client + repository (no resilience/cache decorators)
	granolaClient := granola.NewClient(apiServer.URL, apiServer.Client(), testToken)
	repo := granola.NewRepository(granolaClient)

	// 3. In-memory SQLite
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := localstore.InitSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	noteRepo := localstore.NewNoteRepository(db)
	writeRepo := localstore.NewWriteRepository(db)

	// 4. Events infrastructure
	notifier := events.NewMCPNotifier()
	dispatcher := events.NewDispatcher(notifier)

	// 5. Auth service (minimal mock)
	authSvc := &testAuthService{}

	// 6. Application use cases
	listMeetings := meetingapp.NewListMeetings(repo)
	getMeeting := meetingapp.NewGetMeeting(repo)
	getTranscript := meetingapp.NewGetTranscript(repo)
	searchTranscripts := meetingapp.NewSearchTranscripts(repo)
	getActionItems := meetingapp.NewGetActionItems(repo)
	getMeetingStats := meetingapp.NewGetMeetingStats(repo)
	syncMeetings := meetingapp.NewSyncMeetings(repo)
	exportMeeting := exportapp.NewExportMeeting(repo)

	addNote := annotationapp.NewAddNote(noteRepo, repo, dispatcher)
	listNotes := annotationapp.NewListNotes(noteRepo)
	deleteNote := annotationapp.NewDeleteNote(noteRepo, dispatcher)
	completeActionItem := meetingapp.NewCompleteActionItem(repo, writeRepo, dispatcher)
	updateActionItem := meetingapp.NewUpdateActionItem(repo, writeRepo, dispatcher)
	exportEmbeddings := embeddingapp.NewExportEmbeddings(repo, noteRepo)

	login := authapp.NewLogin(authSvc)
	checkStatus := authapp.NewCheckStatus(authSvc)
	logout := authapp.NewLogout(authSvc)

	// 7. MCP server
	mcpOpts := mcpiface.ServerOptions{
		ListMeetings:       listMeetings,
		GetMeeting:         getMeeting,
		GetTranscript:      getTranscript,
		SearchTranscripts:  searchTranscripts,
		GetActionItems:     getActionItems,
		GetMeetingStats:    getMeetingStats,
		AddNote:            addNote,
		ListNotes:          listNotes,
		DeleteNote:         deleteNote,
		CompleteActionItem: completeActionItem,
		UpdateActionItem:   updateActionItem,
		ExportEmbeddings:   exportEmbeddings,
	}
	mcpServer := mcpiface.NewServer("acai", "test", mcpOpts)

	// 8. CLI dependencies
	buf := new(bytes.Buffer)
	deps := &cli.Dependencies{
		ListMeetings:       listMeetings,
		GetMeeting:         getMeeting,
		GetTranscript:      getTranscript,
		SearchTranscripts:  searchTranscripts,
		GetActionItems:     getActionItems,
		SyncMeetings:       syncMeetings,
		ExportMeeting:      exportMeeting,
		Login:              login,
		CheckStatus:        checkStatus,
		Logout:             logout,
		EventDispatcher:    dispatcher,
		MCPServer:          mcpServer,
		Out:                buf,
		AddNote:            addNote,
		ListNotes:          listNotes,
		DeleteNote:         deleteNote,
		CompleteActionItem: completeActionItem,
		UpdateActionItem:   updateActionItem,
		ExportEmbeddings:   exportEmbeddings,
		GranolaAPIToken:    testToken,
	}

	return &TestEnv{
		MCPServer: mcpServer,
		CLIDeps:   deps,
		Fixtures:  fixtures,
		APIServer: apiServer,
		DB:        db,
		Output:    buf,
	}
}

// resetOutput clears the output buffer between sequential CLI commands.
func (e *TestEnv) resetOutput() {
	e.Output.Reset()
}

// runCLI creates and executes a CLI command with the given args.
func (e *TestEnv) runCLI(args ...string) error {
	cmd := cli.NewRootCmd(e.CLIDeps)
	cmd.SetOut(e.Output)
	cmd.SetErr(e.Output)
	cmd.SetArgs(args)
	return cmd.Execute()
}

// outputString returns the current output buffer contents.
func (e *TestEnv) outputString() string {
	return e.Output.String()
}

// callToolJSON is a convenience wrapper for HandleToolJSON.
func (e *TestEnv) callToolJSON(t *testing.T, tool string, input interface{}) (json.RawMessage, error) {
	t.Helper()
	rawInput, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal tool input: %v", err)
	}
	return e.MCPServer.HandleToolJSON(context.Background(), tool, rawInput)
}

// mustCallToolJSON calls a tool and fails the test on error.
func (e *TestEnv) mustCallToolJSON(t *testing.T, tool string, input interface{}) json.RawMessage {
	t.Helper()
	result, err := e.callToolJSON(t, tool, input)
	if err != nil {
		t.Fatalf("tool %s failed: %v", tool, err)
	}
	return result
}

// unmarshalResult unmarshals a JSON result into the target.
func unmarshalResult(t *testing.T, raw json.RawMessage, target interface{}) {
	t.Helper()
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatalf("unmarshal result: %v (raw: %s)", err, string(raw))
	}
}

// assertContains checks that s contains substr.
func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected output to contain %q, got:\n%s", substr, truncate(s, 500))
	}
}

// assertNotContains checks that s does NOT contain substr.
func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("expected output NOT to contain %q, got:\n%s", substr, truncate(s, 500))
	}
}

// assertError checks that err is not nil.
func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// assertNoError checks that err is nil.
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + fmt.Sprintf("... (%d more bytes)", len(s)-max)
	}
	return s
}
