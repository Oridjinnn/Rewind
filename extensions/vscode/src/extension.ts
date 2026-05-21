/**
 * Rewind IDE Extension — Main entry point for VS Code / Cursor.
 */

import * as vscode from 'vscode';
import { RewindProtocol } from './protocol';
import { RewindRecorder } from './recorder';
import { RewindStatusBar } from './statusbar';
import { spawn, ChildProcess } from 'child_process';

const MAX_SERVER_START_RETRIES = 10; // Max attempts to check health after spawning
const SERVER_START_RETRY_DELAY_MS = 1000; // Delay between health checks

let recorder: RewindRecorder;
let statusBar: RewindStatusBar;
let protocol: RewindProtocol;
let serverProcess: ChildProcess | null = null;

export function activate(context: vscode.ExtensionContext) {
    const config = vscode.workspace.getConfiguration('rewind');
    const port = config.get('serverPort', 9876);
    const ideName = vscode.env.appName.toLowerCase().includes('cursor') ? 'cursor' : 'vscode';

    protocol = new RewindProtocol(port, ideName, vscode.version);
    recorder = new RewindRecorder(protocol);
    statusBar = new RewindStatusBar();

    // Check if recording should start
    const autoStart = config.get('serverEnabled', true);
    if (autoStart) {
        ensureServerRunning().then(success => {
            if (success) {
                recorder.start();
                statusBar.setRecording(true);
            } else {
                statusBar.setRecording(false); // Server failed to start, don't show recording
            }
        });
    }

    // Register commands
    context.subscriptions.push(
        vscode.commands.registerCommand('rewind.toggleRecording', handleToggleRecording),
        vscode.commands.registerCommand('rewind.showStatus', handleShowStatus),
        vscode.commands.registerCommand('rewind.showActivity', handleShowActivity),
        vscode.commands.registerCommand('rewind.showProjects', handleShowProjects),
        vscode.commands.registerCommand('rewind.analyzeProject', handleAnalyzeProject),
        vscode.commands.registerCommand('rewind.searchHistory', handleSearchHistory),
        vscode.commands.registerCommand('rewind.importShellHistory', handleImportHistory),
        vscode.commands.registerCommand('rewind.openPermissions', handleOpenPermissions),
        statusBar,
        recorder
    );

    // Listen for config changes
    context.subscriptions.push(
        vscode.workspace.onDidChangeConfiguration((e) => {
            if (e.affectsConfiguration('rewind')) {
                recorder.reloadConfig();
            }
        })
    );

    // Periodically check server health
    const healthInterval = setInterval(async () => {
        const healthy = await protocol.checkHealth();
        
        if (!healthy) {
            // Jangan langsung mematikan recording, beri tahu user bahwa server tidak merespon
            // Kita tetap biarkan status 'recording' jika user memang ingin merekam, 
            // tapi mungkin status bar bisa menampilkan icon 'error/warning' (perlu update statusbar.ts)
            console.warn('Rewind server is unreachable. Make sure to run "rewind ide start"');
        }
        
        // Biarkan status bar tetap sinkron dengan keadaan 'recorder' yang sebenarnya
    }, 30000); // every 30 seconds

    context.subscriptions.push({
        dispose: () => clearInterval(healthInterval)
    });

    vscode.window.showInformationMessage(
        `Rewind IDE active — recording ${ideName} activity to local SQLite database`
    );
}

export function deactivate() {
    if (recorder) {
        recorder.stop();
    }
    if (serverProcess) {
        serverProcess.kill();
    }
    if (statusBar) {
        statusBar.dispose();
    }
}

// --- Command Handlers ---

async function ensureServerRunning(): Promise<boolean> {
    const healthy = await protocol.checkHealth();
    if (healthy) {
        console.log('[Rewind] Backend server already running and healthy.');
        return true;
    }

    let serverSpawnedSuccessfully = false;

    await vscode.window.withProgress({
        location: vscode.ProgressLocation.Notification,
        title: "Starting Rewind Backend Server",
        cancellable: false
    }, async (progress) => {
        progress.report({ message: "Launching Rewind backend process..." });

        if (!serverProcess) { // Hanya spawn jika belum pernah dicoba atau gagal sebelumnya
            serverProcess = spawn('rewind', ['ide', 'start'], {
                shell: true,
                stdio: 'ignore', // Tetap ini untuk menghindari output yang mengotori VS Code
                detached: true
            });

            serverProcess.on('error', (err) => {
                vscode.window.showErrorMessage(`Failed to start Rewind server: ${err.message}. Pastikan "rewind" CLI ada di PATH Anda.`);
                serverProcess = null; // Izinkan spawn ulang pada percobaan berikutnya
                serverSpawnedSuccessfully = false; // Tandai spawn sebagai gagal
            });

            serverProcess.on('exit', (code, signal) => {
                if (code !== 0 && code !== null) { // Kode keluar bukan nol berarti kegagalan
                    console.error(`[Rewind] Backend server process exited with code ${code}`);
                    // Hanya tampilkan error jika bukan karena sinyal kill dari deactivate
                    if (signal !== 'SIGTERM' && signal !== 'SIGKILL') {
                        vscode.window.showErrorMessage(`Rewind backend server process exited unexpectedly with code ${code}.`);
                    }
                }
                serverProcess = null; // Proses sudah tidak ada
            });

            // Beri waktu sebentar agar proses benar-benar dimulai
            await new Promise(resolve => setTimeout(resolve, 500));
            serverSpawnedSuccessfully = (serverProcess !== null); // Asumsikan berhasil jika objek proses ada
        } else {
            // Jika serverProcess sudah ada, berarti spawn sebelumnya sudah dicoba
            // dan kita hanya menunggu agar menjadi sehat.
            serverSpawnedSuccessfully = true;
        }

        if (!serverSpawnedSuccessfully) {
            // Jika spawn itu sendiri gagal (misalnya, 'rewind' tidak ditemukan), tidak ada gunanya melakukan health check
            return; // Keluar dari callback progress
        }

        let isServerHealthy = false;
        for (let i = 0; i < MAX_SERVER_START_RETRIES; i++) {
            progress.report({ increment: 100 / MAX_SERVER_START_RETRIES, message: `Checking server health (${i + 1}/${MAX_SERVER_START_RETRIES})...` });
            isServerHealthy = await protocol.checkHealth();
            if (isServerHealthy) {
                console.log('[Rewind] Backend server is healthy.');
                break; // Server sehat, keluar dari loop
            }
            await new Promise(resolve => setTimeout(resolve, SERVER_START_RETRY_DELAY_MS));
        }

        if (!isServerHealthy) {
            vscode.window.showErrorMessage(
                'Rewind backend server failed to become healthy after multiple attempts. Please check your "rewind" CLI installation and ensure no other process is using port 9876.'
            );
            console.error('[Rewind] Backend server failed to start or respond.');
        }
    });
    return await protocol.checkHealth(); // Kembalikan status kesehatan akhir
}

