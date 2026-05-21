# 🔄 Rewind

**Record, replay, and recall your terminal sessions with AI-powered memory.**

Rewind captures every command execution, output, and interaction—then lets you search, analyze, and chat about your history using embeddings and local LLMs.

[![Go 1.21+](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go)](https://golang.org)
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
- **`rewind recall <query>`** — Search across all sessions using semantic embeddings
- **`rewind chat <model>`** — Interactive chat with Ollama, using session context
- Automatic context injection from relevant past sessions
- Embedding cache for 900x speedup on repeated queries

### Export & Sharing
- **`rewind export <session_id>`** — Generate HTML report
- **`rewind markdown <session_id>`** — Export as Markdown with formatting
- **`rewind list`** — Show all recorded sessions

---

## 🚀 Quick Start

### Prerequisites

1. **Go 1.21+** — [Install](https://golang.org/doc/install)
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

---

## 📖 Command Reference

### Recording
```bash
rewind run <command> [args...]     # Record and execute a command
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
rewind recall <query>              # Search via embeddings across all sessions
rewind chat <model>                # Start interactive chat with memory

# Example queries
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
    },
    {
      "timestamp": "2025-05-21T10:30:15.124Z",
      "type": "exit",
      "content": "0"
    }
  ]
}
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

Cache is automatically updated when new events are added to a session.

### Benchmarks (i7-4765T, 8GB RAM)

| Operation | Time | Notes |
|-----------|------|-------|
| Record 100-line script | 50ms | Real-time capture |
| First `recall` (18 sessions) | 45s | Embeddings computed once |
| Second `recall` (same query) | 200ms | Cached embeddings loaded |
| `chat` with context | 3-5s | Query + context injection + generation |
| `export` HTML report | 200ms | Full session HTML |

---

## 🏗️ Architecture

```
rewind/
├── cmd/rewind/
│   └── main.go           # CLI entry point
├── internal/
│   ├── recorder/         # Execute and capture output
│   ├── replay/           # Frame-by-frame playback
│   ├── chat/             # Ollama HTTP client + memory integration
│   ├── memory/           # Embeddings (v2.go) + caching
│   ├── summarize/        # Structured summaries (key points, entities, sentiment)
│   ├── export/           # HTML/Markdown generation
│   ├── stats/            # Analysis and scoring
│   └── ... (14 more packages)
└── pkg/types/
    └── session.go        # Data models
```

### Key Design Decisions

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

### 3. **Improve Shell Workflows**
```bash
rewind score session-id     # Find high-quality sessions
rewind diff good-session bad-session  # Compare approaches
```

### 4. **Automated Reports**
```bash
for id in $(rewind list | cut -d' ' -f1); do
  rewind export $id
done
# Generate HTML reports for all sessions
```

### 5. **Knowledge Base**
Store and search your own terminal recipes using embeddings. Perfect for teams managing common troubleshooting procedures.

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

### Extend with New Analyzers
See `internal/stats/`, `internal/detect/`, `internal/score/` for examples of adding custom analysis.

---

## 📝 Configuration

No configuration file needed. Rewind works out of the box:

- **Sessions dir**: `./sessions/` (auto-created)
- **Ollama endpoint**: `http://localhost:11434` (configurable in code)
- **Models**: `qwen2.5:1.5b` (summarization), `nomic-embed-text` (embeddings)

To use different models, edit the model names in:
- `internal/chat/chat.go` (line with `QueryOllama()`)
- `internal/summarize/summarize.go` (model names in functions)

---

## 🤝 Contributing

Issues and PRs welcome. Key areas:

- Streaming responses (right now responses are buffered)
- Sqlite backend for embeddings (vs JSON)
- Web UI for session browser
- Integration with shell history
- Windows support

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

Rewind captures everything, indexes it intelligently, and makes it queryable with natural language. Your terminal becomes a searchable knowledge base.

---

## 🚦 Status

- [x] Core recording/replay
- [x] Analysis commands (stats, diff, detect, score)
- [x] Embedding-based search with cache
- [x] Chat with memory injection
- [x] Structured summaries
- [ ] Streaming responses
- [ ] Web UI
- [ ] Shell integration
- [ ] Database backend

---

**Built for developers who want to learn from their terminal history.**

Questions? Open an issue or [join the discussion](https://github.com/habeldavidson007-glitch/rewind/discussions).
