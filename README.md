# 🔄 Rewind

**Record, replay, and recall your terminal sessions with AI-powered memory.**

Rewind captures every command execution, output, and interaction—then lets you search, analyze, and chat about your history using embeddings and local LLMs.

[![Go 1.21+](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## ✨ Features

### Core Recording
- **`rewind run <command>`** — Record command execution with full stdout/stderr capture
- **`rewind replay <session_id>`** — Replay any recorded session frame-by-frame
- **`rewind timeline <session_id>`** — Visualize session timeline with timestamps

### Analysis & Inspection
- **`rewind stats <session_id>`** — Session statistics (duration, command count, exit code)
- **`rewind inspect <session_id>`** — Deep dive into events and metadata
- **`rewind diff <id1> <id2>`** — Compare two sessions side-by-side
- **`rewind score <session_id>`** — Score session quality (loops, patterns, success)
- **`rewind detect <session_id>`** — Detect infinite loops and performance issues

### AI Memory & Chat
- **`rewind web [port]`** — Start the Rewind web UI in your browser
- **`rewind recall <query>`** — Search across all sessions using semantic embeddings
- **`rewind chat <model>`** — Interactive chat with Ollama, using session context
  - Real-time streaming responses (see AI output as it's generated)
  - Automatic context injection from relevant past sessions
  - Embedding cache for 900x speedup on repeated queries

### 🗄️ Database Backend (SQLite)
- **SQLite storage** — Sessions and events stored in a relational database
  - WAL mode for concurrent read/write performance
  - Full-text search across commands, titles, and summaries
  - Foreign key constraints ensuring data integrity
  - Zero CGO dependency (`modernc.org/sqlite` — pure Go)
- **`rewind migrate`** — One-command migration from JSON files to SQLite
- **`rewind search <query>`** — Full-text search across all sessions
- **Backward compatible** — JSON fallback with `REWIND_USE_JSON=true`

### 🐚 System Shell History Integration
- **`rewind setup`** — Auto-record all your commands via shell hooks (bash, zsh, fish)
- **`rewind import-history [shell|path]`** — Import existing shell history files
  - Parses `.bash_history` (plain text)
  - Parses `.zsh_history` (extended format with timestamps)
  - Parses fish history (YAML format)
  - Auto-detects available shell histories
- **`rewind history [limit]`** — View tracked shell command history
- **`rewind history-stats`** — Usage statistics (total, unique, top commands)
- **Real-time tracking** — Every command automatically logged to SQLite with exit code and working directory

### Export & Sharing
- **`rewind export <session_id>`** — Generate HTML report
- **`rewind markdown <session_id>`** — Export as Markdown with formatting
- **`rewind list`** — Show all recorded sessions

---

## 🚀 Quick Start

### Prerequisites

1. **Go 1.22+** — [Install](https://golang.org/doc/install)
2. **Ollama** — [Install](https://ollama.ai) (for AI features)
3. Models to pull:
   ```bash
   ollama pull qwen2.5:1.5b        # For chat & summarization
   ollama pull nomic-embed-text     # For semantic search
   ```

### Build

```bash
git clone https://github.com/habeldavidson007-glitch/rewind.git
cd rewind
go build ./cmd/rewind
```

### Migrate Existing Sessions to SQLite

```bash
# Convert all JSON sessions to SQLite
./rewind migrate
# Output: Successfully migrated 25 sessions to SQLite.
```

### Import Your Shell History

```bash
# Auto-detect and import all available shell histories
./rewind import-history
# Output:
# Imported 1523 commands from bash history (.bash_history)
# Imported 891 commands from zsh history (.zsh_history)

# Or import a specific shell:
./rewind import-history bash
./rewind import-history zsh
./rewind import-history /path/to/custom/history/file
```

### First Session

```bash
# Record a session
./rewind run echo "Hello from Rewind!"
# Output: session recorded: 1779265115474389394_585791

# Replay it
./rewind replay 1779265115474389394_585791

# View stats
./rewind stats 1779265115474389394_585791

# Export as HTML
./rewind export 1779265115474389394_585791
```

### View Shell History & Stats

```bash
# View last 20 commands
./rewind history

# View last 50 commands
./rewind history 50

# Show usage statistics
./rewind history-stats
```

### Auto-Recording Setup (Shell Integration)

Make Rewind invisible—automatically record every command without forgetting:

```bash
# Step 1: Show setup instructions
./rewind setup

# Step 2: Add the code to your shell config (~/.bashrc, ~/.zshrc, etc)
# Step 3: Reload your shell
source ~/.bashrc  # or ~/.zshrc for zsh

# Done! Every command now auto-records
cd ~/ && ls        # ← this is automatically recorded
./rewind list      # ← see it in the list
```

**Supported shells**: bash, zsh, fish

**Why this matters**: Most people forget to use `rewind run`. Auto-recording means you get the benefits without changing your workflow.

**Control it anytime**:
```bash
export REWIND_ENABLED=false   # Disable
export REWIND_ENABLED=true    # Re-enable
```

---

## 📖 Command Reference

### Recording
```bash
rewind run <command> [args...]     # Record and execute a command
```

### Database & Shell History
```bash
rewind migrate                     # Migrate JSON sessions to SQLite
rewind search <query>              # Full-text search across sessions
rewind history [limit]             # View tracked shell history (default: 20)
rewind import-history [shell|path] # Import from shell history files
rewind history-stats               # Usage statistics (total, unique, top commands)
rewind track-command <cmd> [code]  # Auto-track via shell hooks (internal)
```

### Analysis
```bash
rewind replay <session_id>         # Replay session frame-by-frame
rewind stats <session_id>          # Show duration, exit code, event count
rewind timeline <session_id>       # Visualize timeline with annotations
rewind inspect <session_id>        # Print full event details
rewind score <session_id>          # Rate session (0-100)
rewind detect <session_id>         # Find loops and anomalies
rewind diff <id1> <id2>            # Compare two sessions
```

### AI Features
```bash
rewind web [port]                  # Start the Rewind web UI
rewind recall <query>              # Search via embeddings across all sessions
rewind chat <model>                # Start interactive chat with memory

# Example queries
rewind web 8080
rewind recall "database migration"
rewind recall "connection timeout"
rewind chat qwen2.5:1.5b
```

### Export
```bash
rewind export <session_id>         # Generate HTML report
rewind markdown <session_id>       # Export as Markdown
rewind list                        # List all sessions
```

---

## 🗄️ Database Backend

Rewind uses SQLite as its primary storage backend for sessions, events, and shell history.

### Schema

```sql
-- Sessions table (main session metadata)
sessions (
    id TEXT PRIMARY KEY,
    command TEXT NOT NULL,
    model TEXT,
    title TEXT,
    summary TEXT,
    tags TEXT,            -- JSON array
    mood TEXT,
    started_at DATETIME,
    ended_at DATETIME,
    exit_code INTEGER,
    created_at DATETIME
)

-- Events table (detailed event tracking)
events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT,          -- FK → sessions.id
    timestamp DATETIME,
    type TEXT,                -- stdout, stderr, exit, etc.
    content TEXT
)

-- Shell history table (system shell commands)
shell_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    command TEXT NOT NULL,
    exit_code INTEGER,
    working_dir TEXT,
    executed_at DATETIME,
    session_id TEXT           -- FK → sessions.id (optional)
)
```

### Storage Backend Selection

| Variable | Effect |
|----------|--------|
| (default) | Uses SQLite (`rewind.db`) |
| `REWIND_DB_PATH=/custom/path.db` | Custom SQLite database path |
| `REWIND_USE_JSON=true` | Force JSON file storage (legacy mode) |

The SQLite database uses **WAL journal mode** for better concurrent performance and **foreign key constraints** for data integrity.

### JSON Legacy Support

All existing `sessions/*.json` files continue to work. The JSON loader (`internal/storage/load.go`) remains as a fallback. Set `REWIND_USE_JSON=true` to force the old behavior.

---

## 🐚 System Shell History Integration

Rewind bridges the gap between `bash_history` and full session recording.

### What Gets Tracked

| Field | Description |
|-------|-------------|
| `command` | Full command text (e.g., `git push origin main`) |
| `exit_code` | Command exit status (0 = success) |
| `working_dir` | Directory where command was executed |
| `executed_at` | Timestamp (UTC) |
| `session_id` | Linked session ID (if command was part of a session) |

### Importing Existing History

```bash
# Supported formats:
#   bash  — ~/.bash_history (plain text, one command per line)
#   zsh   — ~/.zsh_history (extended format with timestamps)
#   fish  — ~/.local/share/fish/fish_history (YAML format)

rewind import-history bash        # Import bash history
rewind import-history zsh         # Import zsh history
rewind import-history fish        # Import fish history
rewind import-history /my/file    # Import custom file
rewind import-history             # Auto-detect all available
```

### Viewing & Searching

```bash
rewind history                    # Last 20 commands
rewind history 100                # Last 100 commands
rewind history-stats              # Stats (total, unique, top 10)
rewind search "git push"          # Search across sessions
```

### How Shell Hooks Work

The `rewind setup` command generates shell-specific hook code:

- **bash**: Uses `DEBUG` trap + `PROMPT_COMMAND` for pre/post-execution tracking
- **zsh**: Uses `preexec` and `precmd` hook functions
- **fish**: Uses `fish_postexec` event handler

Each hook calls `rewind track-command` in the background (non-blocking), which saves the command to both the sessions directory and the SQLite `shell_history` table.

---

## 💬 Chat with Memory

Rewind's chat engine automatically searches your session history and injects relevant context.

```bash
./rewind chat qwen2.5:1.5b
Starting chat with model: qwen2.5:1.5b
Type 'exit' to quit
Type 'recall' to test memory recall

You: What deployment issues did I have last week?
[Using memory context from previous sessions]
1. [Session abc123] Connection timeout on database migration
2. [Session def456] Docker image build failed due to npm
3. [Session ghi789] Kubernetes deployment rollback
Assistant: Based on your previous sessions, you had three main issues...
```

### How It Works

1. **Query → Embedding** — Your question is embedded using `nomic-embed-text`
2. **Search Cache** — Searches cached embeddings (⚡ 900x faster than first run)
3. **Top-3 Results** — Ranks sessions by semantic similarity
4. **Context Injection** — Adds relevant excerpts to the prompt
5. **Ollama Generation** — Uses your local LLM to generate response

---

## 📊 Data Format

### JSON Sessions (legacy)

Sessions are stored as JSON in `sessions/`:

```json
{
  "id": "1779265115474389394_585791",
  "command": "echo 'Hello'",
  "title": "Greeting test",
  "summary": "Simple echo test with output capture",
  "started_at": "2025-05-21T10:30:15Z",
  "ended_at": "2025-05-21T10:30:16Z",
  "exit_code": 0,
  "events": [
    {
      "timestamp": "2025-05-21T10:30:15.123Z",
      "type": "stdout",
      "content": "Hello\n"
    }
  ]
}
```

### SQLite Database (primary)

Sessions, events, and shell history are stored in `rewind.db` (or `REWIND_DB_PATH`). Use `sqlite3` to explore directly:

```bash
sqlite3 rewind.db "SELECT id, command, exit_code FROM sessions ORDER BY started_at DESC LIMIT 5"
sqlite3 rewind.db "SELECT command, COUNT(*) c FROM shell_history GROUP BY command ORDER BY c DESC LIMIT 10"
```

Embeddings cached in `sessions/{id}_embeddings.json` for instant recall on repeated queries.

---

## ⚡ Performance

### Embedding Cache

First query on 18 sessions × 50 events:
```
Without cache:  900 HTTP calls to Ollama → 30-120 seconds
With cache:     1 disk read + 1 query call → <1 second (cached)
Speedup:        900x
```

### Benchmarks (i7-4765T, 8GB RAM)

| Operation | Time | Notes |
|-----------|------|-------|
| Record 100-line script | 50ms | Real-time capture |
| First `recall` (18 sessions) | 45s | Embeddings computed once |
| Second `recall` (same query) | 200ms | Cached embeddings loaded |
| `chat` with context | 3-5s | Query + context injection + generation |
| `export` HTML report | 200ms | Full session HTML |
| SQLite `search` (1K sessions) | <10ms | Indexed LIKE query |
| `migrate` 25 JSON → SQLite | <100ms | Bulk insert in transaction |
| `import-history` 1500 commands | <500ms | Prepared statement batch |

---

## 🏗️ Architecture

```
rewind/
├── cmd/rewind/
│   └── main.go              # CLI entry point (SQLite-aware)
├── internal/
│   ├── storage/
│   │   ├── interface.go     # Storage & ShellHistoryStorage interfaces
│   │   ├── sqlite.go        # SQLiteStore implementation (both interfaces)
│   │   ├── migrate.go       # JSON → SQLite migration
│   │   ├── storage.go       # JSON save (legacy)
│   │   ├── load.go          # JSON load (legacy)
│   │   └── list.go          # JSON list (legacy)
│   ├── shellhistory/
│   │   ├── history.go       # Shell history manager (track, view, stats)
│   │   └── import.go        # Import parsers (bash, zsh, fish)
│   ├── recorder/            # Execute and capture output
│   ├── replay/              # Frame-by-frame playback
│   ├── chat/                # Ollama HTTP client + memory integration
│   ├── memory/              # Embeddings (v2.go) + caching
│   ├── summarize/           # Structured summaries
│   ├── export/              # HTML/Markdown generation
│   ├── stats/               # Analysis and scoring
│   ├── shell/               # Shell hook generators (bash/zsh/fish)
│   └── ... (14 more packages)
└── pkg/types/
    └── session.go           # Session, Event, Message data models
```

### Key Design Decisions

- **SQLite as primary storage** — Relational queries, foreign keys, WAL mode, zero-CGO
- **Storage interface** — `Storage` and `ShellHistoryStorage` abstractions enable pluggable backends
- **JSON backward compatibility** — All legacy loaders preserved; toggle with `REWIND_USE_JSON`
- **Local LLMs only** — No external API calls, full privacy
- **Embedding cache** — JSON-based for simplicity and portability
- **Modular CLI** — Each command is independent, can be used in scripts
- **Frame-by-frame replay** — Faithful reconstruction of session timing

---

## 🔍 Use Cases

### 1. **Debug Production Issues**
```bash
rewind recall "connection timeout"
# Get exact commands and output from similar issues
rewind chat qwen2.5:1.5b  # Ask about the pattern
```

### 2. **Learn from Session History**
```bash
rewind timeline my-session-id
rewind inspect my-session-id
```

### 3. **Understand Your Shell Habits**
```bash
rewind history-stats       # What commands do you run most?
rewind history 50          # Browse recent activity
rewind import-history bash # Import years of history
```

### 4. **Improve Shell Workflows**
```bash
rewind score session-id     # Find high-quality sessions
rewind diff good-session bad-session  # Compare approaches
```

### 5. **Automated Reports**
```bash
for id in $(rewind list | cut -d' ' -f1); do
  rewind export $id
done
# Generate HTML reports for all sessions
```

### 6. **Knowledge Base**
Store and search your own terminal recipes using SQLite queries or embeddings. Perfect for teams managing common troubleshooting procedures.

---

## 🛠️ Development

### Run Tests
```bash
go test ./...
```

### Build Distribution
```bash
go build -o rewind ./cmd/rewind
```

### Database Inspection
```bash
# Explore SQLite directly
sqlite3 rewind.db ".tables"
sqlite3 rewind.db "SELECT id, command FROM sessions LIMIT 5"
sqlite3 rewind.db "SELECT command, COUNT(*) FROM shell_history GROUP BY command ORDER BY 2 DESC LIMIT 10"
```

### Extend with New Analyzers
See `internal/stats/`, `internal/detect/`, `internal/score/` for examples of adding custom analysis.

---

## 📝 Configuration

No configuration file needed. Rewind works out of the box:

### Storage Backend
| Variable | Default | Description |
|----------|---------|-------------|
| (none) | SQLite `rewind.db` | Primary storage (recommended) |
| `REWIND_DB_PATH` | `rewind.db` | Custom database path |
| `REWIND_USE_JSON` | `false` | Force JSON file storage |

### Ollama
- **Endpoint**: `http://localhost:11434` (configurable in code)
- **Models**: `qwen2.5:1.5b` (summarization), `nomic-embed-text` (embeddings)

To use different models, edit the model names in:
- `internal/chat/chat.go`
- `internal/summarize/summarize.go`

---

## 🤝 Contributing

Issues and PRs welcome. Key areas:

- Web UI enhancements for session browser
- FTS5 full-text search in SQLite
- Multi-model support (Claude, GPT, etc via HTTP APIs)
- Windows support
- Keyboard shortcuts in chat
- Plugin system for custom analyzers

---

## 📄 License

MIT License — see [LICENSE](LICENSE) for details.

---

## 💡 Why Rewind?

Traditional shell history (`bash_history`) shows commands, not results. You lose:
- Full output context
- Session flow and timing
- Interactions and failures
- Ability to search by what actually happened

Rewind captures everything, indexes it intelligently, and makes it queryable with natural language. **Your terminal becomes a searchable knowledge base.**

With SQLite storage and shell history integration, Rewind now bridges the gap between lightweight shell history and full session recording—giving you the best of both worlds.

---

## 🚦 Status

- [x] Core recording/replay
- [x] Analysis commands (stats, diff, detect, score)
- [x] Embedding-based search with cache
- [x] Chat with memory injection
- [x] Structured summaries
- [x] Streaming responses (real-time output)
- [x] Shell integration (auto-record with hooks)
- [x] Web UI for session browser
- [x] **Database backend (SQLite)**
- [x] **System shell history integration**

---

**Built for developers who want to learn from their terminal history.**

Questions? Open an issue or [join the discussion](https://github.com/habeldavidson007-glitch/rewind/discussions).
