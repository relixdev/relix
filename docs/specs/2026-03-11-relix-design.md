# Relix — Design Specification

**Date:** 2026-03-11
**Status:** Approved
**Company:** Relix / relix.sh
**GitHub:** github.com/relixdev
**CLI:** relixctl

## Problem

Developers using AI coding CLIs (Claude Code, Aider, Codex CLI) can only interact with their sessions from the machine they started on. Walk away from your laptop and your agents are stuck — waiting for approvals, unable to receive new instructions, invisible to you.

## Solution

A React Native mobile app + lightweight agent CLI that lets you monitor and control your AI coding sessions from your phone. Anywhere, any network.

## Products

| Product | Description |
|---|---|
| **Relix Agent** (`relixctl`) | CLI tool installed on dev machines. Connects to Claude Code headless mode, phones home to relay. |
| **Relix Relay** | Hosted WebSocket router. Bridges agent ↔ mobile connections. E2E encrypted, open source. |
| **Relix Mobile** | React Native app (iOS + Android). Dashboard + session view. |
| **Relix Cloud** | Auth, billing, push notifications, machine registry. Proprietary. |

---

## Architecture

```
Relix Agent  →  Relix Relay (Cloud)  ←  Relix Mobile
   (CLI)          (WebSocket hub)        (React Native)
                       ↑
                  Relix Cloud
              (auth, billing, push)
```

### Key Architecture Decisions

- **All connections outbound** — agents and mobile both dial out to the relay. No port forwarding, no NAT traversal, works behind firewalls and on cellular.
- **E2E encrypted** — relay is a dumb pipe that routes opaque encrypted blobs. Cannot read session data even if compelled.
- **Open source relay + agent** (Bitwarden model) — self-hosters run the relay themselves, pay nothing. Builds trust with developers.
- **Claude Code first** — agent has pluggable `CopilotAdapter` interface for future tool support (Aider, Codex CLI), but only Claude Code adapter at launch.
- **Go for everything** — agent, relay, cloud. Single language, trivial cross-compilation, excellent WebSocket/concurrency support.

---

## E2E Encryption & Trust Model

### Pairing Flow

1. User installs `relixctl` on their machine, runs `relixctl login`
2. Agent generates an X25519 key pair locally. Private key never leaves the machine.
3. User opens Relix Mobile, goes to "Add Machine"
4. App generates its own X25519 key pair. Shows a 6-digit pairing code.
5. User enters code on their machine: `relixctl pair <code>`
6. Agent and app exchange public keys via the relay (pairing code proves both sides are the same user)
7. All subsequent traffic encrypted with NaCl box (X25519 + XSalsa20-Poly1305)

### What the Relay Sees

- Encrypted blobs with routing metadata (user ID, machine ID)
- Connection timestamps
- Message sizes
- **Never:** session content, code, file contents, tool call details

### Pairing Security

- Code expires after 5 minutes
- Maximum 5 attempts before code is invalidated (relay-enforced)
- Rate limited: 1 attempt per 3 seconds per IP
- After successful pairing, both sides display a 4-emoji SAS (Short Authentication String) for optional visual verification ("Do both devices show 🐶🌊🔑🎸?")

### Key Rotation

- Initiated by the agent every 30 days via the existing encrypted channel
- New X25519 key pair generated, public key sent to mobile encrypted with current key
- Grace period: old key accepted for 48 hours after rotation to handle offline devices
- If a device is offline for >30 days, re-pairing is required
- User can force rotate from the app ("Revoke machine" → re-pair)

### Self-Hosters

- Run `docker run ghcr.io/relixdev/relay`
- Point agents at their relay: `relixctl config set relay https://relay.mycompany.com`
- Point mobile app at their relay in settings
- Same encryption, they just own the pipe

---

## Claude Code Integration

### How the Agent Talks to Claude Code

Claude Code supports a headless mode via `claude --output-format stream-json` which outputs structured JSON events to stdout and accepts input on stdin. The agent uses this as the integration point.

### Session Discovery

