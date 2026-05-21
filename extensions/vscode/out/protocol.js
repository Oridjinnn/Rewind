"use strict";
/**
 * Rewind IDE Protocol Client
 * Sends JSON events to the Rewind IDE recording server (localhost:9876).
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
exports.RewindProtocol = void 0;
const http = __importStar(require("http"));
const PROTOCOL_VERSION = 'rewind-ide-v1';
class RewindProtocol {
    constructor(port = 9876, ideName = 'vscode', ideVersion = '1.0.0') {
        this.serverUrl = `http://localhost:${port}`;
        this.ideName = ideName;
        this.ideVersion = ideVersion;
    }
    /**
     * Send a single event to the Rewind server.
     */
    sendEvent(eventType, extra, projectName, projectPath) {
        const event = {
            protocol: PROTOCOL_VERSION,
            ide: this.ideName,
            version: this.ideVersion,
            project: projectName,
            project_path: projectPath,
            event: {
                type: eventType,
                timestamp: new Date().toISOString(),
                ...extra
            }
        };
        return this.postJson('/', event);
    }
    /**
     * Send multiple events in batch.
     */
    sendBatch(events) {
        const body = JSON.stringify(events);
        return new Promise((resolve, reject) => {
            const url = new URL('/batch', this.serverUrl);
            const req = http.request({
                hostname: url.hostname,
                port: url.port,
                path: url.pathname,
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Content-Length': Buffer.byteLength(body)
                },
                timeout: 5000
            }, (res) => {
                let data = '';
                res.on('data', chunk => data += chunk);
                res.on('end', () => {
                    try {
                        resolve(JSON.parse(data));
                    }
                    catch {
                        resolve({ accepted: 0, rejected: events.length, errors: ['parse error'] });
                    }
                });
            });
            req.on('error', (err) => {
                // Server may not be running — silently ignore
                console.debug('[Rewind] Server unreachable:', err.message);
                resolve({ accepted: 0, rejected: events.length, errors: [err.message] });
            });
            req.on('timeout', () => {
                req.destroy();
                resolve({ accepted: 0, rejected: events.length, errors: ['timeout'] });
            });
            req.write(body);
            req.end();
        });
    }
    /**
     * Check server health.
     */
    async checkHealth() {
        try {
            const resp = await this.getJson('/health');
            return resp?.status === 'healthy';
        }
        catch {
            return false;
        }
    }
    /**
     * Get recording status.
     */
    async getStatus() {
        return this.getJson('/status');
    }
    postJson(path, data) {
        const body = JSON.stringify(data);
        return new Promise((resolve, _reject) => {
            const url = new URL(path, this.serverUrl);
            const req = http.request({
                hostname: url.hostname,
                port: url.port,
                path: url.pathname,
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Content-Length': Buffer.byteLength(body)
                },
                timeout: 3000
            }, (res) => {
                res.resume();
                resolve(); // fire and forget
            });
            req.on('error', () => resolve()); // silent
            req.on('timeout', () => { req.destroy(); resolve(); });
            req.write(body);
            req.end();
        });
    }
    getJson(path) {
        return new Promise((resolve, reject) => {
            const url = new URL(path, this.serverUrl);
            http.get({
                hostname: url.hostname,
                port: url.port,
                path: url.pathname,
                timeout: 3000
            }, (res) => {
                let data = '';
                res.on('data', chunk => data += chunk);
                res.on('end', () => {
                    try {
                        resolve(JSON.parse(data));
                    }
                    catch {
                        resolve(null);
                    }
                });
            }).on('error', reject);
        });
    }
}
exports.RewindProtocol = RewindProtocol;
//# sourceMappingURL=protocol.js.map