# granola-mcp

Granola meeting intelligence for the MCP ecosystem. A Go CLI and MCP server that exposes [Granola](https://granola.ai) meetings, transcripts, summaries, and action items as structured MCP resources and tools.

## Installation

### Homebrew

```bash
brew tap felixgeelhaar/tap
brew install granola-mcp
```

### Go install

```bash
go install github.com/felixgeelhaar/granola-mcp/cmd/granola-mcp@latest
```

### From source

```bash
git clone https://github.com/felixgeelhaar/granola-mcp.git
cd granola-mcp
make build
```

## Quick Start

```bash
# Authenticate with Granola
export GRANOLA_MCP_GRANOLA_API_TOKEN=gra_xxxxx
granola-mcp auth login --method api_token

# List recent meetings
granola-mcp list meetings

# Export a meeting as markdown
granola-mcp export meeting <meeting-id> --format md

# Start as MCP server (stdio, for Claude Code)
granola-mcp serve
```

## CLI Commands

```
granola-mcp
  auth
    login       Authenticate with Granola (--method oauth|api_token)
    status      Show current authentication status
  list
    meetings    List meetings (--format table|json, --source, --limit, --since, --until)
  export
    meeting     Export a meeting (--format json|md|text)
  sync          Sync meetings from Granola API (--since)
  serve         Start MCP server on stdio
  version       Show version information
```

## MCP Server

When running as an MCP server (`granola-mcp serve`), the following tools and resources are exposed:

### Tools

| Tool | Description |
|------|-------------|
| `list_meetings` | Search and filter meetings with date, source, and text filters |
| `get_meeting` | Get full meeting details including summary and action items |
| `get_transcript` | Get the transcript with speaker utterances |
| `search_transcripts` | Full-text search across all meeting transcripts |
| `get_action_items` | Get action items from a specific meeting |

### Resources

| URI Pattern | Description |
|-------------|-------------|
| `meeting://{id}` | Full meeting details as JSON |
| `transcript://{meeting_id}` | Transcript utterances as JSON |

### Claude Code Integration

Add to your Claude Code MCP configuration:

```json
{
  "mcpServers": {
    "granola": {
      "command": "granola-mcp",
      "args": ["serve"],
      "env": {
        "GRANOLA_MCP_GRANOLA_API_TOKEN": "gra_xxxxx"
      }
    }
  }
}
```

## Configuration

Configuration uses 12-factor principles: sensible defaults with environment variable overrides.

| Variable | Default | Description |
|----------|---------|-------------|
| `GRANOLA_MCP_GRANOLA_API_URL` | `https://api.granola.ai` | Granola API base URL |
| `GRANOLA_MCP_GRANOLA_API_TOKEN` | — | API token for authentication |
| `GRANOLA_MCP_MCP_TRANSPORT` | `stdio` | MCP transport (`stdio` or `http`) |
| `GRANOLA_MCP_CACHE_TTL` | `15m` | Local cache time-to-live |
| `GRANOLA_MCP_LOGGING_LEVEL` | `info` | Log level (`debug`, `info`, `warn`, `error`) |

## Architecture

The project follows strict Domain-Driven Design with hexagonal architecture:

```
cmd/granola-mcp/main.go          Composition root (DI wiring)

internal/
  domain/                         Pure business logic, zero dependencies
    meeting/                      Meeting aggregate, value objects, events
    auth/                         Token, credential, auth service port
    workspace/                    Workspace aggregate

  application/                    Use cases (one per file)
    meeting/                      ListMeetings, GetMeeting, GetTranscript, ...
    auth/                         Login, CheckStatus
    export/                       ExportMeeting

  infrastructure/                 External adapters
    granola/                      Granola API client + repository (ACL)
    resilience/                   Fortify circuit breaker, retry, rate limit, timeout
    cache/                        SQLite local cache (repository decorator)
    auth/                         File-based token storage
    config/                       12-factor configuration

  interfaces/                     Inbound adapters
    mcp/                          MCP server (mcp-go) with tools + resources
    cli/                          CLI commands (cobra)
```

### Key Libraries

| Library | Purpose |
|---------|---------|
| [felixgeelhaar/mcp-go](https://github.com/felixgeelhaar/mcp-go) | MCP server framework with typed tools, resources, and multi-transport |
| [felixgeelhaar/fortify](https://github.com/felixgeelhaar/fortify) | Resilience patterns: circuit breaker, retry, rate limit, timeout |
| [spf13/cobra](https://github.com/spf13/cobra) | CLI framework |
| [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) | SQLite driver for local cache |

## Development

```bash
# Run tests
make test

# Run tests with race detection
go test -race ./...

# Build
make build

# Release (requires GoReleaser)
make release
```

## License

MIT
