# Technical Design Document

## Granola MCP Connector

### 1. Overview

This document describes the technical design for `granola-mcp`, a Go CLI and MCP server that exposes Granola meeting intelligence as MCP resources and tools. The system follows a strict Domain-Driven Design (DDD) architecture, leveraging `felixgeelhaar/mcp-go` for the MCP protocol layer, `felixgeelhaar/fortify` for resilience, and `felixgeelhaar/bolt` for structured logging.

### 2. Architecture

#### 2.1 DDD Layer Structure

```
granola-mcp/
├── cmd/
│   └── granola-mcp/
│       └── main.go                    # Entry point, wires everything together
├── internal/
│   ├── domain/                        # Domain Layer (pure business logic, zero dependencies)
│   │   ├── meeting/
│   │   │   ├── meeting.go             # Meeting aggregate root
│   │   │   ├── transcript.go          # Transcript value object
│   │   │   ├── summary.go            # Summary value object
│   │   │   ├── action_item.go        # ActionItem entity
│   │   │   ├── participant.go        # Participant value object
│   │   │   ├── metadata.go           # Metadata value object
│   │   │   ├── events.go             # Domain events
│   │   │   └── repository.go         # Repository port (interface)
│   │   ├── workspace/
│   │   │   ├── workspace.go          # Workspace aggregate root
│   │   │   └── repository.go         # Repository port
│   │   └── auth/
│   │       ├── token.go              # Token value object
│   │       ├── credential.go         # Credential entity
│   │       └── service.go            # Auth domain service port
│   ├── application/                   # Application Layer (use cases, orchestration)
│   │   ├── meeting/
│   │   │   ├── list_meetings.go      # ListMeetings use case
│   │   │   ├── get_meeting.go        # GetMeeting use case
│   │   │   ├── get_transcript.go     # GetTranscript use case
│   │   │   ├── search_transcripts.go # SearchTranscripts use case
│   │   │   ├── get_action_items.go   # GetActionItems use case
│   │   │   └── sync_meetings.go      # SyncMeetings use case
│   │   ├── auth/
│   │   │   ├── login.go             # Login use case
│   │   │   └── check_status.go      # CheckStatus use case
│   │   └── export/
│   │       └── export_meeting.go    # ExportMeeting use case
│   ├── infrastructure/                # Infrastructure Layer (adapters, implementations)
│   │   ├── granola/
│   │   │   ├── client.go            # Granola API HTTP client
│   │   │   ├── client_test.go
│   │   │   ├── models.go            # API response DTOs
│   │   │   ├── mapper.go            # DTO → Domain model mapper
│   │   │   └── repository.go        # MeetingRepository implementation
│   │   ├── cache/
│   │   │   ├── sqlite.go            # SQLite cache adapter
│   │   │   └── repository.go        # Cached repository decorator
│   │   ├── auth/
│   │   │   ├── oauth.go             # OAuth flow implementation
│   │   │   ├── token_store.go       # Token persistence (keychain/file)
│   │   │   └── service.go           # AuthService implementation
│   │   ├── config/
│   │   │   ├── config.go            # Configuration loading
│   │   │   └── defaults.go          # Default configuration values
│   │   └── resilience/
│   │       ├── middleware.go         # Fortify middleware composition
│   │       └── factory.go           # Resilience pattern factory
│   └── interfaces/                    # Interface Layer (adapters to the outside world)
│       ├── mcp/
│       │   ├── server.go            # MCP server setup (mcp-go)
│       │   ├── resources.go         # MCP resource handlers
│       │   ├── tools.go             # MCP tool handlers
│       │   ├── events.go            # MCP event publishers
│       │   └── middleware.go        # MCP middleware (logging, auth, timeout)
│       └── cli/
│           ├── root.go              # Root command
│           ├── auth.go              # auth login / auth status
│           ├── sync.go              # sync command
│           ├── list.go              # list meetings
│           ├── export.go            # export meeting <id>
│           ├── serve.go             # serve (starts MCP server)
│           └── formatter.go         # Output formatting (table, JSON, markdown)
├── docs/
│   ├── prd.md
│   ├── vision.md
│   └── tdd.md
├── go.mod
├── go.sum
├── Makefile
├── goreleaser.yaml                   # GoReleaser config for Homebrew
└── README.md
```

