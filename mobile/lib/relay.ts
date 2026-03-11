import type { Envelope } from './protocol';

const INITIAL_BACKOFF_MS = 1_000;
const MAX_BACKOFF_MS = 30_000;
const BACKOFF_MULTIPLIER = 2;

export class RelayClient {
  private ws: WebSocket | null = null;
  private relayURL: string;
  private onEnvelope: (env: Envelope) => void;

  private authToken: string | null = null;
  private machineId: string | undefined;

  private backoffMs = INITIAL_BACKOFF_MS;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private intentionalDisconnect = false;

  constructor(relayURL: string, onEnvelope: (env: Envelope) => void) {
    this.relayURL = relayURL;
    this.onEnvelope = onEnvelope;
  }

  connect(authToken: string, machineId?: string): void {
    this.authToken = authToken;
    this.machineId = machineId;
    this.intentionalDisconnect = false;
    this.backoffMs = INITIAL_BACKOFF_MS;
    this._open();
  }

  send(envelope: Envelope): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error('relay: not connected');
    }
    this.ws.send(JSON.stringify(envelope));
  }

  disconnect(): void {
    this.intentionalDisconnect = true;
    this._clearReconnect();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  private _open(): void {
    const url = this._buildURL();
    const ws = new WebSocket(url);
    this.ws = ws;

    ws.onmessage = (event) => {
      try {
        const env = JSON.parse(event.data as string) as Envelope;
        this.onEnvelope(env);
      } catch {
        // ignore malformed messages
      }
    };

    ws.onopen = () => {
      this.backoffMs = INITIAL_BACKOFF_MS;
    };

    ws.onerror = () => {
      // onclose will fire after onerror; handle reconnect there
    };

    ws.onclose = () => {
      this.ws = null;
      if (!this.intentionalDisconnect) {
        this._scheduleReconnect();
      }
    };
  }

  private _buildURL(): string {
    const params = new URLSearchParams({ token: this.authToken ?? '' });
    if (this.machineId) {
      params.set('machine_id', this.machineId);
    }
    return `${this.relayURL}?${params.toString()}`;
  }

  private _scheduleReconnect(): void {
    this._clearReconnect();
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      if (!this.intentionalDisconnect) {
        this._open();
      }
    }, this.backoffMs);
    this.backoffMs = Math.min(this.backoffMs * BACKOFF_MULTIPLIER, MAX_BACKOFF_MS);
  }

  private _clearReconnect(): void {
    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }
}
