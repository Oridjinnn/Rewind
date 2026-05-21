"use strict";
/**
 * Rewind Recorder - hooks into all VS Code/Cursor events.
 * Captures file edits, terminal commands, git operations, builds, tests, and AI interactions.
 */
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.RewindRecorder = void 0;
const vscode = __importStar(require("vscode"));
class RewindRecorder {
    constructor(protocol) {
        this.batch = [];
        this.batchTimer = null;
        this.enabled = true;
        this.disposables = [];
        // Config flags
        this.recordFiles = true;
        this.recordTerminal = true;
        this.recordBuildTest = true;
        this.recordGit = true;
        this.recordAI = true;
        this.recordSnapshots = false;
        this.protocol = protocol;
        const config = vscode.workspace.getConfiguration('rewind');
        this.batchIntervalMs = config.get('batchIntervalMs', 5000);
        this.useBatch = config.get('autoSendBatch', true);
        this.reloadConfig();
    }
    reloadConfig() {
        const config = vscode.workspace.getConfiguration('rewind');
        this.recordFiles = config.get('recordFileEdits', true);
        this.recordTerminal = config.get('recordTerminalCommands', true);
        this.recordBuildTest = config.get('recordBuildAndTest', true);
        this.recordGit = config.get('recordGitOperations', true);
        this.recordAI = config.get('recordAIAssistant', true);
        this.recordSnapshots = config.get('recordContentSnapshots', false);
    }
    /** Start listening to all IDE events. */
    start() {
        this.attachFileWatchers();
        this.attachTerminalWatchers();
        this.attachTaskWatchers();
        this.attachGitWatchers();
        this.attachAIWatchers();
        this.attachEditorWatchers();
    }
    /** Stop all event listeners. */
    stop() {
        this.flushBatch();
        for (const d of this.disposables) {
            d.dispose();
        }
        this.disposables = [];
    }
    /** Toggle recording on/off. */
    setEnabled(on) {
        this.enabled = on;
    }
    getProjectInfo() {
        const folders = vscode.workspace.workspaceFolders;
        if (folders && folders.length > 0) {
            const root = folders[0];
            return {
                name: root.name,
                path: root.uri.fsPath
            };
        }
        return { name: 'unknown', path: '' };
    }
    getLanguage(fileName) {
        const ext = fileName.split('.').pop()?.toLowerCase() || '';
        const map = {
            ts: 'typescript', tsx: 'typescript', js: 'javascript', jsx: 'javascript',
            go: 'go', rs: 'rust', py: 'python', rb: 'ruby', java: 'java',
            kt: 'kotlin', swift: 'swift', c: 'c', cpp: 'cpp', h: 'c',
            cs: 'csharp', php: 'php', lua: 'lua', sql: 'sql',
            html: 'html', css: 'css', scss: 'scss', json: 'json',
            yaml: 'yaml', yml: 'yaml', xml: 'xml', md: 'markdown',
            sh: 'shell', bash: 'shell', dockerfile: 'docker', toml: 'toml'
        };
        return map[ext] || ext;
    }
    async queueEvent(type, extra = {}) {
        if (!this.enabled)
            return;
        const { name, path } = this.getProjectInfo();
        const event = {
            protocol: 'rewind-ide-v1',
            ide: 'vscode',
            version: vscode.version,
            project: name,
            project_path: path,
            event: {
                type,
                timestamp: new Date().toISOString(),
                ...extra
            }
        };
        if (this.useBatch) {
            this.batch.push(event);
            this.scheduleBatchFlush();
        }
        else {
            this.protocol.sendEvent(type, extra, name, path).catch(() => { });
        }
    }
    scheduleBatchFlush() {
        if (this.batchTimer)
            return;
        this.batchTimer = setTimeout(() => {
            this.flushBatch();
        }, this.batchIntervalMs);
    }
    flushBatch() {
        if (this.batchTimer) {
            clearTimeout(this.batchTimer);
            this.batchTimer = null;
        }
        if (this.batch.length === 0)
            return;
        const events = [...this.batch];
        this.batch = [];
        this.protocol.sendBatch(events).catch(() => { });
    }
    // --- File Watchers ---
    attachFileWatchers() {
        if (!this.recordFiles)
            return;
        // File save
        this.disposables.push(vscode.workspace.onDidSaveTextDocument((doc) => {
            this.queueEvent('file_save', {
                file: doc.uri.fsPath,
                language: doc.languageId,
                content_snapshot: this.recordSnapshots ? doc.getText() : undefined
            });
        }));
        // File open
        this.disposables.push(vscode.workspace.onDidOpenTextDocument((doc) => {
            this.queueEvent('file_open', {
                file: doc.uri.fsPath,
                language: doc.languageId
            });
        }));
        // File close
        this.disposables.push(vscode.workspace.onDidCloseTextDocument((doc) => {
            this.queueEvent('file_close', {
                file: doc.uri.fsPath,
                language: doc.languageId
            });
        }));
        // File create
        this.disposables.push(vscode.workspace.onDidCreateFiles((e) => {
            for (const file of e.files) {
                this.queueEvent('file_create', {
                    file: file.fsPath,
                    language: this.getLanguage(file.fsPath)
                });
            }
        }));
        // File delete
        this.disposables.push(vscode.workspace.onDidDeleteFiles((e) => {
            for (const file of e.files) {
                this.queueEvent('file_delete', { file: file.fsPath });
            }
        }));
    }
    // --- Terminal Watchers ---
    attachTerminalWatchers() {
        if (!this.recordTerminal)
            return;
        // Terminal open
        this.disposables.push(vscode.window.onDidOpenTerminal((terminal) => {
            this.queueEvent('terminal_cmd', {
                message: `Terminal opened: ${terminal.name}`,
                metadata: { terminal_name: terminal.name }
            });
        }));
        // Terminal close (exit code not exposed in VS Code API, best-effort)
        this.disposables.push(vscode.window.onDidCloseTerminal((terminal) => {
            this.queueEvent('terminal_cmd', {
                message: `Terminal closed: ${terminal.name}`,
                exit_code: terminal.exitStatus?.code,
                metadata: { terminal_name: terminal.name }
            });
        }));
    }
    // --- Task/Build/Test Watchers ---
    attachTaskWatchers() {
        if (!this.recordBuildTest)
            return;
        this.disposables.push(vscode.tasks.onDidStartTask((e) => {
            const task = e.execution.task;
            const type = task.source === 'Test' ? 'test_run' : 'build_start';
            this.queueEvent(type, {
                message: task.name,
                metadata: {
                    task_name: task.name,
                    task_source: task.source,
                    task_definition: task.definition
                }
            });
        }));
        this.disposables.push(vscode.tasks.onDidEndTask((e) => {
            const task = e.execution.task;
            const exitCode = e.execution?.exitCode;
            if (task.source === 'Test') {
                const type = exitCode === 0 ? 'test_pass' : 'test_fail';
                this.queueEvent(type, {
                    message: task.name,
                    exit_code: exitCode,
                    metadata: { task_name: task.name, duration_ms: 0 }
                });
            }
            else {
                const type = exitCode === 0 ? 'build_end' : 'build_error';
                this.queueEvent(type, {
                    message: task.name,
                    exit_code: exitCode,
                    metadata: { task_name: task.name }
                });
            }
        }));
    }
    // --- Git Watchers ---
    attachGitWatchers() {
        if (!this.recordGit)
            return;
        // VS Code's git extension API
        const gitExtension = vscode.extensions.getExtension('vscode.git');
        if (!gitExtension)
            return;
        gitExtension.activate().then((api) => {
            const git = api?.getAPI(1);
            if (!git)
                return;
            git.repositories.forEach((repo) => {
                repo.state.onDidChange(() => {
                    // Best-effort git operation detection from state changes
                    this.queueEvent('git_branch', {
                        message: `On branch: ${repo.state.HEAD?.name || 'unknown'}`,
                        metadata: {
                            branch: repo.state.HEAD?.name,
                            ahead: repo.state.HEAD?.ahead,
                            behind: repo.state.HEAD?.behind
                        }
                    });
                });
            });
        });
    }
    // --- AI Assistant Watchers (Copilot / Cursor) ---
    attachAIWatchers() {
        if (!this.recordAI)
            return;
        // Cursor AI events
        const cursorAI = vscode.extensions.getExtension('cursor.cursor-ai');
        if (cursorAI) {
            // Hook into inline completions (best-effort via command interception)
            this.disposables.push(vscode.workspace.onDidChangeTextDocument((e) => {
                // Detect AI edits: rapid changes with specific patterns
                for (const change of e.contentChanges) {
                    if (change.text.length > 50) {
                        // Potentially AI-generated content
                        this.queueEvent('ai_completion', {
                            file: e.document.uri.fsPath,
                            language: e.document.languageId,
                            message: 'Inline completion applied',
                            metadata: {
                                text_length: change.text.length,
                                range_offset: change.rangeOffset,
                                range_length: change.rangeLength
                            }
                        });
                    }
                }
            }));
        }
        // Copilot chat / inline (GHCP)
        const copilot = vscode.extensions.getExtension('github.copilot-chat');
        if (copilot) {
            // Listen for Copilot commands
            this.disposables.push(vscode.commands.registerCommand('github.copilot.generate', () => {
                this.queueEvent('ai_chat', { message: 'Copilot generation triggered' });
            }));
        }
    }
    // --- Editor Activity Watchers ---
    attachEditorWatchers() {
        // Cursor position changes (debounced - not sent on every keystroke)
        this.disposables.push(vscode.window.onDidChangeTextEditorSelection((e) => {
            const editor = e.textEditor;
            if (!editor)
                return;
            // Only track significant cursor movements (not every change)
            // This is lightweight position data, sent infrequently
            const doc = editor.document;
            const pos = e.selections[0]?.active;
            if (!pos)
                return;
            this.queueEvent('file_edit', {
                file: doc.uri.fsPath,
                language: doc.languageId,
                cursor_line: pos.line,
                cursor_column: pos.character
            });
        }));
    }
    /** Force flush any pending batched events. */
    dispose() {
        this.stop();
    }
}
exports.RewindRecorder = RewindRecorder;
//# sourceMappingURL=recorder.js.map