# LazyArchon

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Homebrew](https://img.shields.io/badge/Homebrew-Available-orange?style=flat&logo=homebrew)](https://github.com/yousfiSaad/homebrew-lazyarchon)

A terminal task manager with two faces: a [lazygit](https://github.com/jesseduffield/lazygit)-style TUI for you, and an [MCP](https://modelcontextprotocol.io) server for Claude — both working on the same local SQLite database.

```bash
lazyarchon        # you: the TUI
lazyarchon mcp    # Claude: an MCP server on the same tasks
```

No server, no account: the first run creates a local database with a seeded **Inbox** project. Prefer an existing tracker? Point it at [Gitea](https://gitea.io) issues, [Vikunja](https://vikunja.io) or an [Archon](https://github.com/coleam00/Archon) backend — the TUI and the MCP server follow.

## Demo

![LazyArchon Demo](assets/demo/lazyarchon-demo.gif)

## Install

**Homebrew (macOS/Linux):**

```bash
brew install yousfisaad/lazyarchon/lazyarchon
```

**Go (1.25+):**

```bash
go install github.com/yousfisaad/lazyarchon/v2/cmd/lazyarchon@latest
```

**Binary:** grab one from the [latest release](https://github.com/yousfiSaad/lazyarchon/releases/latest) — darwin / linux (amd64, arm64) and windows (amd64).

**From source:**

```bash
git clone https://github.com/yousfiSaad/lazyarchon
cd lazyarchon && make build && ./bin/lazyarchon
```

## Use it with Claude

```bash
claude mcp add lazyarchon -- lazyarchon mcp
```

Then just ask: *"list my projects"*, *"add a high-priority task to review the PR, due Friday"*, *"mark it done"*. Changes appear in the TUI after a refresh (`r`); TUI edits are visible to Claude.

## Keyboard shortcuts

| Key | Action |
| --- | --- |
| `?` | Show help (all shortcuts) |
| `h/l` | Switch panels (Tasks ↔ Details) |
| `j/k` | Navigate up / down |
| `J/K` | Fast scroll |
| `c` | Create a new task |
| `e` | Edit the selected task |
| `t` | Quick status change |
| `d` | Delete task (with confirmation) |
| `s` / `S` | Cycle sort mode forward / backward |
| `f` | Filter by features (multi-select) |
| `/` | Search tasks |
| `n` | Next search result |
| `a` | Show all tasks |
| `y` / `Y` | Copy task ID / title |
| `p` | Select project |
| `r` | Refresh data |
| `q` | Quit |

## Configuration

`~/.config/lazyarchon/config.yaml` (or `./config.yaml`); see [`config.example.yaml`](config.example.yaml):

```yaml
plugin: local   # local (default) | gitea | vikunja | archon
# plugins.local.path overrides the database location
# (or set LAZYARCHON_DB_PATH)
```

Statuses: `todo`, `doing`, `review`, `done`. Priority: 1 (critical) … 4 (low).

## Development

```bash
make build   # build to bin/lazyarchon
make test    # go test ./...
make lint    # golangci-lint (applies fixes)
```

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no cgo) and the [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk).

## License

[MIT](LICENSE)
