import sodium from 'libsodium-wrappers';
import * as SecureStore from 'expo-secure-store';
import type { Payload } from './protocol';

export interface KeyPair {
  publicKey: Uint8Array;
  privateKey: Uint8Array;
}

const OWN_KEYPAIR_KEY = 'relix_own_keypair';

// ─── Initialization ────────────────────────────────────────────────────

/**
 * Initialize libsodium. Must be awaited before calling any other crypto function.
 */
export async function initCrypto(): Promise<void> {
  await sodium.ready;
}

// ─── Key generation ────────────────────────────────────────────────────

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

// ─── Key persistence ───────────────────────────────────────────────────

/**
 * Get or create the mobile app's persistent X25519 key pair.
 * Stored securely via expo-secure-store.
 */
export async function getOrCreateKeyPair(): Promise<KeyPair> {
  await initCrypto();

  const stored = await SecureStore.getItemAsync(OWN_KEYPAIR_KEY);
  if (stored) {
    try {
      const parsed = JSON.parse(stored) as { publicKey: number[]; privateKey: number[] };
      return {
        publicKey: new Uint8Array(parsed.publicKey),
        privateKey: new Uint8Array(parsed.privateKey),
      };
    } catch {
      // Corrupted — regenerate
    }
  }

  const kp = generateKeyPair();
  await SecureStore.setItemAsync(
    OWN_KEYPAIR_KEY,
    JSON.stringify({
      publicKey: Array.from(kp.publicKey),
      privateKey: Array.from(kp.privateKey),
    }),
  );
  return kp;
}

// ─── Encoding helpers ──────────────────────────────────────────────────

export function toBase64(data: Uint8Array): string {
  return sodium.to_base64(data, sodium.base64_variants.URLSAFE_NO_PADDING);
}

export function fromBase64(encoded: string): Uint8Array {
  return sodium.from_base64(encoded, sodium.base64_variants.URLSAFE_NO_PADDING);
}

export function toHex(data: Uint8Array): string {
  return sodium.to_hex(data);
}

export function fromHex(hex: string): Uint8Array {
  return sodium.from_hex(hex);
}

// ─── Encrypt / Decrypt ─────────────────────────────────────────────────

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

// ─── Payload seal / open ───────────────────────────────────────────────

/**
 * Seal a Payload object: JSON-encode then encrypt.
 */
export function sealPayload(
  payload: Payload,
  recipientPub: Uint8Array,
  senderPriv: Uint8Array,
): string {
  const json = JSON.stringify(payload);
  const plaintext = new TextEncoder().encode(json);
  const sealed = encrypt(plaintext, recipientPub, senderPriv);
  return toBase64(sealed);
}

/**
 * Open a sealed Payload: base64-decode, decrypt, then JSON-parse.
 */
export function openPayload(
  encodedCiphertext: string,
  senderPub: Uint8Array,
  recipientPriv: Uint8Array,
): Payload {
  const ciphertext = fromBase64(encodedCiphertext);
  const plaintext = decrypt(ciphertext, senderPub, recipientPriv);
  const json = new TextDecoder().decode(plaintext);
  return JSON.parse(json) as Payload;
}
