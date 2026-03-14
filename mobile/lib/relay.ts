import type { Envelope, Payload } from './protocol';
import { PROTOCOL_VERSION } from './protocol';
import { openPayload, sealPayload, type KeyPair } from './crypto';

const INITIAL_BACKOFF_MS = 1_000;
const MAX_BACKOFF_MS = 30_000;
const BACKOFF_MULTIPLIER = 2;
const PING_INTERVAL_MS = 30_000;

export class RelayClient {
  private ws: WebSocket | null = null;
  private relayURL: string;
  private onEnvelope: (env: Envelope) => void;

  private authToken: string | null = null;
  private machineId: string | undefined;

  private backoffMs = INITIAL_BACKOFF_MS;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private pingTimer: ReturnType<typeof setInterval> | null = null;
  private intentionalDisconnect = false;

  // E2E encryption keys (set per-machine after pairing)
  private ownKeyPair: KeyPair | null = null;
  private peerPublicKeys: Map<string, Uint8Array> = new Map();

  constructor(relayURL: string, onEnvelope: (env: Envelope) => void) {
    this.relayURL = relayURL;
    this.onEnvelope = onEnvelope;
  }

  setOwnKeyPair(kp: KeyPair): void {
    this.ownKeyPair = kp;
  }

  setPeerPublicKey(machineId: string, peerPub: Uint8Array): void {
    this.peerPublicKeys.set(machineId, peerPub);
  }

  connect(authToken: string, machineId?: string): void {
    this.authToken = authToken;
    this.machineId = machineId;
    this.intentionalDisconnect = false;
    this.backoffMs = INITIAL_BACKOFF_MS;
    this._open();
  }

  /**
   * Send a raw envelope (already has payload set).
   */
  send(envelope: Envelope): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error('relay: not connected');
    }
    this.ws.send(JSON.stringify(envelope));
  }

  /**
   * Encrypt a payload and send it as an envelope to a specific machine/session.
   */
  sendEncrypted(
    machineId: string,
    sessionId: string,
    type: Envelope['type'],
    payload: Payload,
  ): void {
    const peerPub = this.peerPublicKeys.get(machineId);
    if (!peerPub || !this.ownKeyPair) {
      throw new Error('relay: missing encryption keys for machine ' + machineId);
    }

    const encryptedPayload = sealPayload(payload, peerPub, this.ownKeyPair.privateKey);

    this.send({
      v: PROTOCOL_VERSION,
      type,
      machine_id: machineId,
      session_id: sessionId,
      timestamp: Date.now(),
      payload: encryptedPayload,
    });
  }

  /**
   * Try to decrypt an envelope's payload using the peer key for its machine.
   * Returns null if decryption is not possible (missing keys or unencrypted message).
   */
  decryptPayload(env: Envelope): Payload | null {
    if (!env.payload || !this.ownKeyPair) return null;
    const peerPub = this.peerPublicKeys.get(env.machine_id);
    if (!peerPub) return null;

    try {
      return openPayload(env.payload, peerPub, this.ownKeyPair.privateKey);
    } catch {
      return null;
    }
  }

  disconnect(): void {
    this.intentionalDisconnect = true;
    this._clearReconnect();
    this._clearPing();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }

  private _open(): void {
    const url = this._buildURL();
    const ws = new WebSocket(url);
    this.ws = ws;

    ws.onmessage = (event) => {
      try {
        const env = JSON.parse(event.data as string) as Envelope;
        if (env.type === 'pong') return; // swallow pong
        this.onEnvelope(env);
      } catch {
        // ignore malformed messages
      }
    };

    ws.onopen = () => {
      this.backoffMs = INITIAL_BACKOFF_MS;
      this._startPing();
    };

    ws.onerror = () => {
      // onclose will fire after onerror; handle reconnect there
    };

    ws.onclose = () => {
      this.ws = null;
      this._clearPing();
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

  private _startPing(): void {
    this._clearPing();
    this.pingTimer = setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({
          v: PROTOCOL_VERSION,
          type: 'ping',
          machine_id: '',
          session_id: '',
          timestamp: Date.now(),
          payload: '',
        }));
      }
    }, PING_INTERVAL_MS);
  }

  private _clearPing(): void {
    if (this.pingTimer !== null) {
      clearInterval(this.pingTimer);
      this.pingTimer = null;
    }
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
