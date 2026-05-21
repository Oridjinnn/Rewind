"use strict";
/**
 * Rewind Status Bar — shows recording indicator and quick controls.
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
exports.RewindStatusBar = void 0;
const vscode = __importStar(require("vscode"));
class RewindStatusBar {
    constructor() {
        this._isRecording = false;
        this._eventCount = 0;
        this.statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
        this.statusBarItem.command = 'rewind.toggleRecording';
        this.statusBarItem.tooltip = 'Rewind IDE Recording — Click to toggle';
        this.updateDisplay();
        this.statusBarItem.show();
    }
    get isRecording() { return this._isRecording; }
    setRecording(on) {
        this._isRecording = on;
        this.updateDisplay();
    }
    setEventCount(count) {
        this._eventCount = count;
        this.updateDisplay();
    }
    updateDisplay() {
        if (this._isRecording) {
            this.statusBarItem.text = `$(circle-filled) Rewind: ${this._eventCount} events`;
            this.statusBarItem.backgroundColor = undefined;
        }
        else {
            this.statusBarItem.text = `$(circle-outline) Rewind: OFF`;
            this.statusBarItem.backgroundColor = new vscode.ThemeColor('statusBarItem.warningBackground');
        }
    }
    dispose() {
        this.statusBarItem.dispose();
    }
}
exports.RewindStatusBar = RewindStatusBar;
//# sourceMappingURL=statusbar.js.map