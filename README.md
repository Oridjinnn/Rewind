# 🔄 Rewind

**Any command. Any session. Any question. Rewind knows.**

[![CI](https://github.com/Oridjinnn/Rewind/actions/workflows/ci.yml/badge.svg)](https://github.com/Oridjinnn/Rewind/actions/workflows/ci.yml)
[![Go 1.22+](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Stars](https://img.shields.io/github/stars/Oridjinnn/Rewind?style=social)](https://github.com/Oridjinnn/Rewind/stargazers)

Rewind is a **local-first AI memory layer** for developers. It records your terminal sessions and IDE activity, then lets you recall, replay, and chat with your history — powered entirely by Ollama on your own machine.

No cloud. No API keys. No subscriptions. Just a single binary.

![Demo](demo.gif)

---

## The problem

You run 50 commands debugging a problem, close the terminal, and tomorrow you can't remember what worked. You ask an AI the same question you asked last week. Nothing connects your shell history to your IDE to your conversations.

Rewind fixes that.

```
$ rewind run docker build -t myapp .
● Recording... [exit 1] 2.3s

$ rewind chat qwen2.5:1.5b
> why did my docker build fail yesterday?
↳ Searching 47 sessions... found 3 relevant

[2h ago] docker build failed: COPY failed, file not found
The build tried to COPY ./dist but the folder didn't exist yet.
Run `npm run build` first, then retry the build.
```

---

## Quick start

### Prerequisites
- [Go 1.22+](https://go.dev/dl/)
- [Ollama](https://ollama.com/)

### Install

```bash
git clone https://github.com/Oridjinnn/Rewind.git
cd Rewind
go build -o rewind ./cmd/rewind
sudo mv rewind /usr/local/bin/
```

Or grab a pre-built binary from [Releases](https://github.com/Oridjinnn/Rewind/releases).

### Pull models

```bash
ollama pull qwen2.5:1.5b       # chat + summarization
ollama pull nomic-embed-text    # semantic recall
```

### First run

```bash
rewind run ls -la               # record a command
rewind setup                    # install shell hooks (auto-record everything)
rewind recall "list files"      # semantic search
rewind chat qwen2.5:1.5b        # chat with your history
```

---

## Features

### 🎬 Record
| Command | What it does |
|---|---|
| `rewind run <cmd>` | Record any command with full output capture |
| `rewind setup` | Install shell hooks for bash / zsh / fish |
| `rewind import-history` | Import existing shell history |
| `rewind track-command` | Track a single command manually |

### 🔍 Recall & Search
| Command | What it does |
|---|---|
| `rewind recall <query>` | Semantic search via embeddings |
| `rewind search <keyword>` | Fast full-text search |
| `rewind chat <model>` | Chat with session history as context |

### 📊 Inspect & Analyze
| Command | What it does |
|---|---|
| `rewind list` | List all recorded sessions |
| `rewind replay <id>` | Replay session frame by frame |
| `rewind timeline <id>` | Visual event timeline |
| `rewind inspect <id>` | Detailed event breakdown |
| `rewind stats <id>` | Execution quality stats |
| `rewind score <id>` | Quality score with indicators |
| `rewind diff <id1> <id2>` | Compare two sessions |
| `rewind detect <id>` | Detect anomalies and patterns |

### 🔌 IDE Integration
Works with **VS Code**, **JetBrains**, and **Neovim** via a unified JSON-RPC server.

```bash
rewind ide start                        # start server
rewind ide permissions vscode on        # enable for VS Code (per-project opt-in)
rewind ide activity vscode 20           # recent activity
rewind ide analyze <project>            # AI project analysis
rewind ide projects                     # list tracked projects
```

### 📤 Export
| Command | What it does |
|---|---|
| `rewind export <id>` | Export session to HTML |
| `rewind markdown <id>` | Export session to Markdown |
| `rewind web` | Browse all sessions in browser |

---

## How it works

```
  Terminal ──►┐
  VS Code  ──►├──► Recorder ──► SQLite (local) ──► Recall
  JetBrains──►│               ── WAL mode        ── Chat
  Neovim   ──►┘               ── 11 indexes      ── Web UI
                                      ▲
                               Ollama (local)
                            embeddings + summaries
```

- **Recorder** captures stdout/stderr, exit codes, and timing
- **Cleaner** strips ANSI/spinner noise before storing
- **Redact** removes secrets (API keys, tokens) automatically before saving
- **Embeddings** generated locally by Ollama — your data never leaves your machine
- **Recall** combines cosine similarity with recency decay

---

## Privacy

- ✅ Zero cloud — everything on your machine
- ✅ Secret redaction — GitHub PAT, AWS keys, OpenAI tokens, Slack tokens, and more detected automatically
- ✅ Opt-in IDE recording — disabled by default, enabled per-project
- ✅ Local embeddings — vectors never leave your device

---

## Configuration

| Variable | Default | Description |
|---|---|---|
| `REWIND_OLLAMA_HOST` | `http://localhost:11434` | Ollama server URL |
| `REWIND_CHAT_MODEL` | `qwen2.5:1.5b` | Model for chat |
| `REWIND_SUMMARIZE_MODEL` | `qwen2.5:1.5b` | Model for summarization |
| `REWIND_EMBED_MODEL` | `nomic-embed-text` | Model for embeddings |
| `REWIND_IDE_PORT` | `9876` | IDE server port |

```bash
# Example: use a remote Ollama or different model
export REWIND_OLLAMA_HOST=http://192.168.1.10:11434
export REWIND_CHAT_MODEL=llama3.2:3b
```

---

## IDE Extensions

| IDE | Status | Location |
|---|---|---|
| VS Code | ✅ | `extensions/vscode/` |
| JetBrains | ✅ | `extensions/jetbrains/` |
| Neovim | ✅ | `extensions/nvim/` |

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

```bash
go test ./...    # run all tests
go vet ./...     # lint
```

Good first issues are labeled [`good first issue`](https://github.com/Oridjinnn/Rewind/issues?q=label%3A%22good+first+issue%22) on the issue tracker.

---

## Roadmap

- [ ] `rewind sync` — optional encrypted backup to S3 / Cloudflare R2
- [ ] MCP server — expose Rewind memory to Claude Code, Cursor, and other AI tools
- [ ] Hosted web dashboard
- [ ] GitHub Actions integration — record CI runs

---

## License

MIT — see [LICENSE](LICENSE).

---

<p align="center">
  <sub>Built by <a href="https://github.com/Oridjinnn">Oridjinn</a> · 100% local · no telemetry · no cloud</sub>
</p>
