import { initCrypto, generateKeyPair, encrypt, decrypt, sealPayload, openPayload } from '../crypto';
import type { Payload } from '../protocol';

beforeAll(async () => {
  await initCrypto();
});

describe('generateKeyPair', () => {
  it('returns 32-byte public and private keys', () => {
    const kp = generateKeyPair();
    expect(kp.publicKey).toBeInstanceOf(Uint8Array);
    expect(kp.privateKey).toBeInstanceOf(Uint8Array);
    expect(kp.publicKey.length).toBe(32);
    expect(kp.privateKey.length).toBe(32);
  });
});

describe('encrypt / decrypt', () => {
  it('round-trips plaintext', () => {
    const alice = generateKeyPair();
    const bob = generateKeyPair();
    const plaintext = new TextEncoder().encode('hello relix');

    const ciphertext = encrypt(plaintext, bob.publicKey, alice.privateKey);
    const recovered = decrypt(ciphertext, alice.publicKey, bob.privateKey);

    expect(recovered).toEqual(plaintext);
  });

  it('throws on wrong key', () => {
    const alice = generateKeyPair();
    const bob = generateKeyPair();
    const eve = generateKeyPair();
    const plaintext = new TextEncoder().encode('secret');

    const ciphertext = encrypt(plaintext, bob.publicKey, alice.privateKey);

    expect(() => decrypt(ciphertext, alice.publicKey, eve.privateKey)).toThrow();
  });

  it('produces different ciphertext on each call (random nonce)', () => {
    const alice = generateKeyPair();
    const bob = generateKeyPair();
    const plaintext = new TextEncoder().encode('same message');

    const ct1 = encrypt(plaintext, bob.publicKey, alice.privateKey);
    const ct2 = encrypt(plaintext, bob.publicKey, alice.privateKey);

    // Nonces (first 24 bytes) should differ
    expect(ct1.slice(0, 24)).not.toEqual(ct2.slice(0, 24));
    // Full ciphertexts should differ
    expect(ct1).not.toEqual(ct2);
  });
});

describe('sealPayload / openPayload', () => {
  it('round-trips a Payload object', () => {
    const alice = generateKeyPair();
    const bob = generateKeyPair();

    const payload: Payload = {
      kind: 'assistant_message',
      seq: 42,
      data: { text: 'hello from agent', tool: 'claude' },
    };

    const sealed = sealPayload(payload, bob.publicKey, alice.privateKey);
    const opened = openPayload(sealed, alice.publicKey, bob.privateKey);

    expect(opened.kind).toBe(payload.kind);
    expect(opened.seq).toBe(payload.seq);
    expect(opened.data).toEqual(payload.data);
  });
});
