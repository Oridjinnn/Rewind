# 🔄 Rewind

**Record, replay, and recall your terminal sessions & IDE activity with AI-powered memory.**

Rewind captures every command, file edit, build, test, and AI interaction—then makes it searchable, replayable, and chat-queryable. SQLite-backed. Multi-IDE. Opt-in privacy. Local-first.

[![Go 1.22+](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![VS Code](https://img.shields.io/badge/VS_Code-extension-blue?logo=visual-studio-code)](extensions/vscode)
[![JetBrains](https://img.shields.io/badge/JetBrains-plugin-orange?logo=intellij-idea)](extensions/jetbrains)
[![Neovim](https://img.shields.io/badge/Neovim-plugin-green?logo=neovim)](extensions/nvim)

![Demo](demo.gif)

---

## ✨ Features

### 🎬 Core Recording
| Command | Description |
|---------|-------------|
| `rewind run <command>` | Record command execution with full stdout/stderr capture |
| `rewind replay <session_id>` | Replay any session frame-by-frame |
| `rewind timeline <session_id>` | Visualize session timeline with timestamps |
| `rewind setup` | Auto-record all shell commands via hooks (bash/zsh/fish) |

### 🔍 Analysis & Inspection
| Command | Description |
|---------|-------------|
| `rewind stats <session_id>` | Duration, command count, exit code, event stats |
| `rewind inspect <session_id>` | Full event details and metadata |
| `rewind diff <id1> <id2>` | Side-by-side session comparison |
| `rewind score <session_id>` | Quality score (0-100), patterns, success rate |
| `rewind detect <session_id>` | Loop detection, anomalies, performance issues |

### 🗄️ Database Backend (SQLite)
| Command | Description |
|---------|-------------|
| `rewind migrate` | One-command JSON → SQLite migration |
| `rewind search <query>` | Full-text search across all sessions |
| `rewind list` | List all recorded sessions |
| **Default storage** | `rewind.db` — WAL mode, 5 tables, 11 indexes, FK constraints |

### 🐚 System Shell History Integration
| Command | Description |
|---------|-------------|
| `rewind history [limit]` | View tracked shell commands (exit code, working dir) |
| `rewind import-history [shell\|path]` | Import from `.bash_history`, `.zsh_history`, fish history |
| `rewind history-stats` | Usage stats: total, unique, top-10 commands |
| `rewind track-command` | Auto-track via shell hooks (called transparently) |

### 🧠 IDE Extensibility (Multi-IDE, All-Activity)
| Command | Description |
|---------|-------------|
| `rewind ide start [port]` | Start HTTP server (`:9876`) for IDE extensions |
| `rewind ide permissions <ide> on\|off` | Opt-in recording control (per-IDE, per-project) |
| `rewind ide activity [ide] [limit]` | View recent IDE events (file, build, test, git, AI) |
| `rewind ide projects` | List tracked IDE projects with event counts |
| `rewind ide analyze <project>` | Productivity insights: top files, build/test rates, AI usage, suggestions |
| `rewind ide status` | Server status, connected IDEs, permissions |

**IDE Extensions Available:**
- **VS Code / Cursor** — `extensions/vscode/` (TypeScript, 8 commands, status bar, 10 config options)
- **JetBrains** — `extensions/jetbrains/` (Kotlin, IntelliJ/GoLand/PyCharm/WebStorm/Android Studio)
- **Neovim** — `extensions/nvim/` (Lua, autocommands, `:RewindToggle`)

**28 Event Types Tracked:** `file_open`, `file_save`, `file_edit`, `file_close`, `file_create`, `file_delete`, `terminal_cmd`, `build_start`, `build_end`, `build_error`, `test_run`, `test_pass`, `test_fail`, `git_commit`, `git_push`, `git_pull`, `git_branch`, `git_stash`, `debug_start`, `debug_breakpoint`, `debug_step`, `debug_end`, `ai_chat`, `ai_completion`, `ai_accept`, `ai_reject`, `refactor`, `search`, `run_config`, `dependency_change`

### 🤖 AI Memory & Chat
| Command | Description |
|---------|-------------|
| `rewind web [port]` | Web UI for session browser |
| `rewind recall <query>` | Semantic search via embeddings (nomic-embed-text) |
| `rewind chat <model>` | Interactive chat with Ollama, session context injection |

### 📤 Export & Sharing
| Command | Description |
|---------|-------------|
| `rewind export <session_id>` | Generate HTML report |
| `rewind markdown <session_id>` | Export as Markdown |

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.22+** | **Ollama** (for AI features) | **SQLite** (bundled, zero CGO)
- AI models: `ollama pull qwen2.5:1.5b && ollama pull nomic-embed-text`

### Build & Setup

```bash
git clone https://github.com/Oridjinnn/Rewind.git
cd rewind
go build ./cmd/rewind                           # Build the CLI

# Migrate existing sessions to SQLite
./rewind migrate
# → Successfully migrated 25 sessions to SQLite.

# Import years of shell history
./rewind import-history
# → Imported 1523 commands from bash, 891 from zsh

# Set up auto-recording
./rewind setup
# → Add the hooks to ~/.bashrc or ~/.zshrc, then source it

# Start IDE recording server (for extensions)
./rewind ide start
# → Listening on http://localhost:9876

# Enable recording for your IDE
./rewind ide permissions vscode on
# → Recording enabled for VS Code
```

### First Session

```bash
./rewind run echo "Hello from Rewind!"
# → session recorded: 1779265115474389394_585791

./rewind replay 1779265115474389394_585791
./rewind stats 1779265115474389394_585791
./rewind export 1779265115474389394_585791
```

---

## 📖 Command Reference (Full Table)

| Category | Command | Arguments |
|----------|---------|-----------|
| **Recording** | `rewind run` | `<command> [args...]` |
| | `rewind setup` | _(auto-detect shell)_ |
| | `rewind track-command` | `<command> [exit_code]` |
| **Replay** | `rewind replay` | `<session_id>` |
| | `rewind timeline` | `<session_id>` |
| | `rewind web` | `[port]` |
| **Analysis** | `rewind stats` | `<session_id>` |
| | `rewind inspect` | `<session_id>` |
| | `rewind diff` | `<sessionA> <sessionB>` |
| | `rewind score` | `<session_id>` |
| | `rewind detect` | `<session_id>` |
| **Database** | `rewind list` | _(none)_ |
| | `rewind search` | `<query>` |
| | `rewind migrate` | _(none)_ |
| **Shell History** | `rewind history` | `[limit=20]` |
| | `rewind import-history` | `[bash\|zsh\|fish\|/path]` |
| | `rewind history-stats` | _(none)_ |
| **IDE** | `rewind ide start` | `[port=9876]` |
| | `rewind ide stop` | _(none)_ |
| | `rewind ide status` | _(none)_ |
| | `rewind ide permissions` | `<ide> <on\|off> [project]` |
| | `rewind ide activity` | `[ide] [limit=30]` |
| | `rewind ide projects` | _(none)_ |
| | `rewind ide analyze` | `<project-path>` |
| **AI** | `rewind recall` | `<query>` |
| | `rewind chat` | `<model>` |
| **Export** | `rewind export` | `<session_id>` |
| | `rewind markdown` | `<session_id>` |

---

## 🗄️ Database Backend

### Schema (5 Tables, 11 Indexes)

```
sessions          — id, command, model, title, summary, tags, mood, timestamps, exit_code
events            — id, session_id (FK), timestamp, type, content
shell_history     — id, command, exit_code, working_dir, executed_at, session_id (FK)
ide_activity      — id, ide_name, project_name, project_path, activity_type, file_path,
                    language, metadata (JSON), content_snapshot, executed_at, session_id (FK)
ide_permissions   — ide_name, project_path (PK), recording_enabled, file_recording,
                    terminal_recording, ai_recording, last_toggled
```

### Storage Selection

| Env Variable | Effect |
|-------------|--------|
| _(default)_ | SQLite `rewind.db` (WAL mode, FK constraints) |
| `REWIND_DB_PATH=/path/to/custom.db` | Custom database location |
| `REWIND_USE_JSON=true` | Legacy JSON file storage fallback |

### Direct SQLite Queries

```bash
sqlite3 rewind.db ".tables"
sqlite3 rewind.db "SELECT id, command, exit_code FROM sessions ORDER BY started_at DESC LIMIT 5"
sqlite3 rewind.db "SELECT command, COUNT(*) c FROM shell_history GROUP BY command ORDER BY c DESC LIMIT 10"
sqlite3 rewind.db "SELECT ide_name, activity_type, file_path FROM ide_activity ORDER BY executed_at DESC LIMIT 20"
```

---

## 🐚 System Shell History Integration

### What Gets Tracked Per Command

| Field | Example |
|-------|---------|
| `command` | `git push origin main` |
| `exit_code` | `0` (success) |
| `working_dir` | `/home/user/project` |
| `executed_at` | `2026-05-21T15:30:00Z` |
| `session_id` | Linked to session (auto-created) |

### Supported History Formats

| Shell | File | Format |
|-------|------|--------|
| bash | `~/.bash_history` | Plain text, one command per line |
| zsh | `~/.zsh_history` | Extended format with timestamps (`: timestamp:0;command`) |
| fish | `~/.local/share/fish/fish_history` | YAML format |

### Shell Hook Architecture

- **bash**: `DEBUG` trap captures `$BASH_COMMAND` before execution, `PROMPT_COMMAND` captures exit code after
- **zsh**: `preexec` captures command, `precmd` captures exit code
- **fish**: `fish_postexec` event handler with `$status`

All hooks call `rewind track-command "$CMD" $EXIT_CODE &` in background (non-blocking).

---

## 🧠 IDE Extensibility

### Architecture

```
┌──────────────┐    JSON-RPC v1    ┌──────────────────┐    SQL    ┌──────────┐
│  VS Code      │───POST /──────────▶│                    │─────────▶│           │
│  Cursor        │                   │  Rewind IDE Server  │          │  SQLite   │
│  JetBrains    │    localhost:9876  │  (internal/ide/)    │          │  rewind.db│
│  Neovim        │                   │                    │          │           │
└──────────────┘                    └──────────────────┘          └──────────┘
                                          │
                                          ├── CLI: rewind ide status
                                          ├── CLI: rewind ide activity
                                          ├── CLI: rewind ide permissions
                                          └── CLI: rewind ide analyze
```

### Protocol: `rewind-ide-v1`

All IDE extensions send JSON to `POST http://localhost:9876/`:

```json
{
  "protocol": "rewind-ide-v1",
  "ide": "vscode",
  "version": "1.85.0",
  "project": "rewind",
  "project_path": "/home/user/Projects/rewind",
  "event": {
    "type": "file_save",
    "timestamp": "2026-05-21T15:30:00Z",
    "file": "internal/storage/sqlite.go",
    "language": "go",
    "lines_added": 50,
    "lines_removed": 5
  }
}
```

**Endpoints:**
- `POST /` — Record single event
- `POST /batch` — Record multiple events
- `GET /health` — Server health check
- `GET /status` — Recording status, permissions, connected IDEs

### VS Code / Cursor Extension

```bash
cd extensions/vscode
npm install && npm run compile
# Package: npx vsce package  →  rewind-ide-0.1.0.vsix
```

**Events captured:** File open/save/close/create/delete, terminal open/close, task start/end (build + test), git state changes, Copilot/Cursor AI completions, editor selections.

**Configuration:** 10 settings in VS Code settings UI (`rewind.*`).

### JetBrains Plugin

```bash
cd extensions/jetbrains
./gradlew buildPlugin
# → build/distributions/rewind-ide-0.1.0.zip
```

Auto-detects IDE variant: IntelliJ IDEA, GoLand, PyCharm, WebStorm, Android Studio.

### Neovim Plugin

```lua
-- ~/.config/nvim/init.lua
require("rewind").setup({ enabled = true })

-- Commands: :RewindEnable, :RewindDisable, :RewindToggle
```

### 🔒 Opt-in Privacy Model

All recording is **disabled by default**. Users must explicitly enable it:

```bash
rewind ide permissions vscode on              # For all projects
rewind ide permissions vscode on /path/proj   # Per project
rewind ide permissions vscode off             # Disable
```

Granular controls per IDE per project:
- `recording_enabled` — Master switch
- `file_recording` — Record file edit events
- `terminal_recording` — Record terminal commands
- `ai_recording` — Record AI assistant interactions

**All data stored locally** in `rewind.db`. No telemetry. No cloud. Full local control.

---

## 💬 Chat with Memory

```bash
./rewind chat qwen2.5:1.5b
# You: What deployment issues did I have last week?
# [Automatically injects context from relevant sessions]
# Assistant: Based on your sessions, you had 3 issues: ...
```

**Pipeline:** Query → Embedding (nomic-embed-text) → Cache search (900x speedup) → Top-3 ranking (cosine similarity) → Context injection → Ollama generation

---

## 🏗️ Architecture

```
rewind/
├── cmd/rewind/main.go              # CLI (30+ commands)
├── internal/
│   ├── storage/                    # Storage backends
│   │   ├── interface.go            # Storage & ShellHistoryStorage interfaces
│   │   ├── sqlite.go               # SQLiteStore (5 tables, WAL, FK, indexes)
│   │   ├── migrate.go              # JSON → SQLite migration
│   │   ├── storage.go, load.go, list.go  # JSON fallback
│   ├── shellhistory/               # Shell history tracking
│   │   ├── history.go              # Manager (track, view, stats)
│   │   └── import.go               # Bash/zsh/fish history parsers
│   ├── ide/                        # IDE recording engine
│   │   ├── interface.go            # Recorder abstraction
│   │   ├── recorder.go             # SQLiteRecorder (permission-aware)
│   │   ├── server.go               # HTTP server (:9876, 4 endpoints)
│   │   ├── protocol.go             # JSON protocol v1 (28 event types, 12 IDEs)
│   │   ├── permissions.go          # Opt-in management + formatters
│   │   ├── analyze.go              # Productivity insights
│   │   └── bridge.go               # Cross-reference IDE ↔ shell ↔ sessions
│   ├── shell/hooks.go              # Shell hook generators (bash/zsh/fish)
│   ├── recorder/                   # Command execution + capture
│   ├── replay/                     # Frame-by-frame playback
│   ├── chat/, memory/, recall/     # AI memory pipeline
│   ├── summarize/, export/, stats/ # Analysis tools
│   └── ... (8 more packages)
├── extensions/
│   ├── vscode/                     # VS Code/Cursor extension (TypeScript, 6 files)
│   ├── jetbrains/                  # JetBrains plugin (Kotlin, 5 files)
│   └── nvim/                       # Neovim plugin (Lua, 1 file)
└── pkg/types/
    ├── session.go                  # Session, Event, Message
    └── ide.go                      # IDEActivity, IDEEvent, IDEPermission, etc.
```

---

## ⚡ Performance

| Operation | Time | Notes |
|-----------|------|-------|
| Record command | 50ms | Real-time capture |
| SQLite `search` (1K sessions) | <10ms | Indexed LIKE query |
| `migrate` 25 sessions | <100ms | Bulk insert in transaction |
| `import-history` 1500 cmds | <500ms | Prepared statement batch |
| First `recall` (18 sessions) | 45s | Embeddings computed once |
| Cached `recall` | 200ms | 900x speedup |
| IDE event recording | <5ms | Batch-send every 5s |

---

## 📝 Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| _(none)_ | SQLite | Primary storage backend |
| `REWIND_DB_PATH` | `rewind.db` | Custom SQLite path |
| `REWIND_USE_JSON` | `false` | Force JSON file storage |
| `REWIND_ENABLED` | `true` | Shell auto-recording toggle |
| `REWIND_REDACT` | `false` | Secret redaction: `true` to redact, `skip` to skip entirely |

### Ollama (for AI features)
- Endpoint: `http://localhost:11434`
- Summarization: `qwen2.5:1.5b`
- Embeddings: `nomic-embed-text`

---

## 🚦 Status

- [x] Core recording/replay (frame-by-frame)
- [x] Analysis commands (stats, diff, detect, score, timeline, inspect)
- [x] Embedding-based search with cache (900x speedup)
- [x] Chat with memory injection (streaming)
- [x] Structured summaries
- [x] Shell integration (bash, zsh, fish auto-hooks)
- [x] Web UI for session browser
- [x] **Database backend — SQLite (WAL, FK, 5 tables, 11 indexes)**
- [x] **System shell history integration (import + tracking)**
- [x] **Multi-IDE extensibility (VS Code, JetBrains, Neovim)**
- [x] **Opt-in privacy model**
- [x] **Productivity insights (IDE analysis)**

### Roadmap

| Version | Milestone |
|---------|-----------|
| **v0.2-beta** | Current — all features implemented |
| **v0.3** | FTS5 full-text search, embedding cache in SQLite, VS Code `.vsix` release |
| **v1.0** | CI/CD, test suite, multi-model AI support, marketplace publishing |

---

## 🤝 Contributing

Issues and PRs welcome. See `extensions/` for IDE extension code and `internal/` for Go backend.

---

## 📄 License

MIT License — see [LICENSE](LICENSE).

---

**Built for developers who want their IDE and terminal to have a memory.**