import sodium from 'libsodium-wrappers';
import type { Payload } from './protocol';

export interface KeyPair {
  publicKey: Uint8Array;
  privateKey: Uint8Array;
}

/**
 * Initialize libsodium. Must be awaited before calling any other crypto function.
 */
export async function initCrypto(): Promise<void> {
  await sodium.ready;
}

/**
 * Generate an X25519 key pair for use with NaCl box.
 */
export function generateKeyPair(): KeyPair {
  const kp = sodium.crypto_box_keypair();
  return {
    publicKey: kp.publicKey,
    privateKey: kp.privateKey,
  };
}

/**
 * Encrypt plaintext using NaCl box (X25519 + XSalsa20-Poly1305).
 * Wire format: nonce (24 bytes) || ciphertext
 * Compatible with Go's golang.org/x/crypto/nacl/box.
 */
export function encrypt(
  plaintext: Uint8Array,
  recipientPub: Uint8Array,
  senderPriv: Uint8Array,
): Uint8Array {
  const nonce = sodium.randombytes_buf(sodium.crypto_box_NONCEBYTES); // 24 bytes
  const ciphertext = sodium.crypto_box_easy(plaintext, nonce, recipientPub, senderPriv);
  const result = new Uint8Array(nonce.length + ciphertext.length);
  result.set(nonce, 0);
  result.set(ciphertext, nonce.length);
  return result;
}

/**
 * Decrypt a message produced by encrypt().
 * Expects wire format: nonce (24 bytes) || ciphertext
 */
export function decrypt(
  message: Uint8Array,
  senderPub: Uint8Array,
  recipientPriv: Uint8Array,
): Uint8Array {
  const nonceLen = sodium.crypto_box_NONCEBYTES; // 24
  if (message.length < nonceLen) {
    throw new Error('crypto: message too short');
  }
  const nonce = message.slice(0, nonceLen);
  const ciphertext = message.slice(nonceLen);
  const plaintext = sodium.crypto_box_open_easy(ciphertext, nonce, senderPub, recipientPriv);
  if (!plaintext) {
    throw new Error('crypto: decryption failed');
  }
  return plaintext;
}

/**
 * Seal a Payload object: JSON-encode then encrypt.
 */
export function sealPayload(
  payload: Payload,
  recipientPub: Uint8Array,
  senderPriv: Uint8Array,
): Uint8Array {
  const json = JSON.stringify(payload);
  const plaintext = new TextEncoder().encode(json);
  return encrypt(plaintext, recipientPub, senderPriv);
}

/**
 * Open a sealed Payload: decrypt then JSON-parse.
 */
export function openPayload(
  ciphertext: Uint8Array,
  senderPub: Uint8Array,
  recipientPriv: Uint8Array,
): Payload {
  const plaintext = decrypt(ciphertext, senderPub, recipientPriv);
  const json = new TextDecoder().decode(plaintext);
  return JSON.parse(json) as Payload;
}