async function handleToggleRecording() {
    const isRecording = statusBar.isRecording;
    if (isRecording) {
        recorder.setEnabled(false);
        statusBar.setRecording(false);
        vscode.window.showInformationMessage('Rewind recording paused.');
    } else {
        const serverReady = await ensureServerRunning();
        if (serverReady) {
            recorder.setEnabled(true);
            statusBar.setRecording(true);
            vscode.window.showInformationMessage('Rewind recording resumed.');
        } else {
            vscode.window.showErrorMessage('Rewind server is not running. Recording cannot be started.');
        }
    }
}

async function handleShowStatus() {
    try {
        const status = await protocol.getStatus();
        const msg = [
            `Rewind Status:`,
            `Server: ${status.server_running ? 'Running' : 'Stopped'}`,
            `Port: ${status.server_port}`,
            `Connected IDEs: ${(status.connected_ides || []).join(', ') || 'none'}`,
            `Activity count: ${status.activity_count}`,
            `Active project: ${status.active_project || 'none'}`
        ].join('\n');
        vscode.window.showInformationMessage(msg, { modal: false });
    } catch {
        vscode.window.showWarningMessage(
            'Rewind server not running. Start with: rewind ide start'
        );
    }
}

async function handleShowActivity() {
    const term = vscode.window.createTerminal('Rewind Activity');
    term.show();
    term.sendText('rewind ide activity vscode 20');
}

async function handleShowProjects() {
    const term = vscode.window.createTerminal('Rewind Projects');
    term.show();
    term.sendText('rewind ide projects');
}

async function handleAnalyzeProject() {
    const folders = vscode.workspace.workspaceFolders;
    if (!folders || folders.length === 0) {
        vscode.window.showWarningMessage('No project open.');
        return;
    }
    const projectPath = folders[0].uri.fsPath;
    const term = vscode.window.createTerminal('Rewind Analysis');
    term.show();
    term.sendText(`rewind ide analyze "${projectPath}"`);
}

async function handleSearchHistory() {
    const query = await vscode.window.showInputBox({
        prompt: 'Search shell history',
        placeHolder: 'e.g., git push'
    });
    if (!query) return;

    const term = vscode.window.createTerminal('Rewind Search');
    term.show();
    term.sendText(`rewind search "${query}"`);
}

async function handleImportHistory() {
    const term = vscode.window.createTerminal('Rewind Import');
    term.show();
    term.sendText('rewind import-history');
}

async function handleOpenPermissions() {
    // Show permission quick pick
    const options = [
        { label: '$(check) Enable All Recording', description: 'Record files, terminal, git, AI' },
        { label: '$(circle-slash) Disable All Recording', description: 'Stop all recording' },
        { label: '$(file) Toggle File Recording', description: 'On/Off file edits' },
        { label: '$(terminal) Toggle Terminal Recording', description: 'On/Off terminal commands' },
        { label: '$(hubot) Toggle AI Recording', description: 'On/Off Copilot/Cursor interactions' }
    ];

    const choice = await vscode.window.showQuickPick(options, {
        placeHolder: 'Rewind Recording Permissions'
    });

    if (!choice) return;

    const term = vscode.window.createTerminal('Rewind Permissions');
    term.show();

    switch (choice.label) {
        case '$(check) Enable All Recording':
            term.sendText('rewind ide permissions vscode on');
            break;
        case '$(circle-slash) Disable All Recording':
            term.sendText('rewind ide permissions vscode off');
            break;
        case '$(file) Toggle File Recording':
            vscode.window.showInformationMessage('Use rewind ide permissions for granular control');
            break;
        case '$(terminal) Toggle Terminal Recording':
            vscode.window.showInformationMessage('Use rewind ide permissions for granular control');
            break;
        case '$(hubot) Toggle AI Recording':
            vscode.window.showInformationMessage('Use rewind ide permissions for granular control');
            break;
    }
}