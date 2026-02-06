# Vision Document

## Granola MCP Connector

### Executive Summary

Granola MCP Connector bridges the gap between Granola's meeting intelligence platform and the Model Context Protocol ecosystem. It is a Go CLI and MCP server that exposes meeting transcripts, summaries, action items, and metadata as structured, queryable MCP resources — enabling AI agents, copilots, and automation pipelines to reason over meeting knowledge without leaving the MCP protocol.

### The Problem

Meeting knowledge is trapped. Granola captures high-quality transcripts, summaries, and action items, but this data lives behind a proprietary UI with no native agent-friendly interface. Users running MCP-based stacks (Claude Code, custom agents, multi-server compositions) cannot access their meeting intelligence programmatically. The result:

- **Knowledge silos** — Meeting insights require manual copy-paste into downstream tools
- **Broken automation** — No path from "meeting happened" to "ticket created, notes filed, follow-ups scheduled" without human glue
- **Agent blindness** — AI agents can't reference past meetings, search transcripts, or act on action items

### The Vision

A single binary — `granola-mcp` — that makes meeting intelligence a first-class citizen in the MCP ecosystem.

**For power users**: A CLI to authenticate, sync, search, and export meeting data in human-readable or machine-readable formats.

**For agents**: An MCP server exposing meetings, transcripts, summaries, and action items as typed resources and tools — queryable, subscribable, and composable with other MCP servers.

**For teams**: A reliable, observable, production-grade connector that handles Granola API failures gracefully, respects rate limits, and provides fine-grained access control over sensitive meeting data.

### Core Principles

1. **Read-first, write-later** — Start with read-only access to Granola data. Write-back (agent-generated notes, bidirectional action item sync) is a future phase.

2. **Privacy by default** — Configurable redaction of speaker names and sensitive keywords. Local-only mode with no outbound calls. Per-resource access control.

3. **Resilience as a feature** — The Granola API is an external dependency. Every call goes through circuit breakers, retries with exponential backoff, rate limiting, and timeouts. The server degrades gracefully, never crashes.

4. **Zero-allocation observability** — Structured logging with trace context propagation. Every request is traceable from MCP tool invocation through Granola API call to response.

5. **Single binary, multiple modes** — One Go binary serves as CLI, embedded MCP server (stdio), standalone daemon (HTTP+SSE), or sidecar. No runtime dependencies beyond the binary itself.

### Target Audience

| Persona | Need | How We Serve It |
|---------|------|-----------------|
| **MCP Power User** | Meeting data in their agent stack | MCP resources + tools, stdio transport |
| **Developer** | Automate meeting follow-ups | CLI commands, JSON output, scripting |
| **Team Lead** | Track action items across meetings | `get_action_items` tool, filtering by owner/date |
| **Security-Conscious User** | Control over exposed data | Redaction config, per-resource ACL, local-only mode |
| **Platform Engineer** | Run as a service alongside other MCPs | HTTP+SSE transport, health checks, metrics |

### What Success Looks Like

**Phase 1 — Foundation (MVP)**
- Authenticate with Granola via OAuth/API token
- Sync meetings to local cache
- Expose meetings, transcripts, summaries, action items as MCP resources
- CLI: `auth`, `sync`, `list`, `export`, `serve`
- Installable via `go install` and Homebrew

**Phase 2 — Intelligence**
- Full-text transcript search across all meetings
- Event streaming: `meeting.created`, `transcript.updated`, `summary.updated`
- Webhook adapter for push-based sync
- Multi-workspace support

**Phase 3 — Collaboration**
- Write-back: agent-generated notes pushed to Granola
- Bidirectional action item sync
- Vector embedding export for semantic search
- Per-meeting agent policies

### Technology Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| **Language** | Go 1.23+ | Single binary, cross-platform, performance |
| **MCP Framework** | `felixgeelhaar/mcp-go` | Production-ready MCP server framework with typed tools, middleware, and multi-transport |
| **Resilience** | `felixgeelhaar/fortify` | Zero-allocation circuit breakers, retries, rate limiting, timeouts, bulkheads |
| **Logging** | `felixgeelhaar/bolt` | Zero-allocation structured logging with OpenTelemetry integration |
| **Architecture** | Domain-Driven Design (DDD) | Clean separation of domain logic, application services, and infrastructure |
| **Distribution** | Homebrew, `go install`, GitHub Releases | Standard Go distribution channels |

### Differentiators

- **Not a wrapper** — This is a domain-aware connector that understands meeting semantics, not a generic API-to-MCP bridge
- **Production-grade from day one** — Built on battle-tested resilience and observability libraries
- **Composable** — Designed to work alongside other MCP servers in a multi-server stack, not as a monolithic solution
- **Open source** — Community-driven, extensible via plugin hooks and custom resource mappers

### Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Granola API instability or breaking changes | Medium | High | Circuit breakers + retry; abstract API behind port interface; version-aware client |
| Rate limit exhaustion | Medium | Medium | Token bucket rate limiter; local cache reduces API calls; configurable polling interval |
| Granola API access restricted to Enterprise tier | High | High | Support both OAuth and API token auth; graceful degradation for missing endpoints |
| MCP protocol evolution | Low | Medium | `mcp-go` framework abstracts protocol details; follow MCP spec releases |

### Non-Goals

- Replacing Granola's UI or core transcription engine
- Building a standalone note-taking application
- Performing LLM summarization (delegated to downstream MCP consumers)
- Real-time transcription streaming (dependent on Granola API capabilities)
