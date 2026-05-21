# Rewind IDE — VS Code Extension

**Record, replay, and recall your IDE & terminal activity with AI-powered memory.**

Track file edits, terminal commands, git operations, builds, tests, and AI assistant interactions. Everything is stored locally in SQLite — fully private, no cloud.

Part of the [Rewind](https://github.com/habeldavidson007-glitch/rewind) ecosystem. Works with VS Code and Cursor.

---

## 🔥 Features

| What's Recorded | Example Events |
|----------------|----------------|
| 📁 File operations | Open, save, close, create, delete |
| 🖥️ Terminal | Open, close, exit codes |
| 🏗️ Build & Test | Start, end, pass, fail, errors |
| 🐙 Git | Branches, commits, state changes |
| 🤖 AI Assistant | Copilot completions, Cursor AI, chat interactions |
| ✏️ Editor activity | Cursor positions, file edits |

**28 event types** — every action becomes searchable data.

---

## 📦 Install

### From GitHub Releases
1. Download the latest `.vsix` from [GitHub Releases](https://github.com/habeldavidson007-glitch/rewind/releases)
2. In VS Code: Extensions → `...` → Install from VSIX...

### From Source (Development)

```bash
cd extensions/vscode
npm install
npm run compile
```

Then in VS Code: `Ctrl+Shift+P` → `Developer: Install Extension from Location...` → select `extensions/vscode/`

---

## 🚀 Quick Start

### Prerequisites
- **Go 1.22+** installed on your system
- Build the rewind CLI: `go build -o rewind ./cmd/rewind` (from repo root)

### Step 1: Start the Rewind Recording Server

```bash
rewind ide start
# → Listening on http://localhost:9876
```

### Step 2: Opt-In for Recording (Privacy First!)

Recording is **disabled by default**. Enable it:

```bash
rewind ide permissions vscode on
# → Recording enabled for VS Code
```

### Step 3: Check Your Activity

```bash
rewind ide activity vscode 10
# → Shows your last 10 IDE events
rewind ide status
# → Shows server status, event count, permissions
```

---

## ⚙️ Configuration

Open VS Code Settings (`Ctrl+,`) and search for `rewind.*`:

| Setting | Default | Description |
|---------|---------|-------------|
| `rewind.serverPort` | `9876` | Rewind server port |
| `rewind.recordFileEdits` | `true` | Record file save/edit events |
| `rewind.recordTerminalCommands` | `true` | Record terminal commands |
| `rewind.recordBuildAndTest` | `true` | Record build and test events |
| `rewind.recordGitOperations` | `true` | Record git operations |
| `rewind.recordAIAssistant` | `true` | Record Copilot/Cursor interactions |
| `rewind.recordContentSnapshots` | `false` | Record file contents (for rich AI context) |
| `rewind.autoSendBatch` | `true` | Batch-send events every 5s |
| `rewind.batchIntervalMs` | `5000` | Batch interval in milliseconds |

---

## 🔒 Privacy

All recording is **opt-in** — nothing is recorded until you explicitly enable it:
```bash
rewind ide permissions vscode on
```

All data is stored **locally** in `rewind.db`. No telemetry. No cloud. Full local control.

---

## 🧠 Supported IDEs

This extension works with:
- **VS Code** — Full feature set
- **Cursor** — Auto-detected, full compatibility
- All editors based on VS Code's extension API

For other IDEs (IntelliJ, Neovim, etc.), see the [extensions directory](https://github.com/habeldavidson007-glitch/rewind/tree/main/extensions).

---

## 📋 Commands

| Command | Description |
|---------|-------------|
| `Rewind: Toggle Recording On/Off` | Enable/disable recording |
| `Rewind: Show Recording Status` | Display current recording status |
| `Rewind: View Recent Activity` | Show last 20 IDE events |
| `Rewind: View Tracked Projects` | List all tracked IDE projects |
| `Rewind: Analyze Current Project` | Productivity insights |
| `Rewind: Search Shell History` | Search across shell commands |
| `Rewind: Import Shell History` | Import bash/zsh/fish history |
| `Rewind: Manage Recording Permissions` | Opt-in settings UI |

---

## 🏗️ Architecture

```
VS Code Extension → POST / → Rewind Server (:9876) → SQLite (rewind.db)
```

Events are sent in batches every 5 seconds for performance. The server runs locally — no data leaves your machine.

---

## 📚 Related

- [Rewind CLI & Backend](https://github.com/habeldavidson007-glitch/rewind) — the main project
- [JetBrains Plugin](../jetbrains/) — Rewind for IntelliJ/GoLand/PyCharm
- [Neovim Plugin](../nvim/) — Rewind for Neovim

---

## 📄 License

MIT License — part of the [Rewind](https://github.com/habeldavidson007-glitch/rewind) project.