// Package main is the composition root for acai.
// All dependencies are wired here — no service locator, no global state.
// This is the only place that knows about all layers simultaneously.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	annotationapp "github.com/felixgeelhaar/acai/internal/application/annotation"
	authapp "github.com/felixgeelhaar/acai/internal/application/auth"
	embeddingapp "github.com/felixgeelhaar/acai/internal/application/embedding"
	exportapp "github.com/felixgeelhaar/acai/internal/application/export"
	meetingapp "github.com/felixgeelhaar/acai/internal/application/meeting"
	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
	infraauth "github.com/felixgeelhaar/acai/internal/infrastructure/auth"
	"github.com/felixgeelhaar/acai/internal/infrastructure/cache"
	"github.com/felixgeelhaar/acai/internal/infrastructure/config"
	"github.com/felixgeelhaar/acai/internal/infrastructure/events"
	"github.com/felixgeelhaar/acai/internal/infrastructure/granola"
	"github.com/felixgeelhaar/acai/internal/infrastructure/localstore"
	"github.com/felixgeelhaar/acai/internal/infrastructure/outbox"
	infraPolicy "github.com/felixgeelhaar/acai/internal/infrastructure/policy"
	"github.com/felixgeelhaar/acai/internal/infrastructure/resilience"
	"github.com/felixgeelhaar/acai/internal/interfaces/cli"
	mcpiface "github.com/felixgeelhaar/acai/internal/interfaces/mcp"
	_ "github.com/mattn/go-sqlite3"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Set version info for the CLI
	cli.Version = version
	cli.Commit = commit
	cli.Date = date

	// Load configuration (file defaults + env overrides)
	cfg := config.Load()

	// --- Infrastructure Layer ---

	// HTTP client for Granola API
	httpClient := &http.Client{Timeout: cfg.Resilience.Timeout}

	// Granola API client (anti-corruption layer)
	granolaClient := granola.NewClient(cfg.Granola.APIURL, httpClient, cfg.Granola.APIToken)

	// Repository: Granola API → domain.Repository
	granolaRepo := granola.NewRepository(granolaClient)

	// Resilience decorator (circuit breaker, timeout, retry, rate limit)
	resilientRepo := resilience.NewResilientRepository(granolaRepo, resilience.Config{
		Timeout:          cfg.Resilience.Timeout,
		MaxRetries:       cfg.Resilience.Retry.MaxAttempts,
		RetryDelay:       cfg.Resilience.Retry.InitialDelay,
		RetryMaxDelay:    cfg.Resilience.Retry.MaxDelay,
		FailureThreshold: cfg.Resilience.CircuitBreaker.FailureThreshold,
		SuccessThreshold: cfg.Resilience.CircuitBreaker.SuccessThreshold,
		HalfOpenTimeout:  cfg.Resilience.CircuitBreaker.HalfOpenTimeout,
		RateLimit:        cfg.Resilience.RateLimit.Rate,
		RateBurst:        cfg.Resilience.RateLimit.Rate * 2,
		RateInterval:     cfg.Resilience.RateLimit.Interval,
	})
	defer func() { _ = resilientRepo.Close() }()

	// Cache decorator (SQLite local cache)
	var repo domain.Repository = resilientRepo
	if cfg.Cache.Enabled {
		cacheDir := cfg.Cache.Dir
		if err := os.MkdirAll(cacheDir, 0o700); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: cannot create cache dir: %v\n", err)
		} else {
			dbPath := filepath.Join(cacheDir, "cache.db")
			db, err := sql.Open("sqlite3", dbPath)
			if err == nil {
				cachedRepo, cacheErr := cache.NewCachedRepository(resilientRepo, db, cfg.Cache.TTL)
				if cacheErr == nil {
					repo = cachedRepo
					defer func() { _ = db.Close() }()
				}
			}
		}
	}

	// Auth infrastructure
	homeDir, err := os.UserHomeDir()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: cannot determine home directory: %v, using temp dir\n", err)
		homeDir = os.TempDir()
	}
	tokenStore := infraauth.NewFileTokenStore(homeDir + "/.acai")
	authService := infraauth.NewService(tokenStore)

	// If we have a stored token, set it on the Granola client
	if cred, err := authService.Status(context.Background()); err == nil && cred.IsValid() {
		granolaClient.SetToken(cred.Token().AccessToken())
	}

	// Local store (SQLite for write-side: notes, action item overrides, outbox)
	localDir := cfg.Cache.Dir // Reuse cache dir for local store
	if err := os.MkdirAll(localDir, 0o700); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: cannot create local store dir: %v\n", err)
	}
	localDBPath := filepath.Join(localDir, "local.db")
	localDB, err := sql.Open("sqlite3", localDBPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: cannot open local store: %v\n", err)
		localDB = nil
	} else {
		defer func() { _ = localDB.Close() }()
		if err := localstore.InitSchema(localDB); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: cannot init local store schema: %v\n", err)
		}
	}

	// Local store repositories (guarded against nil db)
	var noteRepo *localstore.NoteRepository
	var writeRepo *localstore.WriteRepository
	if localDB != nil {
		noteRepo = localstore.NewNoteRepository(localDB)
		writeRepo = localstore.NewWriteRepository(localDB)
	}

	// Event infrastructure: notifier → dispatcher → outbox decorator
	notifier := events.NewMCPNotifier()
	innerDispatcher := events.NewDispatcher(notifier)
	var dispatcher domain.EventDispatcher = innerDispatcher
	if localDB != nil {
		outboxStore := outbox.NewSQLiteStore(localDB)
		dispatcher = outbox.NewDispatcher(innerDispatcher, outboxStore)
	}

	// --- Application Layer (Use Cases) ---

	listMeetings := meetingapp.NewListMeetings(repo)
	getMeeting := meetingapp.NewGetMeeting(repo)
	getTranscript := meetingapp.NewGetTranscript(repo)
	searchTranscripts := meetingapp.NewSearchTranscripts(repo)
	getActionItems := meetingapp.NewGetActionItems(repo)
	getMeetingStats := meetingapp.NewGetMeetingStats(repo)
	syncMeetings := meetingapp.NewSyncMeetings(repo)
	exportMeeting := exportapp.NewExportMeeting(repo)
	login := authapp.NewLogin(authService)
	checkStatus := authapp.NewCheckStatus(authService)
	logout := authapp.NewLogout(authService)

	// Write use cases (require local DB)
	var addNote *annotationapp.AddNote
	var listNotes *annotationapp.ListNotes
	var deleteNote *annotationapp.DeleteNote
	var completeActionItem *meetingapp.CompleteActionItem
	var updateActionItem *meetingapp.UpdateActionItem
	var exportEmbeddings *embeddingapp.ExportEmbeddings
	if localDB != nil {
		addNote = annotationapp.NewAddNote(noteRepo, repo, dispatcher)
		listNotes = annotationapp.NewListNotes(noteRepo)
		deleteNote = annotationapp.NewDeleteNote(noteRepo, dispatcher)
		completeActionItem = meetingapp.NewCompleteActionItem(repo, writeRepo, dispatcher)
		updateActionItem = meetingapp.NewUpdateActionItem(repo, writeRepo, dispatcher)
		exportEmbeddings = embeddingapp.NewExportEmbeddings(repo, noteRepo)
	}

	// --- Interfaces Layer ---

	// Load policy engine (optional)
	var policyEngine *infraPolicy.Engine
	if cfg.Policy.Enabled && cfg.Policy.FilePath != "" {
		loadResult, policyErr := infraPolicy.LoadFromFile(cfg.Policy.FilePath)
		if policyErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: cannot load policy file: %v\n", policyErr)
		} else {
			policyEngine = infraPolicy.NewEngine(loadResult)
		}
	}

	// MCP server
	mcpServer := mcpiface.NewServer(cfg.MCP.ServerName, version, mcpiface.ServerOptions{
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
		PolicyEngine:       policyEngine,
	})

	// CLI dependencies
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
		AddNote:            addNote,
		ListNotes:          listNotes,
		DeleteNote:         deleteNote,
		CompleteActionItem: completeActionItem,
		UpdateActionItem:   updateActionItem,
		ExportEmbeddings:   exportEmbeddings,
		GranolaAPIToken:    cfg.Granola.APIToken,
		Out:                os.Stdout,
	}

	// Execute CLI
	if err := cli.NewRootCmd(deps).Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
