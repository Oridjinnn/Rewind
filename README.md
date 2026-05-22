# Rewind

**Rewind** is a local-first cognitive memory runtime for small LLMs—turning your command logs, IDE activity, and AI conversations into searchable, reusable context.

- **Local-first**: embeddings, ranking, and summaries run on your machine.
- **SQLite-backed**: fast indexed storage with WAL mode.
- **Multi-IDE**: a unified JSON-RPC protocol for VS Code / JetBrains / Neovim.
- **Opt-in privacy**: recording is disabled by default.

![Demo](demo.gif)

---

## Key ideas

Rewind captures what you do, then recalls the most relevant parts when you ask a question.

- **Smart Recall**: embedding-based semantic search + recency decay.
- **Cross-reference**: connects IDE events with shell history via sessions.
- **Deterministic behavior for safety**: configuration centralized in code, defaults via env.

---

## Features

### Recording
- `rewind run <command>` – record command execution
- `rewind setup` – install shell hooks (bash/zsh/fish)

### Replay & inspection
- `rewind replay <session_id>` – replay frames
- `rewind timeline <session_id>` – timeline view
- `rewind inspect <session_id>` – detailed event list
- `rewind diff <id1> <id2>` – compare sessions
- `rewind stats <session_id>` – stats and quality indicators

### Memory & AI
- `rewind recall <query>` – embedding-based recall
- `rewind chat <model>` – chat with injected session context

### IDE integration
- `rewind ide start [port]` – start IDE event server (default: `REWIND_IDE_PORT`)
- `rewind ide permissions <ide> on|off [project]`
- `rewind ide activity [ide] [limit]`
- `rewind ide projects`
- `rewind ide analyze <project>`

---

## Quick start

### Prerequisites
- **Go 1.22+**
- **Ollama** (for AI features)
- **SQLite** (bundled)

Pull models:
```bash
ollama pull qwen2.5:1.5b
ollama pull nomic-embed-text
```

### Build
```bash
git clone https://github.com/Oridjinnn/Rewind.git
cd Rewind

go build ./cmd/rewind
```

### Migrate existing sessions
```bash
./rewind migrate
```

### Enable shell auto-recording
```bash
./rewind setup
```

### Start IDE server (for extensions)
```bash
./rewind ide start
```

Enable recording for an IDE:
```bash
rewind ide permissions vscode on
```

---

## Configuration (env)

| Variable | Default | Purpose |
|---|---:|---|
| `REWIND_CHAT_MODEL` | `qwen2.5:1.5b` | Ollama model for chat/reflection |
| `REWIND_SUMMARIZE_MODEL` | `qwen2.5:1.5b` | Ollama model for summarization |
| `REWIND_EMBED_MODEL` | `nomic-embed-text` | Ollama model for embeddings |
| `REWIND_OLLAMA_HOST` | `http://localhost:11434` | Ollama base URL |
| `REWIND_IDE_PORT` | `9876` | IDE event server port |

---

## Embeddings & recall model

- Embeddings are generated through Ollama using `REWIND_EMBED_MODEL`.
- Recall combines **cosine similarity** with a **recency decay** factor.

---

## Demo

Generate the demo GIF:
```bash
bash ./make_demo.sh
```

---

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md).

---

## License

MIT License. See [LICENSE](LICENSE).

