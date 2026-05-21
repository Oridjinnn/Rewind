/**
 * Rewind Status Bar — shows recording indicator and quick controls.
 */

import * as vscode from 'vscode';

export class RewindStatusBar {
    private statusBarItem: vscode.StatusBarItem;
    private _isRecording: boolean = false;
    private _eventCount: number = 0;

    constructor() {
        this.statusBarItem = vscode.window.createStatusBarItem(
            vscode.StatusBarAlignment.Right,
            100
        );
        this.statusBarItem.command = 'rewind.toggleRecording';
        this.statusBarItem.tooltip = 'Rewind IDE Recording — Click to toggle';
        this.updateDisplay();
        this.statusBarItem.show();
    }

    get isRecording(): boolean { return this._isRecording; }

    setRecording(on: boolean) {
        this._isRecording = on;
        this.updateDisplay();
    }

    setEventCount(count: number) {
        this._eventCount = count;
        this.updateDisplay();
    }

    private updateDisplay() {
        if (this._isRecording) {
            this.statusBarItem.text = `$(circle-filled) Rewind: ${this._eventCount} events`;
            this.statusBarItem.backgroundColor = undefined;
        } else {
            this.statusBarItem.text = `$(circle-outline) Rewind: OFF`;
            this.statusBarItem.backgroundColor = new vscode.ThemeColor(
                'statusBarItem.warningBackground'
            );
        }
    }

    dispose() {
        this.statusBarItem.dispose();
    }
}