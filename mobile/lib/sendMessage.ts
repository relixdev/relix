import { sealPayload } from './crypto';
import type { KeyPair } from './crypto';
import type { Envelope, Payload } from './protocol';
import { PROTOCOL_VERSION } from './protocol';
import { RelayClient } from './relay';

/**
 * Build a user_message Payload, encrypt it, wrap in an Envelope, and send via relay.
 */
export async function sendMessage(
  text: string,
  sessionId: string,
  machineId: string,
  keys: KeyPair,
  peerPub: Uint8Array,
  relay: RelayClient,
): Promise<Payload> {
  const payload: Payload = {
    kind: 'user_message',
    seq: Date.now(),
    data: {
      text,
      timestamp: Date.now(),
    },
  };

  // sealPayload returns a base64-encoded string ready for wire transport
  const base64 = sealPayload(payload, peerPub, keys.privateKey);

  const envelope: Envelope = {
    v: PROTOCOL_VERSION,
    type: 'user_input',
    machine_id: machineId,
    session_id: sessionId,
    timestamp: Date.now(),
    payload: base64,
  };

  relay.send(envelope);

  return payload;
}