The agent daemon watches for Claude Code processes on the machine:
1. Scans `/proc` (Linux) or `ps` (macOS) for running `claude` processes
2. Checks `~/.claude/projects/` for active session state files
3. When a new Claude Code process is detected, the agent spawns a bridge subprocess that connects to it via the headless JSON protocol
4. If Claude Code is already running interactively (not headless), the agent monitors its session state directory for events and provides read-only forwarding to mobile

### Message Flow (Agent ↔ Claude Code)

```
User types on phone
  → Mobile app encrypts message
  → Relay routes to agent
  → Agent decrypts, writes to Claude Code stdin as JSON: {"type":"user_message","content":"..."}
  → Claude Code processes, emits events on stdout
  → Agent reads structured events (assistant_message, tool_use, tool_result, permission_request)
  → Agent encrypts and forwards to relay
  → Relay routes to mobile
  → Mobile decrypts and renders
```

### Key Claude Code Events Handled

| Event Type | Description | Mobile UX |
|---|---|---|
| `assistant_message` | Claude's text response | Chat bubble / terminal text |
| `tool_use` | Claude wants to use a tool (edit file, run command) | Approval card with Allow/Deny |
| `tool_result` | Result of an approved tool use | Inline result card |
| `permission_request` | Claude needs explicit permission | High-priority push notification |
| `error` | Session error | Error banner + push notification |
| `session_end` | Session completed | Completion notification |

### Limitations

- If Claude Code changes its headless protocol, the agent adapter must be updated. This is our primary external dependency risk.
- Interactive-mode sessions (non-headless) get read-only forwarding — the user can observe but not send messages. Full control requires headless mode.

---

## Wire Protocol

### Envelope Format

All messages between agent, relay, and mobile use JSON over WebSocket:

```json
{
  "v": 1,
  "type": "session_event",
  "machine_id": "m_abc123",
  "session_id": "s_def456",
  "timestamp": 1741689600,
  "payload": "<base64 encrypted blob>"
}
```

The `payload` is always E2E encrypted. The relay reads only `v`, `type`, `machine_id`, `session_id`, and `timestamp` for routing.

### Message Types (Unencrypted Envelope)

| Type | Direction | Purpose |
|---|---|---|
| `auth` | client → relay | JWT authentication on connect |
| `session_list` | agent → relay → mobile | Enumerate active sessions on a machine |
| `session_event` | agent → relay → mobile | Claude Code event (encrypted payload) |
| `user_input` | mobile → relay → agent | User message or approval response (encrypted payload) |
| `approval_response` | mobile → relay → agent | Allow/Deny for a permission request (encrypted payload) |
| `ping` / `pong` | bidirectional | Keepalive (every 30 seconds) |
| `machine_status` | agent → relay | Online/offline/active status updates |

### Encrypted Payload Schema (After Decryption)

```json
{
  "kind": "assistant_message | tool_use | user_message | approval | ...",
  "data": { ... },
  "seq": 42
}
```

The `seq` field is a monotonically increasing sequence number per session, used for ordering and replay detection.

### Protocol Versioning

The `v` field in the envelope enables backward compatibility. The relay and clients negotiate the highest mutually supported version on connection. Old agents talking to a new relay continue to work as long as the relay supports their protocol version.

---

## Offline & Disconnection Handling

### Message Buffering

- The relay buffers up to 1000 messages (or 10MB, whichever is smaller) per machine when the mobile client is disconnected
- Buffer TTL: 24 hours. Messages older than 24 hours are dropped.
- When mobile reconnects, buffered messages are replayed in order using `seq` numbers

### Approval Timeout

- When Claude Code requests approval and the mobile client is unreachable:
  - Agent holds the approval request open (Claude Code is already waiting)
  - Push notification is sent via Relix Cloud (works even if WebSocket is down)
  - If no response within 30 minutes, the agent auto-denies the request and notifies mobile when it reconnects
  - Timeout is configurable per-machine in the app

### Agent Reconnection

