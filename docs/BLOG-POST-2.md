# E2E Encrypted Mobile Control for AI Agents

**Published:** 2026-03-14  
**Author:** Zach  
**Reading time:** 6 minutes

---

When you're running AI coding agents, you're dealing with your entire codebase. Proprietary algorithms, customer data, unreleased features. The last thing you want is to route that through a third-party server.

That's why Relix uses end-to-end encryption. Your code never leaves your control.

## The Threat Model

What are we protecting against?

1. **Relay server compromise** — If someone hacks the relay, they should only see encrypted blobs
2. **Network interception** — MITM attacks on the WebSocket connection
3. **Mobile device compromise** — If your phone is compromised, the attacker still can't decrypt old sessions
4. **Relay operator** — We literally cannot read your data even if we wanted to

## The Cryptography

Relix uses a standard, battle-tested crypto stack:

**Key Exchange:** X25519 (Curve25519)
- Each device generates an X25519 keypair on first run
- Public keys are exchanged during pairing
- Shared secret derived using ECDH

**Encryption:** NaCl box (secretbox)
- Symmetric encryption using the shared secret
- 256-bit keys
- Nonce-based (unique per message)
- Authenticated encryption (AEAD)

**Signing:** Ed25519
- Message authentication
- Prevents tampering

## The Flow

### 1. Key Generation

When you install relixctl:

```bash
$ relixctl init
Generating X25519 keypair...
Public key: 8f3a...2b1c
Private key: stored in ~/.relixctl/keys/
```

The private key is encrypted on disk using your system keychain.

### 2. Pairing

When you pair your phone:

```
relixctl pair
Enter this code in the Relix app: 847291
```

The 6-digit code is a short-lived token (5-minute TTL). It's used to exchange public keys:

```
Mobile → Relay: "I want to pair with code 847291"
Relay → CLI: "Mobile wants to pair, here's their public key"
CLI → Relay: "Here's my public key"
Relay → Mobile: "CLI's public key"
```

Both devices now have each other's public keys. The relay never sees the shared secret.

### 3. Message Encryption

Every message is encrypted before it leaves the device:

```go
// CLI encrypts message
sharedSecret := x25519(myPrivateKey, mobilePublicKey)
nonce := randomNonce()
ciphertext := naclBoxEncrypt(plaintext, nonce, sharedSecret)

// Send to relay
relay.Send(ciphertext)
```

The relay sees only ciphertext. It routes the message to the mobile app.

```go
// Mobile decrypts message
sharedSecret := x25519(myPrivateKey, cliPublicKey)
plaintext := naclBoxDecrypt(ciphertext, nonce, sharedSecret)
```

### 4. Session Keys

For long-running sessions, we derive session keys from the master shared secret:

```
sessionKey := HKDF(masterSecret, sessionId, "relix-session")
```

This provides forward secrecy — if the master key is compromised, old sessions are still safe.

## Implementation Details

### Key Storage

**CLI (macOS):**
- Private key stored in `~/.relixctl/keys/`
- Encrypted using macOS Keychain
- Access controlled by file permissions (0600)

**CLI (Linux):**
- Private key stored in `~/.relixctl/keys/`
- Encrypted using libsecret (GNOME Keyring) or KWallet
- Fallback: encrypted with user password

**Mobile (iOS):**
- Private key stored in Keychain
- Protected with `kSecAttrAccessibleWhenUnlockedThisDeviceOnly`

**Mobile (Android):**
- Private key stored in Android Keystore
- Protected with `KEY_PROTECTION_TYPE_DEVICE_CREDENTIAL`

### Message Format

```json
{
  "id": "uuid",
  "timestamp": "2026-03-14T12:00:00Z",
  "from": "cli-uuid",
  "to": "mobile-uuid",
  "type": "message",
  "payload": "base64-encoded-ciphertext",
  "nonce": "base64-encoded-nonce"
}
```

The envelope is plaintext (for routing). The payload is encrypted.

### Performance

- Key generation: ~5ms
- Encryption: ~1ms per KB
- Decryption: ~1ms per KB
- Latency: <50ms round-trip (relay is stateless)

## Why Not Just Use TLS?

TLS protects data in transit between your device and the server. But the server can still read it.

E2E encryption means the server never sees the plaintext. Only the endpoints (CLI and mobile) can decrypt.

Think of it like Signal vs. regular SMS. TLS is like sending a letter in a sealed envelope. E2E encryption is like writing the letter in a language only you and the recipient understand.

## Audit Trail

Every encrypted message is logged with:
- Message ID
- Timestamp
- Sender/recipient IDs
- Size (not content)

These logs are stored for debugging but the content is never logged.

## Self-Hosting

The relay server is open source (MIT). You can run it on your own infrastructure:

```bash
docker run -d \
  -p 8080:8080 \
  -e RELAY_JWT_SECRET=your-secret \
  relixdev/relay:latest
```

Point your CLI and mobile app to your relay URL. Your data never touches our servers.

## The Bottom Line

Your code is yours. Relix just provides the connection between your AI agents and your phone — without ever seeing what's being transmitted.

That's the promise of E2E encryption. And that's how Relix works.

---

**Learn more:**
- [Protocol spec](https://github.com/relixdev/relix/blob/main/protocol/README.md)
- [Crypto implementation](https://github.com/relixdev/relix/blob/main/protocol/crypto.go)
- [Self-hosting guide](https://relix.sh/docs/self-host)
