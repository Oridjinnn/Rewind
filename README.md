# 🔄 Rewind

**Local AI memory for developers. Records your terminal, understands your IDE, recalls what matters.**

[![CI](https://github.com/Oridjinnn/Rewind/actions/workflows/ci.yml/badge.svg)](https://github.com/Oridjinnn/Rewind/actions/workflows/ci.yml)
[![Go 1.22+](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Stars](https://img.shields.io/github/stars/Oridjinnn/Rewind?style=social)](https://github.com/Oridjinnn/Rewind/stargazers)

> **OpenViking manages memory for your AI agents. Rewind manages memory for *you* — the developer running commands in the terminal.**

![Demo](demo.gif)

---

## What is Rewind?

Every developer loses context. You run 50 commands debugging a problem, close the terminal, and tomorrow you can't remember what worked. You switch between VS Code and the shell and nothing connects. You ask an AI the same question you asked last week because you forgot the answer.

Rewind fixes this. It runs **100% locally** — no cloud, no API keys, no subscriptions. Just a single binary that records what you do, stores it in SQLite, and lets you recall it with natural language powered by Ollama.

```
$ rewind run docker build -t myapp .
● Recording... [exit 1] 2.3s

$ rewind chat qwen2.5:1.5b
> why did my docker build fail yesterday?
↳ Searching 47 sessions... found 3 relevant

[Session 2h ago] docker build failed: COPY failed, file not found
The build tried to COPY ./dist but the dist/ folder wasn't created yet.
You need to run `npm run build` first.
```

---

## Features

### 🎬 Record Everything
- **`rewind run <cmd>`** — record any command with full stdout/stderr capture
- **`rewind setup`** — install shell hooks (bash/zsh/fish) for automatic recording
- **`rewind import-history`** — import existing shell history (bash/zsh/fish)

### 🔍 Recall & Search
- **`rewind recall <query>`** — semantic search across all sessions via embeddings
- **`rewind search <keyword>`** — fast full-text search
- **`rewind chat <model>`** — chat with your session history as context

### 📊 Inspect & Analyze
- **`rewind list`** — list all recorded sessions
- **`rewind replay <id>`** — replay a session frame by frame
- **`rewind timeline <id>`** — visualize event timeline
- **`rewind inspect <id>`** — detailed event breakdown
- **`rewind stats <id>`** — execution quality stats
- **`rewind score <id>`** — quality score with indicators
- **`rewind diff <id1> <id2>`** — compare two sessions
- **`rewind detect <id>`** — detect anomalies and patterns

### 🔌 IDE Integration
Works with **VS Code**, **JetBrains**, and **Neovim** via a unified JSON-RPC server.

```bash
rewind ide start              # start server (default port: REWIND_IDE_PORT)
rewind ide permissions vscode on   # enable recording for VS Code
rewind ide activity vscode 20      # show recent activity
rewind ide analyze <project>       # AI analysis of project activity
```

### 📤 Export
- **`rewind export <id>`** — export to HTML
- **`rewind markdown <id>`** — export to Markdown
- **`rewind web`** — browse all sessions in the browser

---

## Quick Start

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Ollama](https://ollama.com/) (for AI features)

### Install

```bash
git clone https://github.com/Oridjinnn/Rewind.git
cd Rewind
go build -o rewind ./cmd/rewind
sudo mv rewind /usr/local/bin/
```

Or download a pre-built binary from [Releases](https://github.com/Oridjinnn/Rewind/releases).

### Pull AI models

```bash
ollama pull qwen2.5:1.5b      # for chat and summarization
ollama pull nomic-embed-text   # for semantic recall
```

### First run

```bash
# Record a command
rewind run ls -la

# Set up automatic shell recording
rewind setup

# Try semantic recall
rewind recall "list files"

# Start chatting with your history
rewind chat qwen2.5:1.5b
```

### Migrate from older versions

```bash
rewind migrate   # migrate JSON sessions to SQLite
```

---

## Configuration

All settings via environment variables — no config files needed.

| Variable | Default | Description |
|---|---|---|
| `REWIND_OLLAMA_HOST` | `http://localhost:11434` | Ollama server URL |
| `REWIND_CHAT_MODEL` | `qwen2.5:1.5b` | Model for chat |
| `REWIND_SUMMARIZE_MODEL` | `qwen2.5:1.5b` | Model for session summarization |
| `REWIND_EMBED_MODEL` | `nomic-embed-text` | Model for embeddings |
| `REWIND_IDE_PORT` | `9876` | IDE event server port |

Example — use a different model or remote Ollama:

```bash
export REWIND_OLLAMA_HOST=http://192.168.1.10:11434
export REWIND_CHAT_MODEL=llama3.2:3b
rewind chat llama3.2:3b
```

---

## How It Works

```
┌─────────────────────────────────────────────────────────┐
│                        Your Machine                      │
│                                                         │
│  Terminal  ──►  Recorder  ──►  SQLite  ──►  Recall      │
│  VS Code   ──►  IDE Server ──►  (local) ──►  Chat       │
│  JetBrains ──►  Bridge    ──►          ──►  Export      │
│  Neovim    ──►            ──►          ──►  Web UI      │
│                                    ▲                    │
│                             Ollama (local)               │
│                          embeddings + chat               │
└─────────────────────────────────────────────────────────┘
```

- **Recorder** captures command execution with stdout/stderr, exit codes, timing
- **Cleaner** strips ANSI/spinner noise before storing
- **Redact** detects and removes secrets (API keys, tokens, passwords) before saving
- **SQLite + WAL** stores everything fast with 11 indexes
- **Embeddings** via Ollama `nomic-embed-text` — your data never leaves your machine
- **Recall** combines cosine similarity with recency decay for relevant results

---

## Privacy & Security

Rewind takes privacy seriously:

- ✅ **Zero cloud** — everything stays on your machine
- ✅ **Secret redaction** — detects GitHub PAT, AWS keys, OpenAI tokens, Slack tokens, and 8 more patterns before saving
- ✅ **Opt-in recording** — IDE recording is disabled by default, enabled per-project
- ✅ **Local embeddings** — vectors are generated by Ollama on your machine

---

## IDE Extensions

| Extension | Status | Install |
|---|---|---|
| VS Code | ✅ Available | See `extensions/vscode/` |
| JetBrains | ✅ Available | See `extensions/jetbrains/` |
| Neovim | ✅ Available | See `extensions/nvim/` |

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) — includes setup, how to run tests, and good first issues.

```bash
go test ./...   # run all tests
go vet ./...    # lint
```

---

## Roadmap

- [ ] `rewind sync` — optional encrypted backup to S3/R2
- [ ] Hosted web dashboard at rewind.sh
- [ ] GitHub Actions integration — record CI runs
- [ ] MCP server — expose Rewind memory to Claude Code, Cursor, and other AI tools

---

## License

MIT — see [LICENSE](LICENSE).

---

<p align="center">
  <sub>Built by <a href="https://github.com/Oridjinnn">Oridjinn</a> · 100% local · no telemetry · no cloud</sub>
</p>