- Agent automatically reconnects to relay on network changes (WiFi → cellular, VPN toggle, etc.)
- Exponential backoff: 1s, 2s, 4s, 8s... up to 60s max
- During disconnection, agent continues to bridge Claude Code events locally and buffers for relay delivery on reconnect

### Conflict Resolution

- If approval comes from both phone and laptop simultaneously, first response wins (agent processes the first `approval_response` it receives, ignores duplicates)
- Multiple phones paired to the same machine all receive events; first approval response wins

---

## Mobile App UX

### Screen Flow

1. **Onboarding** — GitHub OAuth or email signup
2. **Install Agent** — shows brew/curl command, waits for connection
3. **Pair Device** — 6-digit code on phone, enter on machine
4. **Dashboard (Home)** — all machines, pending approvals inline, status at a glance
5. **Session View** — tap into a machine, switchable Chat/Terminal modes

### Dashboard (Home Screen)

- Shows all connected machines with online/offline/active status
- Pending approval cards surfaced at the top with inline Allow/Deny buttons
- Approve from dashboard without entering the session for quick actions
- Bottom nav: Home, New Session (+), Settings

### Session View

Two modes (user's choice, toggle in session header):

**Chat Mode:**
- iMessage-style conversation bubbles
- Tool calls and file edits appear as tappable inline cards
- Tap to preview diffs
- Quick approve/deny for tool use
- Text input at bottom

**Terminal Mode:**
- Faithful reproduction of Claude Code terminal output
- Monospace rendering
- Same approve/deny affordances, styled as terminal prompts
- Keyboard input at bottom

### Push Notifications

- Tool approval requests (high priority — immediate delivery)
- Session errors/failures
- Session completions
- Configurable per-machine in settings

### Security

- Biometric auth (Face ID / fingerprint) to open app
- App lock after configurable timeout
- Active session data held in memory only — cleared when app is backgrounded for >5 minutes
- Session history (7-90 days depending on tier) stored encrypted on Relix Cloud, fetched on demand. Cloud stores E2E encrypted blobs — Relix cannot read them. History is decryptable only by devices holding the session keys.
- Session recording/playback (Plus+ tiers) uses the same encrypted cloud storage

---

## Agent CLI (`relixctl`)

### Commands

```
relixctl login          # Auth via browser (OAuth flow)
relixctl login --code   # Auth via device code (headless servers)
relixctl pair <code>    # Pair with mobile app using 6-digit code
relixctl status         # Show connection status, active sessions
relixctl sessions       # List Claude Code sessions on this machine
relixctl attach <id>    # Attach to existing Claude Code session
relixctl start          # Start daemon (auto-runs after login)
relixctl stop           # Stop daemon
relixctl config set <k> <v>  # e.g. relay URL for self-hosters
relixctl uninstall      # Clean removal (stops daemon, removes launchd/systemd service, deletes keys, removes config)
```

### Daemon Behavior

- Background service (launchd on macOS, systemd on Linux)
- Auto-starts on boot after first `relixctl login`
- Discovers running Claude Code sessions automatically
- Reconnects to relay on network changes
- ~10MB memory, negligible CPU when idle

### Plugin Architecture (Future)

Internal `CopilotAdapter` interface (Go):

```go
type CopilotAdapter interface {
    // Discover returns active sessions for this tool on the machine
    Discover() ([]Session, error)

    // Attach connects to a specific session and returns an event channel
    Attach(sessionID string) (<-chan Event, error)

    // Send delivers user input or an approval response to the session
    Send(sessionID string, msg UserInput) error

    // Detach cleanly disconnects from a session
    Detach(sessionID string) error
}

type Session struct {
    ID        string
    Tool      string    // "claude-code", "aider", etc.
    Project   string    // working directory
    Status    string    // "active", "waiting_approval", "idle"
    StartedAt time.Time
}

type Event struct {
    Kind string // "assistant_message", "tool_use", "permission_request", "error", "session_end"
    Data json.RawMessage
    Seq  uint64
}

type UserInput struct {
    Kind string // "message", "approval_response"
    Data json.RawMessage
}
```

Claude Code adapter is the only implementation at launch. Clean boundary for future Aider/Codex adapters.

---

## Monetization

### Tiers

| | **Free** | **Plus — $4.99/mo** | **Pro — $14.99/mo** | **Team — $24.99/user/mo** |
|---|---|---|---|---|
| Machines | 3 | 10 | Unlimited | Unlimited |
| Concurrent sessions | 2 | 5 | Unlimited | Unlimited |
| Push notifications | Full | Full | Full | Full + team alerts |
| Session history | 7 days | 30 days | 90 days | 90 days |
| View modes | Chat + Terminal | Chat + Terminal | Chat + Terminal | All |
| Session recording/playback | No | Yes | Yes | Yes |
| Priority relay | No | No | Yes | Yes |
| Shared sessions | No | No | No | Yes |
| Admin/SSO | No | No | No | Yes |

### Billing

- Stripe integration
- Monthly or annual (2 months free on annual)
- Free tier is genuinely useful (3 machines, 2 sessions) — drives adoption and word-of-mouth

### Conversion Strategy

- Free users become evangelists (3 machines is enough to love it)
- Plus ($4.99) is impulse-buy pricing — "a coffee a month"
- Pro catches power users hitting 10+ machines
- Team is inbound sales from engineering managers

---

## Tech Stack

| Component | Technology | Notes |
|---|---|---|
| **Relix Agent** | Go 1.23+ | Single binary, nhooyr.io/websocket, NaCl crypto |
| **Relix Relay** | Go 1.23+ | WebSocket router (in-memory connection state, no durable storage — clients reconnect on restart), Docker image, MIT license |
| **Relix Cloud** | Go + PostgreSQL + Redis | Auth, billing, push, machine registry |
| **Relix Mobile** | React Native (Expo), TypeScript | zustand state, react-navigation, expo-notifications |
| **E2E Crypto** | X25519 + XSalsa20-Poly1305 | Go: golang.org/x/crypto/nacl, RN: libsodium-wrappers |
| **Push** | APNs + FCM via Relix Cloud | Triggered on agent "approval needed" events |
| **CI/CD** | GitHub Actions | Goreleaser for Go binaries, EAS for mobile builds |
| **Infrastructure** | Fly.io | Global edge deployment for relay + cloud |
| **Monitoring** | Prometheus + Grafana | Connections, latency, throughput |

### Why Expo

- Managed build pipeline (EAS Build) — no Xcode/Android Studio
- OTA updates for JS changes (skip App Store review)
- Push notification support built in
- Bare workflow available if native modules needed later

---

## Repo Structure

```
github.com/relixdev/
├── relay/          # Open source relay server (MIT)
├── relixctl/       # Open source agent CLI (MIT)
├── mobile/         # Relix Mobile app (proprietary)
├── cloud/          # Relix Cloud backend (proprietary)
└── protocol/       # Shared protocol definitions (MIT)
```

Open source the trust-critical parts (relay, agent, protocol). Proprietary for revenue-generating parts (mobile, cloud).

---

## User Journey

### Day 1 (Setup)

1. Download Relix Mobile from App Store / Play Store
2. Sign up (GitHub OAuth or email)
3. App shows install instructions for agent
4. On laptop: `brew install relixdev/tap/relixctl`
5. `relixctl login` — opens browser, authenticates
6. App shows pairing code, user runs `relixctl pair <code>`
7. Machine appears in app. Done.

### Adding a Remote Server

1. SSH into server
2. `curl -fsSL relix.sh/install | sh`
3. `relixctl login --code` (device code flow, no browser needed)
4. `relixctl pair <code>` from app
5. Server appears in app instantly

### Daily Use

1. Start Claude Code session on laptop
2. Walk away
3. Phone buzzes: "Approval needed: Edit src/auth.ts"
4. Swipe to approve from notification, or open app for details
5. Continue conversation from phone if needed
6. Back at laptop — session continued seamlessly
