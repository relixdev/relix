/**
 * SAS (Short Authentication String) emoji computation.
 *
 * Algorithm matches Go relixctl:
 *   SHA-256(sorted(ownPubHex + peerPubHex))
 *   Map first 4 bytes to EMOJI_PALETTE.
 *
 * "Sorted" means the lexicographically smaller hex string comes first.
 */

// 256-entry emoji palette — one emoji per byte value.
export const EMOJI_PALETTE: string[] = [
  '🐶','🐱','🐭','🐹','🐰','🦊','🐻','🐼','🐨','🐯','🦁','🐮','🐷','🐸','🐵','🙈',
  '🙉','🙊','🐔','🐧','🐦','🐤','🦆','🦅','🦉','🦇','🐺','🐗','🐴','🦄','🐝','🐛',
  '🦋','🐌','🐞','🐜','🦟','🦗','🕷','🦂','🐢','🐍','🦎','🦖','🦕','🐙','🦑','🦐',
  '🦞','🦀','🐡','🐠','🐟','🐬','🐳','🐋','🦈','🐊','🐅','🐆','🦓','🦍','🦧','🦣',
  '🐘','🦛','🦏','🐪','🐫','🦒','🦘','🦬','🐃','🐂','🐄','🐎','🐖','🐏','🐑','🦙',
  '🐐','🦌','🐕','🐩','🦮','🐕‍🦺','🐈','🐈‍⬛','🪶','🐓','🦃','🦤','🦚','🦜','🦢','🦩',
  '🕊','🐇','🦝','🦨','🦡','🦫','🦦','🦥','🐁','🐀','🐿','🦔','🐾','🐉','🐲','🌵',
  '🎄','🌲','🌳','🌴','🌱','🌿','☘','🍀','🎍','🎋','🍃','🍂','🍁','🍄','🌾','💐',
  '🌷','🌹','🥀','🌺','🌸','🌼','🌻','🌞','🌝','🌛','🌜','🌚','🌕','🌖','🌗','🌘',
  '🌑','🌒','🌓','🌔','🌙','🌟','⭐','🌠','🌌','☁','⛅','🌤','🌈','🌂','☂','☔',
  '⛱','⚡','❄','☃','⛄','🌬','💨','💧','💦','🌊','🌀','🌈','🔥','🌪','🌫','🌊',
  '🍏','🍎','🍐','🍊','🍋','🍌','🍉','🍇','🍓','🫐','🍈','🍒','🍑','🥭','🍍','🥥',
  '🥝','🍅','🍆','🥑','🥦','🥬','🥒','🌶','🫑','🧄','🧅','🥔','🍠','🥐','🥯','🍞',
  '🥖','🥨','🧀','🥚','🍳','🧈','🥞','🧇','🥓','🥩','🍗','🍖','🌭','🍔','🍟','🍕',
  '🫓','🥙','🧆','🌮','🌯','🫔','🥗','🥘','🫕','🥫','🍝','🍜','🍲','🍛','🍣','🍱',
  '🥟','🦪','🍤','🍙','🍚','🍘','🍥','🥮','🍢','🧁','🍰','🎂','🍮','🍭','🍬','🍫',
];

/**
 * Compute 4 SAS emojis from two hex-encoded public keys.
 * Returns a Promise so it works in both browser and Node environments.
 */
export async function computeSASEmojis(
  ownPubHex: string,
  peerPubHex: string,
): Promise<string[]> {
  // Sort lexicographically so both sides produce the same hash regardless of order
  const a = ownPubHex < peerPubHex ? ownPubHex : peerPubHex;
  const b = ownPubHex < peerPubHex ? peerPubHex : ownPubHex;
  const combined = a + b;

  const encoder = new TextEncoder();
  const data = encoder.encode(combined);
  const hashBuffer = await globalThis.crypto.subtle.digest('SHA-256', data);
  const hashBytes = new Uint8Array(hashBuffer);

  return [
    EMOJI_PALETTE[hashBytes[0]],
    EMOJI_PALETTE[hashBytes[1]],
    EMOJI_PALETTE[hashBytes[2]],
    EMOJI_PALETTE[hashBytes[3]],
  ];
}