#### 2.2 Dependency Rule

Dependencies flow strictly inward:

```
interfaces → application → domain
infrastructure → domain (implements ports)
```

- **Domain** has zero imports from other layers. No framework, no HTTP, no database.
- **Application** depends on domain only. Uses domain ports (interfaces) for external concerns.
- **Infrastructure** implements domain ports. Depends on domain types and external libraries.
- **Interfaces** adapts the outside world (CLI, MCP protocol) to application use cases.

```
┌─────────────────────────────────────────────────────────────┐
│                    Interfaces Layer                          │
│  ┌───────────────────┐  ┌────────────────────────────────┐ │
│  │   CLI (cobra)      │  │   MCP Server (mcp-go)          │ │
│  │  auth, sync, list, │  │  resources, tools, events,     │ │
│  │  export, serve     │  │  middleware                     │ │
│  └────────┬──────────┘  └──────────────┬─────────────────┘ │
├───────────┼────────────────────────────┼───────────────────-┤
│           ▼           Application Layer ▼                    │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Use Cases: ListMeetings, GetMeeting, GetTranscript, │  │
│  │  SearchTranscripts, GetActionItems, SyncMeetings,     │  │
│  │  Login, CheckStatus, ExportMeeting                    │  │
│  └──────────────────────────┬───────────────────────────┘  │
├─────────────────────────────┼──────────────────────────────-┤
│                             ▼    Domain Layer                │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Aggregates: Meeting, Workspace                       │  │
│  │  Entities: ActionItem, Credential                     │  │
│  │  Value Objects: Transcript, Summary, Participant,     │  │
│  │                 Metadata, Token                        │  │
│  │  Ports: MeetingRepository, WorkspaceRepository,       │  │
│  │         AuthService                                    │  │
│  │  Domain Events: MeetingCreated, TranscriptUpdated,    │  │
│  │                 SummaryUpdated                         │  │
│  └──────────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                    Infrastructure Layer                       │
│  ┌──────────┐ ┌────────┐ ┌──────────┐ ┌────────────────┐  │
│  │ Granola  │ │ Cache  │ │  Auth    │ │  Resilience    │  │
│  │ API      │ │ SQLite │ │  OAuth   │ │  Fortify       │  │
│  │ Client   │ │        │ │  Store   │ │  Composition   │  │
│  └──────────┘ └────────┘ └──────────┘ └────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 3. Domain Model

#### 3.1 Meeting Aggregate Root

```go
package meeting

import "time"

// MeetingID is a strongly-typed identifier for meetings.
type MeetingID string

// Meeting is the aggregate root for meeting data.
type Meeting struct {
    id           MeetingID
    title        string
    datetime     time.Time
    source       Source
    participants []Participant
    transcript   *Transcript
    summary      *Summary
    actionItems  []ActionItem
    metadata     Metadata
    createdAt    time.Time
    updatedAt    time.Time
}

// Source represents the meeting platform origin.
type Source string

const (
    SourceZoom   Source = "zoom"
    SourceMeet   Source = "google_meet"
    SourceTeams  Source = "teams"
    SourceOther  Source = "other"
)
```

#### 3.2 Value Objects

```go
// Participant is an immutable value object representing a meeting attendee.
type Participant struct {
    name  string
    email string
    role  ParticipantRole
}

// Transcript is an immutable value object containing ordered utterances.
type Transcript struct {
    meetingID  MeetingID
    utterances []Utterance
}

type Utterance struct {
    speaker    string
    text       string
    timestamp  time.Time
    confidence float64
}

// Summary is an immutable value object for meeting summaries.
type Summary struct {
    meetingID MeetingID
    content   string
    kind      SummaryKind
}

type SummaryKind string

const (
    SummaryAuto   SummaryKind = "auto"
    SummaryEdited SummaryKind = "user_edited"
)

// Metadata is an immutable value object for extensible meeting metadata.
type Metadata struct {
    tags         []string
    links        []string
    externalRefs map[string]string
}
```

#### 3.3 ActionItem Entity

```go
// ActionItemID is a strongly-typed identifier for action items.
type ActionItemID string

