// Tests for SAS emoji computation.
// Uses Node's built-in crypto to polyfill globalThis.crypto.subtle in the test env.

import { createHash } from 'crypto';

// Polyfill globalThis.crypto.subtle for Node test environment
const nodeCrypto = {
  subtle: {
    digest: async (_algo: string, data: Uint8Array): Promise<ArrayBuffer> => {
      const hash = createHash('sha256').update(Buffer.from(data)).digest();
      return hash.buffer.slice(hash.byteOffset, hash.byteOffset + hash.byteLength) as ArrayBuffer;
    },
  },
};

(globalThis as any).crypto = nodeCrypto;

import { computeSASEmojis, EMOJI_PALETTE } from '../sas';

describe('computeSASEmojis', () => {
  it('returns an array of 4 strings', async () => {
    const emojis = await computeSASEmojis('aabbcc', 'ddeeff');
    expect(emojis).toHaveLength(4);
    emojis.forEach((e) => expect(typeof e).toBe('string'));
  });

  it('is symmetric — same result regardless of argument order', async () => {
    const a = 'aabbccdd';
    const b = '11223344';
    const result1 = await computeSASEmojis(a, b);
    const result2 = await computeSASEmojis(b, a);
    expect(result1).toEqual(result2);
  });

  it('produces different emojis for different key pairs', async () => {
    const result1 = await computeSASEmojis('aaaa', 'bbbb');
    const result2 = await computeSASEmojis('cccc', 'dddd');
    expect(result1).not.toEqual(result2);
  });

  it('is deterministic — same inputs give same output', async () => {
    const r1 = await computeSASEmojis('deadbeef', 'cafebabe');
    const r2 = await computeSASEmojis('deadbeef', 'cafebabe');
    expect(r1).toEqual(r2);
  });

  it('all returned emojis are in the palette', async () => {
    const emojis = await computeSASEmojis('123456', 'abcdef');
    emojis.forEach((e) => expect(EMOJI_PALETTE).toContain(e));
  });
});
