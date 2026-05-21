/**
 * Rewind IDE Protocol Client
 * Sends JSON events to the Rewind IDE recording server (localhost:9876).
 */

import * as http from 'http';

const PROTOCOL_VERSION = 'rewind-ide-v1';

export interface IDEEventData {
    type: string;
    timestamp: string;
    file?: string;
    language?: string;
    lines_added?: number;
    lines_removed?: number;
    cursor_line?: number;
    cursor_column?: number;
    duration_ms?: number;
    exit_code?: number;
    message?: string;
    tags?: string[];
    metadata?: Record<string, any>;
    content_snapshot?: string;
    session_id?: string;
}

export interface IDEEvent {
    protocol: string;
    ide: string;
    version: string;
    project: string;
    project_path: string;
    event: IDEEventData;
}

export class RewindProtocol {
    private serverUrl: string;
    private ideName: string;
    private ideVersion: string;

    constructor(port: number = 9876, ideName: string = 'vscode', ideVersion: string = '1.0.0') {
        this.serverUrl = `http://localhost:${port}`;
        this.ideName = ideName;
        this.ideVersion = ideVersion;
    }

    /**
     * Send a single event to the Rewind server.
     */
    sendEvent(
        eventType: string,
        extra: Partial<IDEEventData>,
        projectName: string,
        projectPath: string
    ): Promise<void> {
        const event: IDEEvent = {
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
    sendBatch(events: IDEEvent[]): Promise<{ accepted: number; rejected: number; errors: string[] }> {
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
                    } catch {
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
    async checkHealth(): Promise<boolean> {
        try {
            const resp = await this.getJson('/health');
            return resp?.status === 'healthy';
        } catch {
            return false;
        }
    }

    /**
     * Get recording status.
     */
    async getStatus(): Promise<any> {
        return this.getJson('/status');
    }

    private postJson(path: string, data: any): Promise<void> {
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

    private getJson(path: string): Promise<any> {
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
                    } catch {
                        resolve(null);
                    }
                });
            }).on('error', reject);
        });
    }
}