// ActionItem is an entity with identity, belonging to a Meeting aggregate.
type ActionItem struct {
    id        ActionItemID
    meetingID MeetingID
    owner     string
    text      string
    dueDate   *time.Time
    completed bool
}
```

#### 3.4 Domain Events

```go
// DomainEvent is the base interface for all domain events.
type DomainEvent interface {
    EventName() string
    OccurredAt() time.Time
}

type MeetingCreated struct {
    MeetingID MeetingID
    Title     string
    Datetime  time.Time
    occurred  time.Time
}

type TranscriptUpdated struct {
    MeetingID      MeetingID
    UtteranceCount int
    occurred       time.Time
}

type SummaryUpdated struct {
    MeetingID MeetingID
    Kind      SummaryKind
    occurred  time.Time
}
```

#### 3.5 Repository Ports

```go
package meeting

import "context"

// ListFilter defines criteria for querying meetings.
type ListFilter struct {
    Since        *time.Time
    Until        *time.Time
    Source       *Source
    Participant  *string
    Query        *string
    Limit        int
    Offset       int
}

// Repository is the port for meeting persistence.
type Repository interface {
    FindByID(ctx context.Context, id MeetingID) (*Meeting, error)
    List(ctx context.Context, filter ListFilter) ([]*Meeting, error)
    GetTranscript(ctx context.Context, id MeetingID) (*Transcript, error)
    SearchTranscripts(ctx context.Context, query string, filter ListFilter) ([]*Meeting, error)
    GetActionItems(ctx context.Context, id MeetingID) ([]ActionItem, error)
    Sync(ctx context.Context, since *time.Time) ([]DomainEvent, error)
}
```

### 4. Application Layer

#### 4.1 Use Case Pattern

Each use case is a single-purpose struct with an `Execute` method. Dependencies are injected via constructor.

```go
package meeting

