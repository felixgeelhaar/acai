# Product Requirements Document (PRD)

## Product Name

Granola MCP Connector (CLI + MCP Server)

## Problem Statement

Granola provides high-quality meeting transcription and summaries, but lacks:

- A native **Model Context Protocol (MCP)** interface
- A **CLI** to export, sync, and stream meeting artifacts
- First-class interoperability with a user’s existing **MCP server suite**

As a result, meeting knowledge is siloed, hard to automate, and not easily consumable by agents, copilots, or downstream tools.

## Goals

- Expose Granola meeting data as structured, queryable MCP resources
- Enable automation-friendly access via a CLI
- Allow seamless integration into multi-MCP environments (local + remote)
- Preserve privacy, security, and user control

## Non-Goals

- Replacing Granola’s UI or core transcription engine
- Building a new note-taking app
- Performing heavy LLM summarization (can be delegated to downstream MCPs)

## Target Users

- Power users with MCP-based agent stacks
- Developers running multiple MCP servers (local + cloud)
- Teams automating meeting follow-ups, tickets, and knowledge bases

## User Stories

1. As a user, I want my meetings to appear as MCP resources so agents can reason over them.
2. As a developer, I want a CLI to sync and inspect meeting data locally.
3. As an agent, I want to subscribe to new meetings or transcript updates.
4. As a security-conscious user, I want fine-grained control over what data is exposed.

## Architecture Overview

Components:

1. **Granola MCP Server** (read-only by default)
2. **Granola CLI** (sync, auth, export, debug)
3. Optional local cache (SQLite / JSONL)

The MCP server may run:

- Embedded inside the CLI
- As a standalone daemon
- As a sidecar alongside other MCP servers

## MCP Server Responsibilities

### Resources

The MCP server MUST expose the following resource types:

- `meeting`
  - id
  - title
  - datetime
  - participants
  - source (Zoom, Meet, etc.)

- `transcript`
  - meeting_id
  - speaker
  - timestamp
  - text
  - confidence (if available)

- `summary`
  - meeting_id
  - type (auto, user-edited)
  - content

- `action_item`
  - meeting_id
  - owner
  - due_date
  - text

- `metadata`
  - tags
  - links
  - external_refs

### Tools

The MCP server SHOULD expose tools such as:

- `list_meetings(filters)`
- `get_meeting(id)`
- `get_transcript(meeting_id, range?)`
- `search_transcripts(query)`
- `get_action_items(meeting_id)`
- `subscribe_new_meetings()`

### Events

Support MCP event streaming:

- `meeting.created`
- `transcript.updated`
- `summary.updated`

## CLI Responsibilities

### Core Commands

- `granola-mcp auth login`
- `granola-mcp auth status`
- `granola-mcp sync`
- `granola-mcp list meetings`
- `granola-mcp export meeting <id>`
- `granola-mcp serve` (starts MCP server)

### Flags

- `--since <date>`
- `--format json|md|txt`
- `--cache-dir`
- `--read-only`

### Output Modes

- Human-readable (tables, summaries)
- Machine-readable (JSON, NDJSON)

## Security & Privacy

- OAuth / API token-based auth
- Per-resource access control
- Configurable redaction:
  - Speaker names
  - Sensitive keywords

- Local-only mode (no outbound calls)

## Configuration

Config file (`~/.granola-mcp.yaml`):

- enabled_resources
- polling_interval
- cache_ttl
- mcp_server_name

## Extensibility

- Webhook adapter (push instead of poll)
- Plugin hooks (pre-export, post-sync)
- Custom resource mappers

## Observability

- Structured logs
- Verbose debug mode
- Sync stats (meetings pulled, failures)

## Success Metrics

- Time-to-first-meeting exposed via MCP
- Number of MCP consumers connected
- Sync reliability (>99%)
- User adoption of CLI workflows

## Open Questions

- Granola API stability and rate limits
- Support for real-time transcription streaming
- Multi-workspace support

## Future Enhancements

- Write-back support (agent-generated notes)
- Bi-directional action item sync
- Vector embedding export
- Per-meeting agent policies
