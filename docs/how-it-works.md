# acai — What It Does and How It Works

## What It Is

A single Go binary that bridges [Granola](https://granola.ai) (a meeting intelligence platform) with the [Model Context Protocol](https://modelcontextprotocol.io) ecosystem. It runs in two modes:

1. **CLI** — Human-friendly commands for managing meeting data from the terminal
2. **MCP Server** — Machine-friendly tools and resources that AI agents (Claude Code, custom agents) can call to query, annotate, and export meeting knowledge

## The Problem It Solves

Meeting knowledge is trapped in Granola's UI. An AI agent helping you write a project update can't check what was discussed in yesterday's standup. A developer automating follow-ups can't programmatically access action items. This tool makes meeting data a first-class citizen in agent workflows.

---

## How It Works

### Data Flow (Read Path)

```
Granola Cloud API
       ↓
  Granola Client          Anti-corruption layer: API DTOs → domain types
       ↓
  Resilient Repo          Circuit breaker, retry w/ backoff, rate limit, timeout
       ↓
  Cached Repo             SQLite local cache (configurable TTL, default 15min)
       ↓
  Use Cases               Application layer — one per operation
       ↓
  ┌─────────┐
  │ MCP     │  Tools + Resources for AI agents (stdio or HTTP+SSE)
  │ Server  │
  └─────────┘
  ┌─────────┐
  │ CLI     │  Human-readable commands (table, JSON, markdown output)
  └─────────┘
```

When an agent calls `get_meeting` or a user runs `acai list meetings`, the request flows through the decorator chain. If the cache has a fresh copy, it returns immediately without hitting the Granola API. If the API is down, the circuit breaker trips and returns the cached version. Every external call has a timeout, retry logic, and rate limiting.

### Data Flow (Write Path)

```
Agent calls add_note / complete_action_item
       ↓
  Use Case                Validates input, creates domain entity
       ↓
  Local SQLite Store      Persists to local.db (separate from cache.db)
       ↓
  Outbox Dispatcher       Persists event to outbox table for future upstream sync
       ↓                  AND dispatches to MCP notifier for real-time updates
  Event Dispatcher → MCPNotifier → subscribed MCP sessions
```

The Granola API is read-only, so writes are local-first. Agent notes and action item overrides live in a local SQLite database. An outbox table captures every write event so a future sync mechanism can push changes upstream when the API supports it.

### Policy Enforcement

```
Agent calls get_transcript with tags: ["confidential"]
       ↓
  PolicyMiddleware
    1. Extract meeting context (tags) from request
    2. Evaluate ACL rules (first-match-wins)
       → If denied: return error immediately
    3. Delegate to inner server
    4. Redact response (emails, speakers, keywords, patterns)
       ↓
  Redacted JSON response back to agent
```

The policy middleware sits between the MCP transport and the server handlers. It's configured via a YAML file and supports two mechanisms:
- **ACL** — Block specific tools for meetings matching tag conditions (e.g., deny `get_transcript` for `confidential` meetings)
- **Redaction** — Scrub sensitive data from all responses: emails → `[EMAIL]`, speaker names → `Speaker 1`, keywords → `[REDACTED]`, custom regex patterns

---

## MCP Server Capabilities

When an AI agent connects (e.g., Claude Code via stdio), it sees:

### Tools (13)

| Tool | What It Does |
|------|-------------|
| `list_meetings` | Search meetings by date range, source (Zoom/Teams/etc), participant, text query |
| `get_meeting` | Full meeting details: title, participants, summary, action items |
| `get_transcript` | Speaker-attributed transcript with timestamps and confidence scores |
| `search_transcripts` | Full-text search across all meeting transcripts |
| `get_action_items` | Action items with owner, text, due date, completion status |
| `meeting_stats` | Aggregated statistics: frequency, platform distribution, speaker talk time, heatmap |
| `list_workspaces` | List all Granola workspaces |
| `add_note` | Attach an agent-generated note to a meeting |
| `list_notes` | List agent notes for a meeting |
| `delete_note` | Remove an agent note |
| `complete_action_item` | Mark an action item as done (local override) |
| `update_action_item` | Change action item text (local override) |
| `export_embeddings` | Chunk meeting content into JSONL for embedding pipelines |

### Resources (5)

| URI | What It Returns |
|-----|----------------|
| `meeting://{id}` | Meeting JSON (metadata, participants, summary, action items) |
| `transcript://{meeting_id}` | Transcript JSON (speaker utterances with timestamps) |
| `note://{meeting_id}` | Agent notes JSON |
| `workspace://{id}` | Workspace JSON (name, slug) |
| `ui://meeting-stats` | Self-contained D3.js HTML dashboard (MCP App) |

### Real-time Events

The server dispatches domain events (`MeetingCreated`, `TranscriptUpdated`, `NoteAdded`, `ActionItemCompleted`, etc.) to connected MCP sessions. Agents can react to meeting updates as they happen.

---

## CLI Capabilities

```bash
# Authentication
acai auth login --method api_token    # Set API token
acai auth status                      # Check auth state

# Read operations
acai list meetings --since 2025-01-01 --source zoom --format json
acai export meeting <id> --format md

# Write operations
acai note add <meeting-id> "Agent observation about Q4 targets"
acai note list <meeting-id>
acai note delete <note-id>
acai action complete <meeting-id> <action-item-id>
acai action update <meeting-id> <action-item-id> "Revised text"

# Embedding export
acai export embeddings --meetings m-1,m-2 --strategy speaker_turn --max-tokens 512

# Sync & serve
acai sync --since 2025-01-01
acai serve                            # Start MCP server on stdio
```

---

## Architecture (DDD + Hexagonal)

The codebase is organized in four strict layers with dependency arrows pointing inward:

### Domain (`internal/domain/`)

Pure Go, zero external dependencies. Contains:

- `meeting/` — Meeting aggregate root, value objects (Participant, Summary, ActionItem, Transcript, Utterance, Chunk), domain events, repository port
- `annotation/` — Separate bounded context for agent notes (AgentNote entity, NoteRepository port, events)
- `policy/` — Policy value objects (Rule, Conditions, Effect, RedactionConfig) with first-match-wins evaluation
- `auth/`, `workspace/` — Auth tokens and workspace aggregates

### Application (`internal/application/`)

One use case per file, each with `Execute(ctx, input) (output, error)`:

- `meeting/` — ListMeetings, GetMeeting, GetTranscript, SearchTranscripts, GetActionItems, GetMeetingStats, SyncMeetings, CompleteActionItem, UpdateActionItem
- `annotation/` — AddNote, ListNotes, DeleteNote
- `embedding/` — ExportEmbeddings with pluggable chunking strategies (BySpeakerTurn, ByTimeWindow, ByTokenLimit) and format abstraction (JSONL)

### Infrastructure (`internal/infrastructure/`)

External adapters:

- `granola/` — HTTP client + repository mapping API DTOs to domain types (anti-corruption layer)
- `resilience/` — Fortify decorator: circuit breaker, retry, rate limit, timeout
- `cache/` — SQLite cached repository decorator
- `localstore/` — SQLite store for notes and action item overrides
- `outbox/` — Event dispatcher decorator that persists write events
- `policy/` — YAML loader, redaction engine (email regex, speaker anonymization, keyword replacement, compiled patterns)
- `events/` — Domain event dispatcher with MCP notifier bridge
- `sync/` — Background polling goroutine
- `webhook/` — HMAC-SHA256 validated HTTP handler

### Interfaces (`internal/interfaces/`)

Inbound adapters:

- `mcp/` — MCP server (tools, resources, handlers, policy middleware) built on [mcp-go](https://github.com/felixgeelhaar/mcp-go)
- `cli/` — Cobra commands mapping 1:1 to use cases

### Composition Root (`cmd/acai/main.go`)

The only place that knows about all layers. Wires everything with constructor injection, no service locator.

---

## Resilience Model

Every call to the Granola API passes through:

1. **Rate Limiter** — Token bucket (100 req/min default), prevents API abuse
2. **Circuit Breaker** — Opens after 5 failures, half-open after 30s, closes after 2 successes
3. **Timeout** — 30s per request
4. **Retry** — Up to 3 attempts with exponential backoff (500ms → 10s cap)

If the API is completely down, the cache serves stale data. The server never crashes — it degrades gracefully.

---

## Test Suite

377 tests, 0 race conditions, 82.7% overall coverage. Domain layer at 94-100%, application layer at 86-100%, infrastructure at 85-97%. Every layer is tested in isolation with mock implementations of ports/interfaces.