import (
    "context"

    domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

// ListMeetingsInput defines the input for the ListMeetings use case.
type ListMeetingsInput struct {
    Since       *time.Time
    Until       *time.Time
    Source      *string
    Participant *string
    Query       *string
    Limit       int
    Offset      int
}

// ListMeetingsOutput defines the output for the ListMeetings use case.
type ListMeetingsOutput struct {
    Meetings []*domain.Meeting
    Total    int
}

// ListMeetings orchestrates meeting list retrieval.
type ListMeetings struct {
    repo domain.Repository
}

func NewListMeetings(repo domain.Repository) *ListMeetings {
    return &ListMeetings{repo: repo}
}

func (uc *ListMeetings) Execute(ctx context.Context, input ListMeetingsInput) (*ListMeetingsOutput, error) {
    filter := domain.ListFilter{
        Since:       input.Since,
        Until:       input.Until,
        Limit:       input.Limit,
        Offset:      input.Offset,
    }

    if input.Source != nil {
        src := domain.Source(*input.Source)
        filter.Source = &src
    }
    filter.Participant = input.Participant
    filter.Query = input.Query

    meetings, err := uc.repo.List(ctx, filter)
    if err != nil {
        return nil, err
    }

    return &ListMeetingsOutput{
        Meetings: meetings,
        Total:    len(meetings),
    }, nil
}
```

#### 4.2 Use Case Inventory

| Use Case | Input | Output | Description |
|----------|-------|--------|-------------|
| `ListMeetings` | Filters (since, until, source, participant, query, pagination) | Meeting list + total | Paginated, filtered meeting retrieval |
| `GetMeeting` | MeetingID | Full meeting with transcript, summary, action items | Single meeting with all related data |
| `GetTranscript` | MeetingID, optional time range | Transcript with utterances | Transcript retrieval, optionally scoped to a time range |
| `SearchTranscripts` | Query string, filters | Matching meetings with highlighted excerpts | Full-text search across all transcripts |
| `GetActionItems` | MeetingID (optional) | Action items list | Action items for a meeting or across all meetings |
| `SyncMeetings` | Since timestamp (optional) | Domain events (created, updated) | Pull new/updated meetings from Granola API |
| `Login` | Auth method (OAuth / API token) | Credential | Authenticate with Granola |
| `CheckStatus` | None | Auth status, token expiry, workspace info | Verify current auth state |
| `ExportMeeting` | MeetingID, format (JSON, Markdown, text) | Formatted output | Export meeting data in specified format |

### 5. Infrastructure Layer

#### 5.1 Granola API Client

The HTTP client is the primary adapter implementing `domain.Repository`. It translates between Granola's REST API and domain types.

```go
package granola

import (
    "context"
    "net/http"

    domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
    "github.com/felixgeelhaar/bolt/v3"
)

// Client wraps the Granola REST API.
type Client struct {
    baseURL    string
    httpClient *http.Client
    logger     *bolt.Logger
}

func NewClient(baseURL string, httpClient *http.Client, logger *bolt.Logger) *Client {
    return &Client{
        baseURL:    baseURL,
        httpClient: httpClient,
        logger:     logger,
    }
}
```

#### 5.2 Resilience Composition

Fortify patterns are composed as a decorator around the Granola API client, not embedded inside it. This keeps the client clean and the resilience strategy configurable.

```go
package resilience

import (
    "time"

    "github.com/felixgeelhaar/fortify/circuitbreaker"
    "github.com/felixgeelhaar/fortify/ratelimit"
    "github.com/felixgeelhaar/fortify/retry"
    "github.com/felixgeelhaar/fortify/timeout"
)

// Config defines the resilience configuration for Granola API calls.
type Config struct {
    CircuitBreaker CircuitBreakerConfig
    RateLimit      RateLimitConfig
    Retry          RetryConfig
    Timeout        time.Duration
}

type CircuitBreakerConfig struct {
    FailureThreshold int
    SuccessThreshold int
    HalfOpenTimeout  time.Duration
}

type RateLimitConfig struct {
    Rate     int
    Interval time.Duration
}

type RetryConfig struct {
    MaxAttempts   int
    BackoffPolicy string // "exponential", "linear", "constant"
    InitialDelay  time.Duration
    MaxDelay      time.Duration
}

// DefaultConfig returns production-safe defaults.
func DefaultConfig() Config {
    return Config{
        CircuitBreaker: CircuitBreakerConfig{
            FailureThreshold: 5,
            SuccessThreshold: 2,
            HalfOpenTimeout:  30 * time.Second,
        },
        RateLimit: RateLimitConfig{
            Rate:     100,
            Interval: time.Minute,
        },
        Retry: RetryConfig{
            MaxAttempts:   3,
            BackoffPolicy: "exponential",
            InitialDelay:  500 * time.Millisecond,
            MaxDelay:      10 * time.Second,
        },
        Timeout: 30 * time.Second,
    }
}
```

#### 5.3 Cache Layer

The cache is implemented as a **repository decorator** — it wraps the Granola repository, checking local SQLite first and falling through to the API on cache miss.

```go
package cache

import (
    "context"
    "database/sql"
    "time"

    domain "github.com/felixgeelhaar/granola-mcp/internal/domain/meeting"
)

// CachedRepository decorates a domain.Repository with local SQLite caching.
type CachedRepository struct {
    inner domain.Repository
    db    *sql.DB
    ttl   time.Duration
}

func NewCachedRepository(inner domain.Repository, db *sql.DB, ttl time.Duration) *CachedRepository {
    return &CachedRepository{inner: inner, db: db, ttl: ttl}
}

func (r *CachedRepository) FindByID(ctx context.Context, id domain.MeetingID) (*domain.Meeting, error) {
    // 1. Check cache
    // 2. On miss → delegate to inner.FindByID
    // 3. Store result in cache
    // 4. Return
}
```

#### 5.4 Authentication

```go
package auth

import (
    "context"

    domain "github.com/felixgeelhaar/granola-mcp/internal/domain/auth"
)

// OAuthService implements domain.AuthService using WorkOS OAuth.
type OAuthService struct {
    clientID     string
    redirectURI  string
    tokenStore   TokenStore
}

// TokenStore abstracts token persistence.
type TokenStore interface {
    Save(ctx context.Context, cred domain.Credential) error
    Load(ctx context.Context) (*domain.Credential, error)
    Delete(ctx context.Context) error
}
```

### 6. Interfaces Layer

#### 6.1 MCP Server

The MCP server is built with `mcp-go` and exposes domain data as MCP resources and tools.

```go
package mcp

import (
    "context"

    mcpfw "github.com/felixgeelhaar/mcp-go"
    meetingapp "github.com/felixgeelhaar/granola-mcp/internal/application/meeting"
)

func NewServer(
    listMeetings *meetingapp.ListMeetings,
    getMeeting *meetingapp.GetMeeting,
    getTranscript *meetingapp.GetTranscript,
    searchTranscripts *meetingapp.SearchTranscripts,
    getActionItems *meetingapp.GetActionItems,
) *mcpfw.Server {
    srv := mcpfw.NewServer(mcpfw.ServerInfo{
        Name:    "granola-mcp",
        Version: "0.1.0",
    })

    registerResources(srv, getMeeting, getTranscript)
    registerTools(srv, listMeetings, getMeeting, getTranscript, searchTranscripts, getActionItems)

    return srv
}
```

##### 6.1.1 MCP Resources

```go
func registerResources(srv *mcpfw.Server, getMeeting *meetingapp.GetMeeting, getTranscript *meetingapp.GetTranscript) {
    // meeting://{id} — Full meeting details
    srv.Resource("meeting://{id}").
        Description("A Granola meeting with metadata, participants, and source").
        Handler(meetingResourceHandler(getMeeting))

    // transcript://{meeting_id} — Meeting transcript
    srv.Resource("transcript://{meeting_id}").
        Description("Ordered transcript utterances for a meeting").
        Handler(transcriptResourceHandler(getTranscript))
}
```

##### 6.1.2 MCP Tools

| Tool | Input Schema | Description |
|------|-------------|-------------|
| `list_meetings` | `{ since?: string, until?: string, source?: string, participant?: string, query?: string, limit?: int, offset?: int }` | Search and filter meetings |
| `get_meeting` | `{ id: string }` | Get full meeting details |
| `get_transcript` | `{ meeting_id: string, from?: string, to?: string }` | Get transcript, optionally scoped to time range |
| `search_transcripts` | `{ query: string, since?: string, until?: string, limit?: int }` | Full-text search across transcripts |
| `get_action_items` | `{ meeting_id?: string, owner?: string, completed?: bool }` | Get action items, optionally filtered |

```go
type ListMeetingsInput struct {
    Since       *string `json:"since,omitempty"       description:"ISO 8601 date filter (from)"`
    Until       *string `json:"until,omitempty"       description:"ISO 8601 date filter (to)"`
    Source      *string `json:"source,omitempty"      description:"Meeting platform (zoom, google_meet, teams)"`
    Participant *string `json:"participant,omitempty" description:"Filter by participant name or email"`
    Query       *string `json:"query,omitempty"       description:"Free-text search in titles"`
    Limit       *int    `json:"limit,omitempty"       description:"Max results (default 20)"`
    Offset      *int    `json:"offset,omitempty"      description:"Pagination offset"`
}

func registerTools(srv *mcpfw.Server, /* use cases */) {
    srv.Tool("list_meetings").
        Description("Search and filter Granola meetings").
        Handler(func(ctx context.Context, input ListMeetingsInput) ([]*MeetingResult, error) {
            // Map input → use case input → execute → map output
        })

    srv.Tool("get_meeting").
        Description("Get full details for a specific meeting").
        Handler(func(ctx context.Context, input GetMeetingInput) (*MeetingDetail, error) {
            // ...
        })

    srv.Tool("get_transcript").
        Description("Get the transcript for a meeting").
        Handler(func(ctx context.Context, input GetTranscriptInput) (*TranscriptResult, error) {
            // ...
        })

    srv.Tool("search_transcripts").
        Description("Full-text search across all meeting transcripts").
        Handler(func(ctx context.Context, input SearchTranscriptsInput) ([]*SearchResult, error) {
            // ...
        })

    srv.Tool("get_action_items").
        Description("Get action items from meetings").
        Handler(func(ctx context.Context, input GetActionItemsInput) ([]*ActionItemResult, error) {
            // ...
        })
}
```

##### 6.1.3 MCP Middleware Stack

```go
func applyMiddleware(srv *mcpfw.Server, logger *bolt.Logger) {
    srv.Use(
        middleware.RequestID(),                        // Correlation ID for every request
        middleware.Logging(logger),                    // Structured request/response logging
        middleware.Timeout(30 * time.Second),          // Per-request deadline
        middleware.Recovery(),                         // Panic recovery
    )
}
```

#### 6.2 CLI

Built with `cobra`. Each command maps to an application use case.

```
granola-mcp
├── auth
│   ├── login       → Login use case
│   └── status      → CheckStatus use case
├── sync            → SyncMeetings use case
├── list
│   └── meetings    → ListMeetings use case
├── export
│   └── meeting     → ExportMeeting use case
└── serve           → Starts MCP server (stdio or HTTP+SSE)
```

##### 6.2.1 Command Structure

```go
package cli

import "github.com/spf13/cobra"

func NewRootCmd(deps *Dependencies) *cobra.Command {
    root := &cobra.Command{
        Use:   "granola-mcp",
        Short: "Granola meeting intelligence for the MCP ecosystem",
    }

    root.AddCommand(
        newAuthCmd(deps),
        newSyncCmd(deps),
        newListCmd(deps),
        newExportCmd(deps),
        newServeCmd(deps),
    )

    return root
}
```

##### 6.2.2 Global Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | string | `~/.granola-mcp.yaml` | Config file path |
| `--format` | string | `table` | Output format: `table`, `json`, `md` |
| `--verbose` | bool | `false` | Enable debug logging |
| `--no-color` | bool | `false` | Disable colored output |

### 7. Dependency Injection (Composition Root)

All dependencies are wired in `cmd/granola-mcp/main.go`. No service locator, no global state.

```go
package main

import (
    "database/sql"
    "net/http"
    "os"
    "time"

    "github.com/felixgeelhaar/bolt/v3"
    meetingapp "github.com/felixgeelhaar/granola-mcp/internal/application/meeting"
    authapp "github.com/felixgeelhaar/granola-mcp/internal/application/auth"
    "github.com/felixgeelhaar/granola-mcp/internal/infrastructure/cache"
    "github.com/felixgeelhaar/granola-mcp/internal/infrastructure/granola"
    infraauth "github.com/felixgeelhaar/granola-mcp/internal/infrastructure/auth"
    "github.com/felixgeelhaar/granola-mcp/internal/infrastructure/resilience"
    "github.com/felixgeelhaar/granola-mcp/internal/interfaces/cli"
    mcpiface "github.com/felixgeelhaar/granola-mcp/internal/interfaces/mcp"
)

func main() {
    // Logger
    logger := bolt.New(bolt.NewConsoleHandler(os.Stderr))

    // Config
    cfg := config.Load()

    // Infrastructure
    httpClient := &http.Client{Timeout: cfg.Resilience.Timeout}
    granolaClient := granola.NewClient(cfg.GranolaAPIURL, httpClient, logger)
    granolaRepo := granola.NewRepository(granolaClient)

    // Resilience decorator
    resilientRepo := resilience.NewResilientRepository(granolaRepo, resilience.DefaultConfig())

    // Cache decorator
    db, _ := sql.Open("sqlite3", cfg.CacheDir+"/cache.db")
    cachedRepo := cache.NewCachedRepository(resilientRepo, db, cfg.CacheTTL)

    // Auth
    tokenStore := infraauth.NewFileTokenStore(cfg.ConfigDir)
    authService := infraauth.NewOAuthService(cfg.OAuth.ClientID, cfg.OAuth.RedirectURI, tokenStore)

    // Application use cases
    listMeetings := meetingapp.NewListMeetings(cachedRepo)
    getMeeting := meetingapp.NewGetMeeting(cachedRepo)
    getTranscript := meetingapp.NewGetTranscript(cachedRepo)
    searchTranscripts := meetingapp.NewSearchTranscripts(cachedRepo)
    getActionItems := meetingapp.NewGetActionItems(cachedRepo)
    syncMeetings := meetingapp.NewSyncMeetings(cachedRepo)
    login := authapp.NewLogin(authService)
    checkStatus := authapp.NewCheckStatus(authService)

    // MCP server
    mcpServer := mcpiface.NewServer(listMeetings, getMeeting, getTranscript, searchTranscripts, getActionItems)

    // CLI
    deps := &cli.Dependencies{
        ListMeetings: listMeetings,
        GetMeeting:   getMeeting,
        SyncMeetings: syncMeetings,
        Login:        login,
        CheckStatus:  checkStatus,
        MCPServer:    mcpServer,
        Logger:       logger,
    }

    if err := cli.NewRootCmd(deps).Execute(); err != nil {
        logger.Error().Str("error", err.Error()).Msg("Command failed")
        os.Exit(1)
    }
}
```

### 8. Configuration

#### 8.1 Configuration File

`~/.granola-mcp.yaml`:

```yaml
# Granola API
granola:
  api_url: "https://api.granola.ai"
  auth_method: "oauth"  # "oauth" or "api_token"
  api_token: ""         # Only if auth_method is "api_token"

# MCP Server
mcp:
  server_name: "granola-mcp"
  transport: "stdio"    # "stdio" or "http"
  http_port: 8080       # Only if transport is "http"
  enabled_resources:
    - "meeting"
    - "transcript"
    - "summary"
    - "action_item"
    - "metadata"

# Cache
cache:
  enabled: true
  dir: "~/.granola-mcp/cache"
  ttl: "15m"

# Resilience
resilience:
  circuit_breaker:
    failure_threshold: 5
    success_threshold: 2
    half_open_timeout: "30s"
  rate_limit:
    rate: 100
    interval: "1m"
  retry:
    max_attempts: 3
    backoff: "exponential"
    initial_delay: "500ms"
    max_delay: "10s"
  timeout: "30s"

# Privacy
privacy:
  redact_speakers: false
  redact_keywords: []
  local_only: false

# Sync
sync:
  polling_interval: "5m"
  auto_sync: false

# Logging
logging:
  level: "info"         # "debug", "info", "warn", "error"
  format: "console"     # "console" or "json"
```

#### 8.2 Environment Variable Overrides

All config values can be overridden via environment variables with `GRANOLA_MCP_` prefix:

```
GRANOLA_MCP_GRANOLA_API_URL=https://api.granola.ai
GRANOLA_MCP_GRANOLA_API_TOKEN=gra_xxxxx
GRANOLA_MCP_MCP_TRANSPORT=http
GRANOLA_MCP_MCP_HTTP_PORT=9090
GRANOLA_MCP_CACHE_TTL=30m
GRANOLA_MCP_LOGGING_LEVEL=debug
```

### 9. Distribution

#### 9.1 Homebrew

The project uses GoReleaser to build cross-platform binaries and publish a Homebrew tap.

**`goreleaser.yaml`:**

```yaml
project_name: granola-mcp

builds:
  - main: ./cmd/granola-mcp
    binary: granola-mcp
    env:
      - CGO_ENABLED=1
    goos:
      - darwin
      - linux
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"

brews:
  - name: granola-mcp
    repository:
      owner: felixgeelhaar
      name: homebrew-tap
    homepage: "https://github.com/felixgeelhaar/granola-mcp"
    description: "Granola meeting intelligence for the MCP ecosystem"
    license: "MIT"
    install: |
      bin.install "granola-mcp"
    test: |
      system "#{bin}/granola-mcp", "--version"

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^ci:"
```

**Installation:**

```bash
# Homebrew
brew tap felixgeelhaar/tap
brew install granola-mcp

# Go install
go install github.com/felixgeelhaar/granola-mcp/cmd/granola-mcp@latest

# Binary (GitHub Releases)
# Download from https://github.com/felixgeelhaar/granola-mcp/releases
```

#### 9.2 Build & Release

```makefile
VERSION ?= $(shell git describe --tags --always --dirty)

.PHONY: build test lint release

build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/granola-mcp ./cmd/granola-mcp

test:
	go test -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...

release:
	goreleaser release --clean
```

### 10. Observability

#### 10.1 Structured Logging with Bolt

Every layer logs through `bolt.Logger` with consistent fields:

```go
// Infrastructure layer — API call logging
logger.Info().
    Str("component", "granola_client").
    Str("method", "GET").
    Str("path", "/v2/get-documents").
    Int("status", resp.StatusCode).
    Dur("duration", elapsed).
    Msg("Granola API call")

// Application layer — use case logging
logger.Info().
    Str("component", "list_meetings").
    Int("results", len(meetings)).
    Str("filter_since", input.Since.String()).
    Msg("Meetings listed")

// Interface layer — MCP tool logging
logger.Info().
    Str("component", "mcp_tool").
    Str("tool", "list_meetings").
    Str("request_id", ctx.Value(middleware.RequestIDKey).(string)).
    Dur("duration", elapsed).
    Msg("Tool invocation completed")
```

#### 10.2 Metrics (via Fortify)

Fortify exposes Prometheus metrics for all resilience patterns:

- `granola_mcp_circuit_breaker_state` — Current circuit breaker state
- `granola_mcp_requests_total` — Total API requests (by status)
- `granola_mcp_request_duration_seconds` — API request latency histogram
- `granola_mcp_rate_limit_rejected_total` — Rejected requests due to rate limiting
- `granola_mcp_retry_attempts_total` — Retry attempts by outcome

### 11. Testing Strategy

| Layer | Test Type | Approach |
|-------|-----------|----------|
| **Domain** | Unit tests | Pure logic, no mocks needed. Test aggregate invariants, value object equality, event generation. |
| **Application** | Unit tests | Mock repository ports. Test use case orchestration and edge cases. |
| **Infrastructure** | Integration tests | Test against real Granola API (sandbox) or HTTP recorder (go-vcr). Test SQLite cache. |
| **Interfaces/MCP** | Integration tests | Test MCP protocol compliance using mcp-go test utilities. |
| **Interfaces/CLI** | Integration tests | Test command output using cobra test helpers. |
| **End-to-End** | E2E tests | Full stack: CLI command → use case → cached API → MCP response. |

**Coverage target:** 90%+ on domain and application layers. 80%+ overall.

### 12. Error Handling

#### 12.1 Domain Errors

```go
package meeting

import "errors"

var (
    ErrMeetingNotFound    = errors.New("meeting not found")
    ErrTranscriptNotReady = errors.New("transcript not yet available")
    ErrAccessDenied       = errors.New("access denied to meeting")
    ErrInvalidFilter      = errors.New("invalid filter parameters")
)
```

#### 12.2 Error Mapping

Errors are mapped at layer boundaries:

- **Domain errors** → Application returns them as-is
- **Infrastructure errors** (HTTP 404, 429, 500) → Mapped to domain errors in the repository adapter
- **Application errors** → Mapped to MCP error responses or CLI exit codes in the interfaces layer

```
Granola API 404  →  domain.ErrMeetingNotFound  →  MCP error "Meeting not found"
Granola API 429  →  Fortify rate limit          →  MCP error "Rate limited, retry later"
Granola API 500  →  Fortify circuit breaker     →  MCP error "Service temporarily unavailable"
```

### 13. Security

- **Token storage**: OAuth tokens stored in OS keychain (macOS Keychain, Linux secret-service) with file-based fallback
- **No secrets in config**: API tokens read from environment variables or secure storage, never persisted in plaintext config
- **TLS only**: All Granola API calls over HTTPS, TLS 1.2+ enforced
- **Redaction**: Configurable speaker name and keyword redaction applied at the application layer before data reaches MCP consumers
- **Read-only default**: MCP server exposes only read operations. No write-back until Phase 3.
- **Input validation**: All MCP tool inputs validated via struct tags and mcp-go's automatic schema validation

### 14. Migration Path

| Phase | Scope | Timeline |
|-------|-------|----------|
| **Phase 1** | Core CLI + MCP server (read-only), Homebrew distribution, SQLite cache | MVP |
| **Phase 2** | Transcript search, event streaming, webhook adapter, multi-workspace | +1 release |
| **Phase 3** | Write-back, bidirectional sync, vector embeddings, agent policies | Future |
