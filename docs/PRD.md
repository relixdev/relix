# Relix — Product Requirements Document

**Version:** 1.0
**Date:** 2026-03-13
**Status:** Active
**Owner:** Relix (relix.sh)
**Audience:** OpenClaw autonomous shipping bot — no human available for clarification. Every decision must be documented here.

---

## Table of Contents

1. [Product Overview](#1-product-overview)
2. [User Stories & Requirements](#2-user-stories--requirements)
3. [Functional Requirements](#3-functional-requirements)
   - 3.1 Agent (relixctl)
   - 3.2 Relay
   - 3.3 Cloud
   - 3.4 Mobile App
4. [Non-Functional Requirements](#4-non-functional-requirements)
5. [Monetization Requirements](#5-monetization-requirements)
6. [Success Metrics](#6-success-metrics)
7. [Open Questions & Risks](#7-open-questions--risks)

---

## 1. Product Overview

### 1.1 Vision

Relix is the universal command center for AI coding agents. Developers run Claude Code, Aider, Cline, and other tools across multiple machines — Relix connects all of them to a single mobile dashboard with end-to-end encryption, push notifications, and one-tap approvals.

### 1.2 Mission

Make AI coding agents truly mobile-native: notifications arrive on your phone, approvals take one tap, and you never have to run to your laptop because an agent is blocked.

### 1.3 Positioning

**Category:** AI agent control plane / mobile developer tools
**Tagline:** "Command center for all your AI coding agents"
**Anti-positioning:** Do NOT position as "Control Claude Code from your phone" — that is a feature Anthropic gives away free with claude.ai. Relix's differentiation is multi-tool, multi-machine, E2E encryption, and team features.

### 1.4 Target Users (Personas)

#### Persona 1: Solo Dev — "The Remote Worker" (Primary, Free → Plus)
- **Who:** Individual developer, 1-3 machines (laptop + cloud server or desktop)
- **Tools used:** Claude Code, possibly Aider
- **Pain:** Starts a long Claude Code task, walks to kitchen, agent blocks on an approval. Returns 20 minutes later to find nothing happened.
- **Goal:** Get push notifications when agent needs attention. Approve from phone without opening laptop.
- **Conversion trigger:** Hits the 2-session or 3-machine free tier limit within first week.
- **Willingness to pay:** Yes, at $4.99/mo — "coffee price"

#### Persona 2: Power Dev — "The Multi-Machine Developer" (Plus → Pro)
- **Who:** Developer with 4+ machines: laptop, desktop, 2+ cloud VMs for different projects
- **Tools used:** Claude Code + Aider or Cline, running concurrently across machines
- **Pain:** Juggling multiple terminals, missing approvals, no unified view of what all agents are doing
- **Goal:** Single dashboard, all machines, all agents, all approvals
- **Conversion trigger:** Hits 10-machine Plus limit, or wants priority relay for low-latency approval responses
- **Willingness to pay:** Yes, at $14.99/mo — serious tool spend

#### Persona 3: Engineering Manager — "The Team Lead" (Team)
- **Who:** EM at 5-50 person startup, team uses Claude Code or Cline for AI-assisted development
- **Pain:** No visibility into what agents are doing on team machines. Compliance requires audit trails. Devs occasionally approve risky operations without a second look.
- **Goal:** Shared session visibility, audit trail for agent actions, team-wide approval policies
- **Conversion trigger:** Inbound from a developer on the team who already uses Relix personally
- **Willingness to pay:** Yes, at $24.99/user/mo — IT/tooling budget

#### Persona 4: Self-Hoster — "The Privacy-First Dev" (Free, non-paying)
- **Who:** Security-conscious developer, potentially at a company with data policies that prevent using third-party relay
- **Pain:** Can't route code through any third-party cloud
- **Goal:** Run the relay on-premise. Use Relix mobile app pointed at their own relay.
- **Conversion trigger:** May upgrade to paid tier for push notifications (which require Relix Cloud) or team features
- **Note:** Self-hosters generate community credibility and GitHub stars, not direct revenue. Worth serving.

### 1.5 Core Value Proposition vs. Alternatives

#### vs. Anthropic Remote Control

Anthropic ships `claude remote-control` built into Claude Code, connected to claude.ai and the Claude mobile app. It is free on all Claude plans.

| Capability | Anthropic Remote Control | Relix |
|---|---|---|
| Tools supported | Claude Code only | Claude Code, Aider, Cline (P0/P1 launch); Cursor, Continue.dev (P2) |
| Multi-machine dashboard | No — per-session, one at a time | Yes — all machines, all sessions, one view |
| End-to-end encryption | No — decrypted at Anthropic servers | Yes — relay sees only ciphertext |
| Self-hostable relay | No | Yes — open source Docker image |
| Push notifications | No | Yes — APNs + FCM |
| Offline message queue | No — missed if phone not connected | Yes — 1000 msg / 24h buffer, delivered on reconnect |
| Automation rules | No | Yes — auto-approve reads, block destructive ops (P1) |
| Team shared sessions | No | Yes (Team tier) |
| Session recording/playback | No | Yes (Plus+ tiers) |
| Cross-session history | No | Yes — 7-90 days depending on tier |
| Price | Free | Free tier + paid plans |

**Key insight for messaging:** Users who already use Anthropic Remote Control are warm prospects — they have proven they want mobile agent control. Relix upsells them on multi-tool and privacy.

#### vs. No Solution (doing nothing)
Most developers simply walk back to their laptop. The status quo is high friction. Any working solution wins.

#### vs. tmux / ssh from phone
Some developers SSH to their machine from mobile SSH apps (Termius, Blink). This is unusable UX — tiny screen, no context, typing code on glass. Relix is purpose-built for approval workflows.

---

## 2. User Stories & Requirements

Priority codes:
- **P0** = Required for launch. Blocking.
- **P1** = Ship within 2 weeks of launch. High value.
- **P2** = Ship within 8 weeks. Nice to have.
- **P3** = Future. Don't build now.

### 2.1 Authentication Stories

**US-AUTH-001** (P0): As a new user, I can sign up with GitHub OAuth so I don't need to create a new password.
- Acceptance: GitHub OAuth completes, JWT issued, user record created in DB, redirected to onboarding
- Edge cases: GitHub account has no email → use GitHub username as display name; OAuth state mismatch → show error, offer retry

**US-AUTH-002** (P0): As a new user, I can sign up with email and password so I don't need a GitHub account.
- Acceptance: Email validated (format check + uniqueness), password min 8 chars, bcrypt stored, verification email sent, JWT issued after verification
- Edge cases: Email already exists → "Email taken, sign in?" prompt; verification link expires after 24h → resend option

**US-AUTH-003** (P0): As a returning user, I can log in on the mobile app and stay logged in across app restarts.
- Acceptance: JWT stored in Expo SecureStore, refreshed automatically before expiry; biometric prompt on app open (if enabled)
- Edge cases: Refresh token expired → back to login screen with "Session expired" message; no network → show cached dashboard in read-only mode

**US-AUTH-004** (P0): As a developer on a headless server, I can authenticate relixctl without a browser using a device code flow.
- Acceptance: `relixctl login --code` prints a URL and short code; user visits URL on any device, approves; CLI polls and receives JWT within 5 minutes
- Edge cases: Code expires (10 min) → "Code expired, run login --code again"; polling fails → exponential backoff, surface error after 5 min

**US-AUTH-005** (P0): As a user, I can log out of the mobile app and all my data is cleared from the device.
- Acceptance: JWT deleted from SecureStore, in-memory session data cleared, biometric lock reset, redirected to login screen
- Edge cases: Logout while active session open → warn "Active session will be disconnected", confirm before proceeding

### 2.2 Agent Installation & Setup Stories

**US-INSTALL-001** (P0): As a macOS developer, I can install relixctl with a single brew command.
- Acceptance: `brew install relixdev/tap/relixctl` installs latest binary; `relixctl --version` returns current version
- Edge cases: Homebrew not installed → brew install instructions shown; Apple Silicon vs Intel → correct binary auto-selected by Homebrew

**US-INSTALL-002** (P0): As a Linux developer, I can install relixctl with a curl one-liner.
- Acceptance: `curl -fsSL relix.sh/install | sh` downloads correct binary for architecture (x64/arm64, glibc/musl), places in `/usr/local/bin`, makes executable
- Edge cases: No sudo → install to `~/.local/bin` with PATH note; musl vs glibc auto-detected via `ldd`; curl not available → wget fallback

**US-INSTALL-003** (P0): As a developer, I can authenticate relixctl to my Relix account.
- Acceptance: `relixctl login` opens browser to relix.sh/auth, completes OAuth, JWT stored in `~/.config/relixctl/credentials.json` (mode 0600), prints "Logged in as {email}"
- Edge cases: Browser won't open (headless) → prints URL + suggests `relixctl login --code`; already logged in → "Already logged in as X. Use --force to re-authenticate"

**US-INSTALL-004** (P0): As a developer, I can pair my machine with the mobile app using a 6-digit code.
- Acceptance: App shows 6-digit code; `relixctl pair <code>` exchanges keys, both sides show 4-emoji SAS for verification; machine appears in app dashboard within 10 seconds
- Edge cases: Wrong code → "Invalid or expired code"; code expired → "Code expired, generate a new one in the app"; SAS mismatch → user can choose to re-pair or abort

**US-INSTALL-005** (P0): The relixctl daemon starts automatically after login without requiring any additional steps.
- Acceptance: After `relixctl login` + `relixctl pair`, daemon is registered with launchd (macOS) or systemd (Linux) and starts on next boot; runs immediately in background
- Edge cases: launchd/systemd unavailable → daemon runs in foreground with note to add to init system manually; plist/service file already exists → overwrite silently

### 2.3 Session Discovery & Control Stories

**US-SESSION-001** (P0): As a mobile user, I can see all Claude Code sessions running on my paired machines in real time.
- Acceptance: Dashboard shows machine list; each machine shows active session count; tapping machine shows session list with project path, status, and start time
- Edge cases: Machine offline → show "Offline" badge, last-seen timestamp; no active sessions → "No active sessions — start one from your terminal"

**US-SESSION-002** (P0): As a mobile user, I receive a push notification when a Claude Code session needs my approval.
- Acceptance: Push arrives within 5 seconds of approval request; notification shows machine name, session project, and brief description of the tool use (e.g., "Edit src/auth.ts"); tapping opens approval detail
- Edge cases: Phone in DND → notification queued and delivered when DND clears; app killed → system push wakes it; multiple approvals → batched after 3 in 30 seconds into "X approvals pending"

**US-SESSION-003** (P0): As a mobile user, I can approve or deny a tool use request from the notification without opening the app.
- Acceptance: iOS notification has "Allow" and "Deny" action buttons; tapping either sends approval response within 3 seconds; Claude Code resumes within 5 seconds of response
- Edge cases: Response fails to deliver → notification remains actionable for 30 minutes; agent auto-denies after 30 minutes with "Approval timeout" event

**US-SESSION-004** (P0): As a mobile user, I can view the full conversation history of an active session in Chat mode.
- Acceptance: Session view shows all messages in chronological order; assistant messages as left-aligned bubbles, user messages right-aligned; tool use shown as inline cards with status (pending/approved/denied/completed); scrolls to bottom on new messages
- Edge cases: Long session history (>500 messages) → paginate, load earlier messages on scroll-up; session disconnected mid-view → show "Disconnected" banner, attempt reconnect

**US-SESSION-005** (P0): As a mobile user, I can view session output in Terminal mode that faithfully renders ANSI escape codes.
- Acceptance: Terminal mode uses monospace font; ANSI colors rendered correctly (at minimum: 16-color, ideally 256-color); bold/italic/underline supported; dark background theme; horizontal scroll for long lines
- Edge cases: Unsupported ANSI sequences → strip silently, don't show garbage; terminal wider than screen → horizontal scroll; rapid output (>100 lines/sec) → throttle rendering to avoid jank

**US-SESSION-006** (P0): As a mobile user, I can send a text message to an active Claude Code session.
- Acceptance: Text input at bottom of session view; send button or return key sends; message appears as user bubble immediately (optimistic); delivered to Claude Code within 2 seconds; typing indicator shown while waiting for response
- Edge cases: Session disconnected → message queued locally, delivered on reconnect (within session) or error shown if session ended; empty message → send button disabled

**US-SESSION-007** (P1): As a mobile user, I can start a new Claude Code session on a connected machine from the app.
- Acceptance: "+" button opens modal with machine selector and working directory input; submits to agent; agent spawns headless Claude Code session; session appears in list within 5 seconds
- Edge cases: Machine offline → "Machine is offline, cannot start session"; invalid working directory → error from agent bubbled to UI; session fails to start → error message with stderr output

**US-SESSION-008** (P1): As a mobile user, I can see session history from the past 7 days (Free) / 30 days (Plus) / 90 days (Pro/Team) even for sessions that have ended.
- Acceptance: Completed sessions listed under "History" tab per machine; tap to replay; history stored as encrypted blobs on Relix Cloud
- Edge cases: History beyond tier limit → show lock icon with "Upgrade to see older history"; history for a deleted machine → retain for tier TTL then delete

### 2.4 Security Stories

**US-SEC-001** (P0): As a user, my session data is never readable by Relix's servers.
- Acceptance: All session payloads encrypted with NaCl box before leaving the agent; relay stores and forwards only ciphertext; E2E encryption verified by independent code audit before launch
- Edge cases: Key loss (device wiped) → history is unrecoverable (by design); document this clearly in onboarding

**US-SEC-002** (P0): As a mobile user, I can require biometric authentication to open the Relix app.
- Acceptance: Biometric prompt (Face ID / fingerprint) appears on cold open; disabled by default, enabled in Settings > Security; falls back to device PIN if biometric fails 3 times
- Edge cases: Device has no biometric → option not shown; biometric enrollment changes → require re-authentication once, then proceed

**US-SEC-003** (P0): As a user, the app automatically locks after being backgrounded for more than 5 minutes.
- Acceptance: Timer starts on app background; on foreground after timer elapsed → biometric prompt shown (if enabled) OR data cleared and PIN shown; default timeout = 5 minutes; configurable to 1/5/15/30 min or never in Settings
- Edge cases: App backgrounded mid-approval → approval card preserved in memory if timer not elapsed; lock clears in-memory session data but not stored credentials

**US-SEC-004** (P0): As a user, I can unpair a machine from the app, which revokes its access immediately.
- Acceptance: "Revoke" button in Machine Settings sends revocation to Cloud; Cloud marks machine as revoked; relay rejects connections from revoked machine_id; agent on machine shows "Access revoked" error
- Edge cases: Agent offline when revoked → revocation takes effect on next connection attempt; re-pairing a revoked machine requires new pair code

**US-SEC-005** (P1): As a user, I can set automation rules that auto-approve specific tool types.
- Acceptance: Rule editor in Settings > Machines > [machine] > Rules; rule types: auto-approve file reads, auto-approve shell commands matching pattern, block destructive ops (rm -rf, DROP TABLE etc.); rules applied by agent's PreToolUse hook
- Edge cases: Rule conflict (auto-approve + block same pattern) → block wins; rule syntax error → validation at save time with clear error message

### 2.5 Monetization Stories

**US-MON-001** (P0): As a new user, I can use Relix free with 3 machines and 2 concurrent sessions with no credit card required.
- Acceptance: Free tier enforced at API level; attempting to add 4th machine returns 402 with upgrade prompt in app; free tier fully functional, no nag screens

**US-MON-002** (P0): As a free user, I can upgrade to Plus, Pro, or Team from within the app.
- Acceptance: Settings > Billing shows current tier and upgrade options; tapping upgrade opens Stripe Checkout in a webview; successful payment upgrades tier immediately (Stripe webhook → DB update → push notification "You're now on Plus!")
- Edge cases: Payment fails → show Stripe error message, remain on current tier; subscription created but webhook delayed → poll for up to 60 seconds, then show "Processing..." and notify when ready

**US-MON-003** (P0): As a paid user, I can cancel my subscription and revert to the free tier at the end of the billing period.
- Acceptance: Settings > Billing > Cancel Subscription; confirms cancellation; Stripe marks for cancellation at period end; app shows "Cancels on [date]" state; on expiry, tier downgraded, data beyond free tier limits retained for 7 days then deleted
- Edge cases: User on Pro with 15 machines cancels → on downgrade to Free, oldest 12 machines shown as "Inactive" with option to remove or upgrade; active sessions on inactive machines disconnected

**US-MON-004** (P1): As a Team admin, I can invite team members and manage their access.
- Acceptance: Settings > Team > Invite; email invitation sent; invited user joins team on account creation or links existing account; admin can remove members; member removal revokes their relay token immediately
- Edge cases: Invite to email that already has Relix account → accept invite flow; team admin leaves team → must transfer ownership or delete team first

### 2.6 Self-Hosting Stories

**US-SELFHOST-001** (P1): As a self-hoster, I can run the Relix relay on my own infrastructure with a single Docker command.
- Acceptance: `docker run -e JWT_SECRET=xxx -p 8080:8080 ghcr.io/relixdev/relay` starts a functional relay; agent can connect with `relixctl config set relay https://my-relay.example.com`; mobile app can point to custom relay in Settings > Advanced > Custom Relay URL
- Edge cases: Custom relay requires auth → relay supports JWT_SECRET env var, same JWT scheme as hosted; SSL termination handled by reverse proxy (nginx/caddy), not the relay binary itself

---

## 3. Functional Requirements

### 3.1 Agent (relixctl)

#### 3.1.1 Installation

**Homebrew (macOS — Primary)**

Tap repository: `github.com/relixdev/homebrew-tap`
Formula: `relixdev/tap/relixctl`
Install command: `brew install relixdev/tap/relixctl`

The Homebrew formula must:
- Define the binary download URLs for `darwin-arm64` and `darwin-x64`
- Compute SHA256 checksums for each release
- Install to `$(brew --prefix)/bin/relixctl`
- Include a completion script for zsh and bash
- Goreleaser generates the formula automatically on release via the `brews` config block

**curl installer (Linux + macOS fallback)**

URL: `https://relix.sh/install` (served from Cloud or CDN)
Script behavior:
1. Detect OS: `$(uname -s)` → `darwin` or `linux`
2. Detect arch: `$(uname -m)` → `x86_64` → `amd64`, `aarch64` → `arm64`
3. Detect libc (Linux only): run `ldd --version 2>&1 | head -1`; if output contains "musl" → use musl binary; else → use glibc binary
4. Construct download URL: `https://github.com/relixdev/relixctl/releases/latest/download/relixctl_linux_amd64` (or appropriate variant)
5. Determine install path: try `/usr/local/bin` (test writability with `-w`); if not writable and `$HOME/.local/bin` exists → use that; else create `$HOME/.local/bin` and add PATH note
6. Download binary, verify SHA256 against `{binary_url}.sha256`, make executable
7. Print success: `relixctl installed to /usr/local/bin/relixctl. Run: relixctl login`

**Binary release targets (Goreleaser)**
- `linux-amd64` (glibc)
- `linux-arm64` (glibc)
- `linux-amd64-musl` (static, CGO_ENABLED=0)
- `linux-arm64-musl` (static, CGO_ENABLED=0)
- `darwin-amd64`
- `darwin-arm64`
- `windows-amd64` (future, not at launch)

All binaries signed. macOS binaries notarized via GitHub Actions with Apple Developer ID.

#### 3.1.2 Authentication (Login)

**Command:** `relixctl login`
**Command:** `relixctl login --code` (device code flow for headless)

**Browser flow (`relixctl login`):**
1. Generate random `state` parameter (32 bytes, hex encoded)
2. Start local HTTP server on a random available port (try 9876-9886, use first available)
3. Open browser to `https://cloud.relix.sh/auth/cli?state={state}&redirect_uri=http://localhost:{port}/callback`
4. Wait for HTTP callback at `/callback?code={code}&state={state}`
5. Validate state matches
6. Exchange code for JWT: POST `https://cloud.relix.sh/api/v1/auth/cli/exchange` with `{"code": "{code}"}`
7. Response: `{"token": "...", "user": {"id": "...", "email": "...", "tier": "free"}}`
8. Store JWT in `~/.config/relixctl/credentials.json` (permissions: 0600)
9. Print: `Logged in as {email}`
10. Auto-run `relixctl start` to launch daemon

Credentials file format:
```json
{
  "token": "eyJ...",
  "user_id": "u_abc123",
  "email": "user@example.com",
  "relay_url": "wss://relay.relix.sh",
  "created_at": "2026-03-13T00:00:00Z"
}
```

**Device code flow (`relixctl login --code`):**
1. POST `https://cloud.relix.sh/api/v1/auth/device/start` → response: `{"device_code": "...", "user_code": "XXXX-XXXX", "verification_url": "https://relix.sh/activate", "expires_in": 600, "interval": 5}`
2. Print:
   ```
   Open https://relix.sh/activate and enter code: XXXX-XXXX
   Waiting for authentication...
   ```
3. Poll `https://cloud.relix.sh/api/v1/auth/device/token` every 5 seconds with `{"device_code": "..."}`
4. Responses: `{"error": "authorization_pending"}` → continue polling; `{"error": "expired_token"}` → exit with error; `{"token": "..."}` → success
5. On success: same credential storage and `relixctl start` as browser flow

**Already authenticated:** If credentials file exists and token is valid, `relixctl login` prints "Already logged in as {email}. Use --force to re-authenticate." and exits 0. With `--force`, deletes credentials and starts fresh.

**Token validation:** On every command, agent reads credentials file and checks JWT expiry. If within 7 days of expiry, silently refreshes: POST `https://cloud.relix.sh/api/v1/auth/refresh` with current token, store new token. If expired, print "Session expired. Run: relixctl login" and exit 1.

#### 3.1.3 Pairing (Key Exchange)

**Command:** `relixctl pair <6-digit-code>`

Full pairing sequence:
1. Validate code format: exactly 6 digits (0-9), else error "Code must be 6 digits"
2. Load credentials (must be authenticated, else "Not logged in. Run: relixctl login")
3. Check if machine already has a keypair: read `~/.config/relixctl/keys.json`. If not, generate X25519 keypair, store private key (base64) and public key (base64) in `keys.json` (mode 0600)
4. Connect to relay WebSocket: `wss://relay.relix.sh/pair`
5. Send auth message: `{"type": "auth", "token": "{jwt}"}`
6. Wait for `{"type": "auth_ok", "machine_id": "m_xxx"}` — relay assigns machine_id on first pair
7. Send pair request: `{"type": "pair_request", "code": "{code}", "public_key": "{base64}", "machine_id": "{id}", "machine_name": "{hostname}"}`
8. Relay validates code (5-min TTL, 5-attempt max, rate-limited)
9. On success, relay forwards mobile's public key: `{"type": "pair_complete", "mobile_public_key": "{base64}", "sas": ["🐶", "🌊", "🔑", "🎸"]}`
10. Store mobile public key in `~/.config/relixctl/peers.json`
11. Print:
    ```
    Pairing complete!
    Verify on your phone: 🐶 🌊 🔑 🎸
    If the emoji match, your connection is secure.
    Machine will now appear in your Relix app.
    ```
12. On failure:
    - Code not found: "Invalid or expired code. Generate a new code in the Relix app."
    - Rate limited: "Too many attempts. Wait 60 seconds and try again."
    - Timeout (30s): "Pairing timed out. Is the Relix app open on your phone?"

**SAS Generation:** The relay derives the 4-emoji SAS from HKDF-SHA256 of both public keys concatenated. Both sides compute independently from their copy of the other's public key — if they match, MITM is ruled out. SAS emoji pool: 64 distinct, visually distinct emoji (animals, objects, symbols). Index = first 4 bytes of HKDF output, each byte mod 64.

**Keys storage format** (`~/.config/relixctl/keys.json`, mode 0600):
```json
{
  "machine_id": "m_abc123",
  "private_key": "{base64-32-bytes}",
  "public_key": "{base64-32-bytes}",
  "created_at": "2026-03-13T00:00:00Z"
}
```

**Peers storage format** (`~/.config/relixctl/peers.json`, mode 0600):
```json
{
  "peers": [
    {
      "mobile_id": "mob_abc123",
      "public_key": "{base64-32-bytes}",
      "added_at": "2026-03-13T00:00:00Z"
    }
  ]
}
```

#### 3.1.4 Session Discovery

The agent continuously discovers Claude Code sessions. Discovery runs on startup, then every 30 seconds.

**Filesystem scan:**
1. Enumerate `~/.claude/projects/*/` directories
2. For each project directory, list `.jsonl` files sorted by modification time (most recent first)
3. Each `.jsonl` file is a session. Filename is the session ID.
4. Parse last line of `.jsonl` to determine session status: if last event `type` is `assistant` or `tool_use` and timestamp within 10 minutes → `active`; if `session_end` event exists → `completed`; else → `idle`

**Process detection (optional, belt-and-suspenders):**
1. Scan process table for processes matching `claude` binary name
2. Read `/proc/{pid}/cmdline` (Linux) or `ps aux` (macOS) to extract `--resume` session ID if present
3. Mark those sessions as `running` (has live process)

**Session object:**
```go
type Session struct {
    ID          string
    Tool        string    // "claude-code"
    ProjectPath string    // decoded from directory name
    Status      string    // "active", "idle", "waiting_approval", "completed"
    HasProcess  bool      // live process detected
    StartedAt   time.Time
    UpdatedAt   time.Time
}
```

**Project path encoding:** Claude Code encodes the working directory as the directory name under `~/.claude/projects/`. The encoding is: replace `/` with `-SLASH-`, replace spaces with `-SPACE-`. Reverse to decode. (Verify this against actual Claude Code behavior before shipping — see Section 7 risks.)

**Session list message:** Every 30 seconds (or immediately on change), agent sends to relay:
```json
{
  "type": "session_list",
  "machine_id": "m_abc123",
  "sessions": [
    {
      "id": "s_def456",
      "tool": "claude-code",
      "project": "/Users/zach/myproject",
      "status": "waiting_approval",
      "started_at": 1741689600,
      "updated_at": 1741689900
    }
  ]
}
```

#### 3.1.5 Session Bridging

The agent bridges Claude Code's stream-json protocol to the Relix wire protocol.

**Attaching to an existing session (read-forwarding mode):**
1. Tail the session `.jsonl` file using `fsnotify` (inotify/kqueue)
2. On new line appended: parse JSON event, map to Relix event type, encrypt payload, send to relay
3. This is read-only — user cannot send messages to an existing interactive session
4. If file is deleted or session process ends → send `session_end` event, remove from session list

**Attaching to a new headless session (full bidirectional mode):**
1. Spawn: `claude --resume {session_id} -p --input-format stream-json --output-format stream-json --verbose`
   Or for new sessions: `claude -p --input-format stream-json --output-format stream-json --verbose` in the project directory
2. Hold stdin open (never close until session ends or detach)
3. Read stdout line-by-line (each line is a JSON event)
4. Parse events (see Event Mapping below)
5. For each event: encrypt with NaCl box using mobile's public key, send to relay as `session_event`
6. For incoming `user_input` or `approval_response` messages from relay: decrypt, dispatch to Claude Code stdin

**Event Mapping (Claude Code → Relix):**

| Claude Code stream-json event | Relix Event Kind | Notes |
|---|---|---|
| `{"type": "assistant", "message": {...}}` | `assistant_message` | Forward message content as-is |
| `{"type": "user", "message": {...}}` | `user_message` | Echo back to mobile for confirmation |
| `{"type": "tool_use", "tool": {...}}` | `tool_use` | Includes tool name, input params |
| `{"type": "tool_result", ...}` | `tool_result` | Result or error from tool execution |
| `{"type": "system", "subtype": "init", ...}` | `session_start` | Session metadata |
| `{"type": "result", "subtype": "error_max_turns"}` | `session_end` | Session ended |
| `{"type": "result", "subtype": "success"}` | `session_end` | Session completed successfully |
| Lines that fail JSON parse | Dropped | Log to debug, never crash |

**Sending user input to Claude Code stdin:**
```json
{"type": "user", "message": {"role": "user", "content": "{text}"}}
```
Followed by a newline. The Claude Code binary reads this and processes as user message.

**Sending approval to Claude Code:**
Claude Code does not have a separate approval wire format for `--input-format stream-json`. Approvals work via the hooks system:
- When a `PreToolUse` hook is registered, it can return `{"decision": "allow"}` or `{"decision": "block", "reason": "..."}` via stdout
- The agent registers a PreToolUse hook via `~/.claude/settings.json` `hooks` section on `relixctl start`
- Hook is a small binary or shell script that communicates with the daemon via a local Unix socket
- Daemon holds approval requests, waits for mobile response (or timeout), writes decision to Unix socket
- Hook reads decision from socket and exits with appropriate output

**Hook installation (`~/.claude/settings.json`):**
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "relixctl hook-handler"
          }
        ]
      }
    ]
  }
}
```

The `relixctl hook-handler` subcommand:
1. Reads hook event JSON from stdin (provided by Claude Code)
2. Sends to daemon via Unix socket at `~/.config/relixctl/daemon.sock`
3. Waits for approval response (blocks until daemon responds or timeout)
4. Writes `{"decision": "allow"}` or `{"decision": "block", "reason": "Denied by user"}` to stdout
5. Exits 0

If the daemon is not running, `hook-handler` defaults to `allow` (fail-open) to avoid blocking Claude Code.

#### 3.1.6 Daemon Lifecycle

**Start:** `relixctl start`
1. Check if already running: read PID from `~/.config/relixctl/daemon.pid`, check process liveness (`kill -0 {pid}`)
2. If running: print "Daemon already running (pid {pid})" and exit 0
3. Fork to background (double-fork, write PID to daemon.pid)
4. Open Unix socket at `~/.config/relixctl/daemon.sock`
5. Connect to relay WebSocket with JWT auth
6. Start session discovery loop (every 30s)
7. Start ping/pong keepalive (every 30s)
8. Log to `~/.config/relixctl/daemon.log` (rotate at 10MB, keep 3)

**Stop:** `relixctl stop`
1. Read PID from `~/.config/relixctl/daemon.pid`
2. Send SIGTERM
3. Wait up to 10 seconds for clean shutdown
4. If not exited, SIGKILL
5. Remove daemon.pid
6. Print "Daemon stopped"

**Status:** `relixctl status`
Output:
```
Daemon:    running (pid 12345)
Relay:     connected (wss://relay.relix.sh)
Sessions:  3 active
User:      user@example.com
Machine:   m_abc123 (my-macbook)
Tier:      free
```
If not running: "Daemon: not running. Start with: relixctl start"

**Auto-start on boot:**

macOS (launchd):
- Plist location: `~/Library/LaunchAgents/sh.relix.relixctl.plist`
- `RunAtLoad: true`, `KeepAlive: true` with crash restart
- `relixctl start` writes the plist and runs `launchctl load`
- `relixctl stop --permanent` unloads and removes plist

Linux (systemd user session):
- Service file: `~/.config/systemd/user/relixctl.service`
- `WantedBy=default.target`, `Restart=on-failure`
- `relixctl start` writes service file, runs `systemctl --user enable relixctl && systemctl --user start relixctl`
- Requires `loginctl enable-linger {username}` for boot autostart without login session — print this instruction if linger not enabled

Linux (headless servers without systemd):
- Fall back to cron: add `@reboot relixctl start` to crontab
- `relixctl start` detects no systemd, attempts cron fallback, prints which method was used

#### 3.1.7 Configuration Management

**Command:** `relixctl config set <key> <value>`
**Command:** `relixctl config get <key>`
**Command:** `relixctl config list`

Config file: `~/.config/relixctl/config.json` (mode 0644)

```json
{
  "relay_url": "wss://relay.relix.sh",
  "approval_timeout_seconds": 1800,
  "log_level": "info",
  "machine_name": "my-macbook",
  "auto_approve_reads": false
}
```

Supported config keys:

| Key | Type | Default | Description |
|---|---|---|---|
| `relay_url` | string | `wss://relay.relix.sh` | Relay WebSocket URL. For self-hosters. |
| `approval_timeout_seconds` | int | 1800 (30 min) | Seconds before auto-deny if no response |
| `log_level` | string | `info` | `debug`, `info`, `warn`, `error` |
| `machine_name` | string | `hostname` | Display name in mobile app |
| `auto_approve_reads` | bool | false | Auto-approve file read operations without prompting |
| `cloud_url` | string | `https://cloud.relix.sh` | Cloud API URL. For self-hosters. |

`relixctl config set relay_url wss://my-relay.example.com` — updates the key, restarts daemon if running.

**Precedence:** CLI flag > environment variable > config file > default
**Environment variable naming:** `RELIXCTL_{KEY_UPPERCASE}` (e.g., `RELIXCTL_RELAY_URL`)

#### 3.1.8 Uninstall

**Command:** `relixctl uninstall`

Steps (with confirmation prompt):
1. "This will remove relixctl, your credentials, and all paired machines. Continue? [y/N]"
2. Stop daemon (SIGTERM, wait 5s)
3. Unload and remove launchd plist (macOS) or systemd service (Linux)
4. Remove hooks from `~/.claude/settings.json` (remove only the relixctl entries, leave other hooks intact)
5. Remove `~/.config/relixctl/` directory entirely
6. Print: "relixctl uninstalled. Remove the binary with: rm $(which relixctl)"
   (Don't remove the binary itself — it was installed by brew or the user and shouldn't be auto-deleted)
7. API call: POST `https://cloud.relix.sh/api/v1/machines/{machine_id}/unregister` to mark machine as removed in Cloud

#### 3.1.9 Multi-Tool Adapter System

The `CopilotAdapter` interface allows the agent to support multiple AI coding tools without changing the core relay logic.

**Interface (Go):**
```go
type CopilotAdapter interface {
    Name() string
    Discover() ([]Session, error)
    Attach(sessionID string, opts AttachOptions) (*SessionBridge, error)
}

type AttachOptions struct {
    ReadOnly bool   // true for existing interactive sessions
    WorkDir  string // for spawning new sessions
}

type SessionBridge struct {
    Events <-chan Event
    Send   func(UserInput) error
    Close  func() error
}
```

**Claude Code adapter** (ships at launch):
- `Name()` returns `"claude-code"`
- `Discover()` scans `~/.claude/projects/`
- `Attach()` spawns headless process or tails .jsonl

**Aider adapter** (P1, +2 weeks post-launch):
- Aider is a Python CLI: `aider --model claude-3-5-sonnet-20241022`
- Integration surface: Aider supports `--input-history-file` and `--output-chat-history-file`; also has `--watch-files` mode
- Strategy: Aider does not have a native stream-json mode. Approach: spawn Aider as a subprocess with a PTY; parse its output (markdown-formatted conversation); intercept its "Apply edits? y/n" prompts for approval flow
- Discovery: scan for running `aider` processes via process table; parse `--input-history-file` path from cmdline
- Note: Aider integration requires PTY handling (golang.org/x/term) and ANSI-aware output parsing — non-trivial. Allocate dedicated sprint.

**Cline adapter** (P1, +4 weeks post-launch):
- Cline is a VS Code extension (TypeScript)
- Integration surface: Cline writes state to VS Code's global storage; also exposes actions via VS Code's extension host
- Strategy: Cline has an open source codebase — read it to identify state files and IPC mechanisms. Primary approach: watch Cline's task files at `~/.config/Code/User/globalStorage/saoudrizwan.claude-dev/tasks/`; parse task JSON for conversation and approval state
- Approval injection: Cline supports MCP. Consider implementing a Relix MCP server that Cline connects to, which allows injecting approval decisions as MCP responses. This is cleaner than scraping files.
- Discovery: check for running `code` process + Cline extension loaded (check extension storage directory exists and has recent modification)

---

### 3.2 Relay

#### 3.2.1 WebSocket Connection Lifecycle

**Server:** Go HTTP server with WebSocket upgrade at `/ws`
**Library:** `nhooyr.io/websocket` (already in use per codebase)
**TLS:** Terminated by reverse proxy (nginx or Fly.io's anycast). Relay binds to `0.0.0.0:8080` HTTP only.

**Connection sequence:**
1. Client (agent or mobile) opens WebSocket to `wss://relay.relix.sh/ws`
2. Immediately sends auth message (within 5 seconds, else connection closed):
   ```json
   {"type": "auth", "token": "{jwt}", "client_type": "agent|mobile", "machine_id": "m_xxx", "protocol_version": 1}
   ```
3. Relay validates JWT against `JWT_SECRET` env var (HS256)
4. Extracts `user_id` from JWT claims
5. Registers connection in in-memory hub: `map[userID][]Connection`
6. Sends auth acknowledgment:
   ```json
   {"type": "auth_ok", "machine_id": "m_xxx", "connection_id": "conn_yyy"}
   ```
7. Starts ping/pong loop: relay sends `{"type": "ping"}` every 30 seconds; client must respond with `{"type": "pong"}` within 10 seconds; failure → close connection

**Connection types:**
- `agent`: represents a machine running relixctl. Identified by `machine_id`. One machine has one active agent connection. If a second agent connection arrives with the same `machine_id`, the relay closes the older one.
- `mobile`: represents a mobile app session. One user can have multiple mobile connections (multiple devices). All receive the same events.

**Disconnection handling:**
1. Relay detects close (normal, error, or ping timeout)
2. Removes connection from hub
3. If agent connection: mark machine as offline in hub; buffering begins for that machine's outbound messages
4. If mobile connection: no special action; if no mobile connections remain for user, buffering for inbound messages begins

#### 3.2.2 Authentication (JWT Validation)

JWT format (HS256):
```json
{
  "sub": "u_abc123",
  "email": "user@example.com",
  "tier": "free",
  "machine_id": "m_abc123",
  "iat": 1741689600,
  "exp": 1744281600
}
```

Relay validates:
1. Signature (HS256 with `JWT_SECRET` env var)
2. `exp` not past
3. `sub` present (user ID)

Relay does NOT make any HTTP calls for validation — purely local JWT verification. This keeps the relay stateless and fast.

**JWT_SECRET:** Shared between Cloud (issuer) and Relay (validator). Must be the same value. In production: Fly.io secret injected as env var on both services.

#### 3.2.3 Message Routing

**Agent → Mobile routing:**
1. Relay receives message from agent connection (identified by machine_id)
2. Extract `machine_id` and `user_id` from connection context
3. Find all active mobile connections for `user_id`
4. Forward message to all mobile connections (fan-out)
5. If no mobile connections active: add to buffer (see 3.2.4)

**Mobile → Agent routing:**
1. Relay receives message from mobile connection
2. Extract `machine_id` from message envelope (`"machine_id"` field required on all mobile→agent messages)
3. Find active agent connection for `machine_id`
4. Verify `user_id` of the agent matches `user_id` of the sending mobile connection (authorization check)
5. If authorized: forward message to agent
6. If agent not connected: return `{"type": "error", "code": "agent_offline", "message": "Machine is offline"}` to mobile
7. If unauthorized: return `{"type": "error", "code": "forbidden", "message": "Machine does not belong to your account"}` to mobile

**Message pass-through:** Relay does NOT inspect, modify, or decrypt the `payload` field. It only reads envelope fields for routing: `type`, `machine_id`, `session_id`, `timestamp`.

#### 3.2.4 Buffering

**Buffer scope:** Per `(user_id, machine_id)` pair. Buffer stores messages destined for mobile while mobile is disconnected.

**Buffer limits:**
- Max messages: 1000 per machine
- Max size: 10MB per machine
- TTL: 24 hours from message creation time

**Buffer behavior:**
- On mobile connect: drain buffer in order (oldest first) before forwarding live messages. Send drain complete marker: `{"type": "buffer_drained", "count": N}`
- On buffer full (1000 messages or 10MB): drop oldest messages, log dropped count metric
- On TTL expiry: background goroutine sweeps expired entries every 5 minutes

**Buffer is in-memory only.** Relay restart = buffer lost. This is acceptable per design (relay is explicitly a dumb pipe; clients handle reconnection). Document in self-hosting guide.

#### 3.2.5 Pairing Relay

The relay facilitates the pairing key exchange without being able to read the exchanged keys (they're public keys — no secrecy needed, but integrity matters).

**Pairing WebSocket endpoint:** `/pair` (separate from `/ws`)

**Flow:**
1. Mobile connects to `/pair`, authenticates with JWT, sends:
   ```json
   {"type": "request_pair_code", "mobile_public_key": "{base64}"}
   ```
2. Relay generates 6-digit code (crypto/rand, not math/rand), stores `{code: {user_id, mobile_public_key, expires_at: now+5min, attempts: 0}}` in memory
3. Relay returns: `{"type": "pair_code", "code": "847291"}`
4. Mobile displays code to user
5. Agent connects to `/pair`, authenticates, sends:
   ```json
   {"type": "claim_pair_code", "code": "847291", "agent_public_key": "{base64}", "machine_id": "m_xxx", "machine_name": "my-laptop"}
   ```
6. Relay looks up code:
   - Not found: `{"type": "error", "code": "invalid_code"}`
   - Expired: `{"type": "error", "code": "expired_code"}`
   - Wrong user (code was generated by different user's mobile): `{"type": "error", "code": "invalid_code"}` (don't reveal why)
   - Max attempts (5): `{"type": "error", "code": "code_locked"}`
   - Attempts incremented on each claim attempt
7. Relay computes SAS: HKDF-SHA256 of `{agent_public_key || mobile_public_key}`, take first 4 bytes, map to 4 emoji
8. Relay sends to mobile (on pairing WebSocket):
   ```json
   {"type": "pair_complete", "agent_public_key": "{base64}", "machine_id": "m_xxx", "machine_name": "my-laptop", "sas": ["🐶", "🌊", "🔑", "🎸"]}
   ```
9. Relay sends to agent (on pairing WebSocket):
   ```json
   {"type": "pair_complete", "mobile_public_key": "{base64}", "sas": ["🐶", "🌊", "🔑", "🎸"]}
   ```
10. Relay deletes the code from memory
11. Both connections close their pairing WebSocket

**Rate limiting (per IP):**
- Max 1 pairing attempt per 3 seconds per IP
- Max 10 pairing initiations (step 1) per hour per IP
- Implemented with token bucket in memory

**Code collision:** If code already in use, generate a new one (retry up to 10 times before returning 503).

#### 3.2.6 Health, Metrics, Monitoring

**Health endpoint:** `GET /health`
Response:
```json
{"status": "ok", "connections": 142, "uptime_seconds": 86400}
```
HTTP 200 if healthy, 503 if degraded.

**Metrics endpoint:** `GET /metrics` (Prometheus format)

Metrics exposed:
```
relix_relay_connections_total{type="agent|mobile"} N
relix_relay_connections_active{type="agent|mobile"} N
relix_relay_messages_total{direction="agent_to_mobile|mobile_to_agent"} N
relix_relay_buffer_size{machine_id="..."} N  # labeled by machine for top offenders
relix_relay_buffer_dropped_total N
relix_relay_pairing_attempts_total N
relix_relay_pairing_success_total N
relix_relay_ws_errors_total N
relix_relay_message_latency_seconds histogram (buckets: 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0)
```

**Logging:** Structured JSON logs to stdout. Fields: `timestamp`, `level`, `msg`, `connection_id`, `user_id` (hashed, not raw), `machine_id`, `latency_ms`

**Docker Compose (for self-hosters):**
```yaml
services:
  relay:
    image: ghcr.io/relixdev/relay:latest
    ports:
      - "8080:8080"
    environment:
      - JWT_SECRET=${JWT_SECRET}
    restart: unless-stopped
  prometheus:
    image: prom/prometheus
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
```

---

### 3.3 Cloud

Relix Cloud is a Go HTTP server backed by PostgreSQL and Redis. It handles authentication, machine registry, billing, and push notifications. It is the only component with persistent storage and the only component that reads user identities (though not session content, which is E2E encrypted).

Base URL: `https://cloud.relix.sh`
All endpoints use JSON. Content-Type: `application/json`.
All authenticated endpoints require `Authorization: Bearer {jwt}` header.

#### 3.3.1 Database Schema

**users:**
```sql
CREATE TABLE users (
    id          VARCHAR(32) PRIMARY KEY,   -- u_xxxx (nanoid)
    email       VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),            -- NULL for OAuth-only users
    github_id   VARCHAR(64) UNIQUE,
    github_login VARCHAR(255),
    display_name VARCHAR(255),
    tier        VARCHAR(16) NOT NULL DEFAULT 'free',
    tier_expires_at TIMESTAMPTZ,          -- NULL = no expiry (free/active sub)
    stripe_customer_id VARCHAR(64),
    stripe_subscription_id VARCHAR(64),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**machines:**
```sql
CREATE TABLE machines (
    id          VARCHAR(32) PRIMARY KEY,   -- m_xxxx
    user_id     VARCHAR(32) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    public_key  VARCHAR(64) NOT NULL,      -- base64 X25519 public key
    status      VARCHAR(16) NOT NULL DEFAULT 'offline',  -- online/offline/revoked
    last_seen_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ON machines(user_id);
```

**sessions (metadata only, content is E2E encrypted):**
```sql
CREATE TABLE sessions (
    id          VARCHAR(32) PRIMARY KEY,   -- s_xxxx
    machine_id  VARCHAR(32) NOT NULL REFERENCES machines(id) ON DELETE CASCADE,
    user_id     VARCHAR(32) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tool        VARCHAR(64) NOT NULL DEFAULT 'claude-code',
    project_path VARCHAR(1024),
    status      VARCHAR(16) NOT NULL DEFAULT 'active',
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at    TIMESTAMPTZ,
    CONSTRAINT fk_machine_user CHECK (true)  -- enforced in app layer
);
CREATE INDEX ON sessions(user_id, started_at DESC);
CREATE INDEX ON sessions(machine_id);
```

**push_tokens:**
```sql
CREATE TABLE push_tokens (
    id          SERIAL PRIMARY KEY,
    user_id     VARCHAR(32) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token       VARCHAR(512) NOT NULL,
    platform    VARCHAR(8) NOT NULL,   -- 'ios' or 'android'
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, token)
);
CREATE INDEX ON push_tokens(user_id);
```

**device_codes (for headless login):**
```sql
CREATE TABLE device_codes (
    device_code VARCHAR(64) PRIMARY KEY,
    user_code   VARCHAR(9) NOT NULL UNIQUE,  -- XXXX-XXXX format
    user_id     VARCHAR(32) REFERENCES users(id),  -- NULL until authorized
    expires_at  TIMESTAMPTZ NOT NULL,
    authorized  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### 3.3.2 Auth Providers

**GitHub OAuth:**

Flow:
1. User clicks "Sign in with GitHub" in mobile app
2. Mobile opens `https://cloud.relix.sh/auth/github?state={random}` in system browser
3. Cloud redirects to GitHub OAuth authorize URL
4. GitHub redirects to `https://cloud.relix.sh/auth/github/callback?code={code}&state={state}`
5. Cloud exchanges code for GitHub access token
6. Cloud calls GitHub API: `GET https://api.github.com/user` → get `id`, `login`, `email`
7. Upsert user record (by `github_id`; if email matches existing user, link accounts)
8. Issue JWT (30-day expiry)
9. Redirect to `relix://auth/callback?token={jwt}` (deep link back to mobile app)

GitHub OAuth app credentials stored as env vars: `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`
Callback URL registered in GitHub: `https://cloud.relix.sh/auth/github/callback`

**Email/Password:**

Registration: `POST /api/v1/auth/register`
```json
{"email": "user@example.com", "password": "s3curepassword"}
```
- Validate email format (RFC 5322 basic check)
- Check uniqueness (case-insensitive)
- Check password min length: 8 chars
- Hash with bcrypt (cost 12)
- Create user record
- Send verification email with link: `https://cloud.relix.sh/auth/verify?token={random-hex-32}`
- Do NOT issue JWT yet — require email verification first
- Response 200: `{"message": "Check your email to verify your account"}`

Email verification: `GET /api/v1/auth/verify?token={token}`
- Look up token in `email_verifications` table (token, user_id, expires_at, used)
- If expired (24h TTL) or used: redirect to `relix://auth/error?reason=expired_token`
- Mark used, set `email_verified_at` on user
- Issue JWT
- Redirect to `relix://auth/callback?token={jwt}`

Login: `POST /api/v1/auth/login`
```json
{"email": "user@example.com", "password": "s3curepassword"}
```
- Find user by email (case-insensitive)
- Verify bcrypt hash
- If email not verified: return 403 `{"error": "email_not_verified", "message": "Check your inbox to verify your email"}`
- Issue JWT
- Response: `{"token": "{jwt}", "user": {"id": "...", "email": "...", "tier": "free"}}`

Failed login: 400 `{"error": "invalid_credentials"}` — same response for "not found" and "wrong password" to prevent enumeration.

**JWT issuance:**
```json
{
  "sub": "u_abc123",
  "email": "user@example.com",
  "tier": "free",
  "iat": 1741689600,
  "exp": 1744281600,   // 30 days
  "jti": "random-uuid" // for future revocation
}
```
Signed HS256 with `JWT_SECRET`. Relay uses same secret for validation.

**Token refresh:** `POST /api/v1/auth/refresh`
- Bearer token required
- If token valid and not expired: issue new JWT (30-day expiry), return `{"token": "{new_jwt}"}`
- If expired by >7 days: return 401, require re-login

#### 3.3.3 Machine Registry

**List machines:** `GET /api/v1/machines`
- Auth required
- Returns all non-revoked machines for the authenticated user
- Response:
```json
{
  "machines": [
    {
      "id": "m_abc123",
      "name": "my-macbook",
      "status": "online",
      "last_seen_at": "2026-03-13T10:00:00Z",
      "created_at": "2026-03-01T00:00:00Z"
    }
  ]
}
```

**Register machine:** `POST /api/v1/machines`
- Called by relay during pairing (relay calls Cloud to register the machine_id)
- Or called directly by agent after pairing
- Body: `{"id": "m_xxx", "name": "my-macbook", "public_key": "{base64}"}`
- Enforce tier machine limits:
  - Free: max 3 machines
  - Plus: max 10 machines
  - Pro/Team: unlimited
- If over limit: 402 `{"error": "machine_limit", "message": "Upgrade to add more machines", "upgrade_url": "https://relix.sh/upgrade"}`
- On success: 201 `{"id": "m_xxx"}`

**Update machine status:** `PATCH /api/v1/machines/{id}`
- Body: `{"status": "online|offline", "last_seen_at": "{timestamp}"}`
- Called by relay on agent connect/disconnect (relay authenticates to Cloud with a service token, not user JWT)
- Returns 200 on success, 404 if machine not found

**Revoke machine:** `DELETE /api/v1/machines/{id}`
- Auth required (user must own machine)
- Sets `status = 'revoked'`
- Notifies relay via internal API to close any active connections for this machine_id
- Returns 204

**Get machine:** `GET /api/v1/machines/{id}`
- Auth required
- Returns full machine details
- 404 if not found or not owned by user

#### 3.3.4 Billing (Stripe Integration)

**Tier definitions (authoritative):**

| Tier | Price (monthly) | Price (annual) | Machines | Concurrent Sessions | History | Recording | Priority Relay | Shared Sessions |
|---|---|---|---|---|---|---|---|---|
| `free` | $0 | $0 | 3 | 2 | 7 days | No | No | No |
| `plus` | $4.99 | $49.90 ($4.16/mo) | 10 | 5 | 30 days | Yes | No | No |
| `pro` | $14.99 | $149.90 ($12.49/mo) | Unlimited | Unlimited | 90 days | Yes | Yes | No |
| `team` | $24.99/user | $249.90/user ($20.82/mo) | Unlimited | Unlimited | 90 days | Yes | Yes | Yes |

Stripe Products and Price IDs must be created in Stripe Dashboard and stored in Cloud config. Map:
- `STRIPE_PRICE_PLUS_MONTHLY` = Stripe price ID for Plus monthly
- `STRIPE_PRICE_PLUS_ANNUAL` = Stripe price ID for Plus annual
- `STRIPE_PRICE_PRO_MONTHLY` = etc.
- `STRIPE_PRICE_PRO_ANNUAL`
- `STRIPE_PRICE_TEAM_MONTHLY`
- `STRIPE_PRICE_TEAM_ANNUAL`

**Create Checkout Session:** `POST /api/v1/billing/checkout`
- Auth required
- Body: `{"tier": "plus|pro|team", "interval": "monthly|annual"}`
- Create or retrieve Stripe Customer for user (store `stripe_customer_id` on user)
- Create Stripe Checkout Session with:
  - `mode: subscription`
  - `price: {STRIPE_PRICE_PLUS_MONTHLY}`
  - `customer: {stripe_customer_id}`
  - `success_url: relix://billing/success?session_id={CHECKOUT_SESSION_ID}`
  - `cancel_url: relix://billing/cancel`
  - `metadata: {"user_id": "u_xxx"}`
- Response: `{"checkout_url": "https://checkout.stripe.com/..."}`
- Mobile opens checkout_url in system browser (not webview — Stripe prohibits webview for checkout)

**Stripe Webhook:** `POST /api/v1/billing/webhook`
- No auth (Stripe-signed). Verify with `STRIPE_WEBHOOK_SECRET` using Stripe's signature verification.
- Handle events:
  - `checkout.session.completed`: extract `metadata.user_id`, update user tier, set `stripe_subscription_id`, send "You're now on {tier}!" push notification
  - `customer.subscription.updated`: handle tier changes, update `tier` and `tier_expires_at`
  - `customer.subscription.deleted`: downgrade to `free`, set `tier_expires_at` to `current_period_end`, keep data during grace period
  - `invoice.payment_failed`: send push notification "Payment failed — update your billing info"; do NOT immediately downgrade; Stripe retries per dunning config
  - `invoice.payment_succeeded`: update `tier_expires_at` to next period end
- All webhook events: return 200 immediately (Stripe requires fast response); process async via goroutine

**Manage Billing:** `GET /api/v1/billing/portal`
- Auth required
- Create Stripe Customer Portal session
- Response: `{"portal_url": "https://billing.stripe.com/..."}`
- Mobile opens in system browser

**Current subscription:** `GET /api/v1/billing/subscription`
- Auth required
- Response:
```json
{
  "tier": "plus",
  "interval": "monthly",
  "current_period_end": "2026-04-13T00:00:00Z",
  "cancel_at_period_end": false,
  "price_monthly": 4.99
}
```

#### 3.3.5 Push Notifications

**Register device token:** `POST /api/v1/push/register`
- Auth required
- Body: `{"token": "{expo-push-token-or-apns-token}", "platform": "ios|android"}`
- Upsert into `push_tokens` table (token unique per user)
- Response 200 OK

**Unregister token:** `DELETE /api/v1/push/token/{token}`
- Auth required
- Remove from `push_tokens` where user_id matches
- Called when user logs out of mobile app

**Send notification (internal, called by relay or by agent via Cloud):**
`POST /internal/push/send` (authenticated with service token, not user JWT)
```json
{
  "user_id": "u_abc123",
  "machine_id": "m_abc123",
  "session_id": "s_def456",
  "type": "approval_needed|session_complete|session_error|payment_failed",
  "title": "Approval Needed",
  "body": "Edit src/auth.ts — my-macbook",
  "data": {
    "machine_id": "m_abc123",
    "session_id": "s_def456",
    "event_type": "tool_use",
    "deep_link": "relix://session/s_def456"
  }
}
```

Cloud implementation:
1. Look up all push tokens for `user_id`
2. For iOS tokens: send via APNs HTTP/2 API using `APNS_KEY_ID`, `APNS_TEAM_ID`, `APNS_PRIVATE_KEY` env vars
3. For Android tokens: send via FCM HTTP v1 API using `FCM_SERVICE_ACCOUNT_JSON` env var
4. Handle token expiry: if APNs returns 410 (BadDeviceToken) or FCM returns `registration-token-not-registered` → delete token from DB
5. Log delivery: success/failure per token, but do NOT block agent operation on push failure

**APNs payload:**
```json
{
  "aps": {
    "alert": {"title": "Approval Needed", "body": "Edit src/auth.ts"},
    "sound": "default",
    "badge": 1,
    "category": "APPROVAL_REQUEST",
    "mutable-content": 1
  },
  "machine_id": "m_abc123",
  "session_id": "s_def456",
  "deep_link": "relix://session/s_def456"
}
```

**APNs notification categories (registered at app launch):**
- `APPROVAL_REQUEST`: actions = [Allow (foreground: false, authRequired: true), Deny (foreground: false, destructive: true)]
- `SESSION_COMPLETE`: actions = [Open Session (foreground: true)]
- `SESSION_ERROR`: actions = [Open Session (foreground: true), Dismiss]

**FCM payload:**
```json
{
  "message": {
    "token": "{fcm-token}",
    "notification": {"title": "Approval Needed", "body": "Edit src/auth.ts"},
    "data": {"machine_id": "m_abc123", "session_id": "s_def456", "type": "approval_needed"},
    "android": {"priority": "high", "notification": {"channel_id": "approvals"}}
  }
}
```

Android notification channels (created at app launch):
- `approvals`: high priority, "Approval Requests"
- `status`: default priority, "Session Status"
- `billing`: default priority, "Billing"

#### 3.3.6 Full REST API Specification

All endpoints prefixed with `/api/v1/`. All require `Authorization: Bearer {jwt}` unless marked `[public]`.

**Auth:**
- `POST /auth/register` [public] — email signup
- `POST /auth/login` [public] — email login
- `GET /auth/github` [public] — GitHub OAuth initiate
- `GET /auth/github/callback` [public] — GitHub OAuth callback
- `POST /auth/refresh` — refresh JWT
- `POST /auth/logout` — invalidate token (for future revocation support; currently a no-op that returns 200)
- `POST /auth/cli/exchange` [public] — exchange CLI auth code for JWT
- `POST /auth/device/start` [public] — start device code flow
- `POST /auth/device/token` [public] — poll for device code token
- `GET /auth/verify` [public] — verify email with token

**Users:**
- `GET /users/me` — get current user profile
- `PATCH /users/me` — update display name
- `DELETE /users/me` — delete account (requires "DELETE" in body for confirmation)

**Machines:**
- `GET /machines` — list machines
- `POST /machines` — register machine
- `GET /machines/{id}` — get machine
- `PATCH /machines/{id}` — update machine name
- `DELETE /machines/{id}` — revoke machine

**Sessions:**
- `GET /sessions` — list sessions (with machine_id filter, status filter, pagination)
- `GET /sessions/{id}` — get session metadata

**Billing:**
- `POST /billing/checkout` — create Stripe checkout session
- `GET /billing/portal` — create Stripe customer portal URL
- `GET /billing/subscription` — current subscription details
- `POST /billing/webhook` [public, Stripe-signed] — Stripe webhook handler

**Push:**
- `POST /push/register` — register push token
- `DELETE /push/token/{token}` — unregister push token

**Team (Team tier only):**
- `GET /team` — get team details
- `POST /team/invite` — invite member by email
- `DELETE /team/members/{user_id}` — remove member
- `GET /team/members` — list members
- `POST /team/transfer` — transfer ownership
- `DELETE /team` — delete team

**Standard error response format:**
```json
{
  "error": "machine_limit",
  "message": "You've reached your plan's machine limit. Upgrade to add more.",
  "code": 402
}
```

HTTP status codes used:
- 200: success
- 201: created
- 204: no content (delete success)
- 400: bad request (validation error)
- 401: unauthorized (missing or invalid JWT)
- 403: forbidden (valid JWT but insufficient permissions)
- 404: not found
- 409: conflict (duplicate email, etc.)
- 402: payment required (tier limit hit)
- 429: rate limited
- 500: internal server error (never expose stack traces)
- 503: service unavailable

---

### 3.4 Mobile App

**Framework:** Expo (React Native, managed workflow)
**Language:** TypeScript
**State management:** zustand
**Navigation:** react-navigation (stack + bottom tabs)
**Crypto:** libsodium-wrappers (WASM port, works in Expo)
**Push:** expo-notifications
**Storage:** expo-secure-store (credentials), @react-native-async-storage/async-storage (non-sensitive data)
**Deep links:** expo-linking

#### 3.4.1 Auth Flow

**App launch sequence:**
1. Check expo-secure-store for JWT (`relix.jwt`)
2. If found and not expired: go to Dashboard
3. If found and within 7 days of expiry: go to Dashboard, refresh token in background
4. If expired or not found: go to Login screen

**Login screen:**
- Two options: "Continue with GitHub" and "Sign in with Email"
- GitHub button: open `https://cloud.relix.sh/auth/github` via `expo-web-browser`; handle deep link callback `relix://auth/callback?token={jwt}`; store JWT in SecureStore; navigate to Dashboard (or Onboarding if new user)
- Email/password form: standard inputs; "Log In" → POST `/auth/login`; on success store JWT; on error show inline error message
- Link: "Don't have an account? Sign up" → Registration screen
- Link: "Forgot password?" → Password reset screen (P1)

**Registration screen:**
- Email input, password input, password confirm input
- Validation: email format, password min 8 chars, passwords match
- POST `/auth/register` → show "Check your email" screen
- Back to login link

**Token storage:**
- JWT stored in `expo-secure-store` under key `relix.jwt`
- User profile cached in AsyncStorage under key `relix.user` (non-sensitive: id, email, tier, display_name)
- Machine list cached in AsyncStorage under key `relix.machines` (non-sensitive metadata)

**Deep link handling:**
- Scheme: `relix://`
- `relix://auth/callback?token={jwt}` — store JWT, navigate to Dashboard
- `relix://auth/error?reason={reason}` — show error screen
- `relix://session/{session_id}` — navigate to Session view for that session
- `relix://billing/success` — show upgrade success modal
- `relix://billing/cancel` — dismiss billing screen

#### 3.4.2 Onboarding

Shown only once: after first login if no machines are paired.

**Step 1 — Welcome:**
- "Welcome to Relix"
- Short value prop: "Control all your AI coding agents from your phone"
- "Get Started" button → Step 2

**Step 2 — Install Agent:**
- Heading: "Install the Agent"
- Two tabs: "Mac" and "Linux"
- Mac tab shows:
  ```
  brew install relixdev/tap/relixctl
  relixctl login
  ```
- Linux tab shows:
  ```
  curl -fsSL relix.sh/install | sh
  relixctl login
  ```
- Both with copy-to-clipboard button
- "I've run the command" button → Step 3
- Connection polling: every 3 seconds, call `GET /api/v1/machines`. If a machine appears, auto-advance to pairing

**Step 3 — Pair Your Machine:**
- "Now pair your machine with this app"
- Pairing code display (6-digit, large monospace font)
- Code generated by: call relay pairing endpoint, display returned code
- Refresh button with 5-minute countdown timer
- Instruction: "On your machine, run: relixctl pair {code}"
- Polling: WebSocket connection to relay pairing endpoint; when `pair_complete` received → Step 4

**Step 4 — Verify (SAS):**
- 4 emoji displayed large
- "Do you see these emoji on your terminal?"
- "Yes, they match" → setup complete → Dashboard
- "No, they don't match" → "Pairing failed. Please try again." → back to Step 3
- Skip button (for users who don't want SAS verification) — shows warning "Skipping verification reduces security"

**Step 5 — Done:**
- "{machine_name} is connected!"
- Brief summary of what to do next: "Start a Claude Code session on your laptop. You'll get push notifications here when approval is needed."
- "Go to Dashboard" button

#### 3.4.3 Dashboard (Home Screen)

**Layout:**
- Bottom tab bar: Home, New Session (+), Settings
- Home = Dashboard

**Machine cards:**
Each machine has a card showing:
- Machine name (left)
- Status badge (right): green dot "Online" / grey dot "Offline" / yellow dot "Busy"
- Active session count: "3 sessions active" or "No active sessions"
- Last seen (if offline): "Last seen 2h ago"
- Tap card → Machine Detail screen

**Approval cards (surfaced at top of dashboard):**
If any pending approvals exist, they appear above the machine cards:
- Card shows: machine name, session project path, tool type (e.g., "File Edit"), brief description
- "Allow" button (green) and "Deny" button (red/grey)
- Tapping "Allow" or "Deny" sends `approval_response` via WebSocket
- On response sent: card animates out, show brief success feedback
- If multiple approvals: show all, most recent first

**Pull-to-refresh:** Refreshes machine list and session counts from relay

**Connection status banner:**
- If WebSocket disconnected: show amber banner at top: "Reconnecting..." with spinner
- On reconnect: banner dismisses

**Empty state:** "No machines connected yet. Tap + to add a machine." with large CTA

#### 3.4.4 Session View — Chat Mode

Accessible via: Machine Detail → tap session, or deep link.

**Header:**
- Back button (←)
- Session title: project path (abbreviated)
- Mode toggle: "Chat" | "Terminal" (segmented control)
- Machine name chip

**Message list:**
- User messages: right-aligned bubble, blue background
- Assistant messages: left-aligned bubble, grey background
- System messages (session start/end): centered, small grey text
- Tool use cards (see below)
- Scroll to bottom on new message; "↓ new messages" button if scrolled up

**Tool use card:**
```
┌─────────────────────────────────┐
│ 🔧 Edit File                    │
│ src/auth.ts                     │
│ +12 lines, -4 lines             │
│ [View Diff]    [Allow]  [Deny]  │
└─────────────────────────────────┘
```
- Pending state: Allow/Deny buttons active
- Approved state: "✓ Allowed" green chip, diff preview still tappable
- Denied state: "✗ Denied" grey chip
- Tool types with icons: Edit File (pencil), Shell Command (terminal), Read File (eye), Create File (document+), Delete File (trash)

**Diff viewer (modal, tapped from "View Diff"):**
- Full-screen modal
- Filename in header
- Unified diff with color coding (red = removed, green = added)
- Line numbers
- "Close" button
- "Allow from here" button (approves without closing)

**Text input:**
- Multiline text input at bottom
- Send button (disabled when empty)
- On send: optimistic UI (bubble appears immediately), awaiting delivery indicator
- Keyboard avoiding behavior (input stays above keyboard)

**Pending indicator:**
- After sending user message: animated "..." typing indicator
- Disappears when first assistant token arrives

#### 3.4.5 Session View — Terminal Mode

Same header as Chat Mode. Toggle switches mode; preserves scroll position where possible.

**Terminal renderer:**
- Monospace font: SF Mono (iOS) / Roboto Mono (Android), fallback to system monospace
- Dark background (#1a1a1a or similar)
- ANSI color support: 16 basic colors + 256-color palette. Map to correct hex values.
- ANSI attributes: bold (font-weight), italic, underline, strikethrough
- No cursor blinking (complexity not worth it)
- Horizontal scroll for lines exceeding screen width (ScrollView horizontal)
- Performance: use FlatList for vertical scroll to handle sessions with thousands of lines

**Approval prompts in terminal mode:**
Rendered as terminal-style prompts:
```
[relix] Allow: Edit src/auth.ts? (y/n) > [Allow] [Deny]
```
Styled to look like a terminal prompt but with tappable buttons.

**Keyboard input:**
- Full-width text input at bottom
- Styled as terminal input (monospace, dark)
- Send on return key

#### 3.4.6 Pairing Flow (Add Machine)

Accessible from: Dashboard "+" button → "Add Machine", or from Onboarding.

Same as Onboarding Steps 2-5, but without the welcome step. Full screen modal, can be dismissed.

**Code refresh:** If timer expires, automatically request a new code from relay and display it.

#### 3.4.7 Push Notifications

**Registration:**
On first app launch (after auth): request permission for push notifications via `expo-notifications`.
- If granted: get Expo push token, also get APNs/FCM token directly; POST both to `/api/v1/push/register`
- If denied: show in-app banner once "Enable notifications to get approval alerts" with "Enable" button → opens iOS/Android settings
- Store registration status in AsyncStorage; do not re-request if already registered

**Foreground notifications:**
When app is in foreground, use `expo-notifications` `addNotificationReceivedListener` to handle incoming pushes; show in-app notification banner at top of screen (Toastable or custom component). Do not play sound if app is open.

**Background notification handling:**
iOS: notification action buttons (Allow/Deny) work via APNs categories even if app is closed.
Android: same via FCM notification actions.

**Deep linking from notification:**
- Tap notification body → open app → navigate to `relix://session/{session_id}`
- Tap "Allow" action button → send approval via background fetch, do NOT open app
- Tap "Deny" action button → send denial, do NOT open app

**Background approval handling:**
When user taps "Allow" or "Deny" from notification without opening app:
1. `expo-notifications` `addNotificationResponseReceivedListener` fires in background
2. Retrieve JWT from SecureStore
3. POST `https://cloud.relix.sh/api/v1/sessions/{session_id}/approve` with `{"decision": "allow|deny"}`
4. Cloud forwards to relay, relay forwards to agent
5. Badge count decremented

**Badge count:**
- Incremented when approval push arrives
- Decremented when approval is actioned (from app or notification)
- Set to 0 on app open (all approvals viewed)
- Managed via `/api/v1/push/badge` (POST `{"count": N}`) and APNs badge field

#### 3.4.8 Settings

**Settings screen structure (bottom tab → Settings):**

**Account section:**
- Email / GitHub username display
- "Display Name" → editable text field (PATCH /users/me)
- "Change Password" (email users only) → P1
- "Sign Out" → logout + clear data

**Billing section:**
- Current tier badge (Free / Plus / Pro / Team)
- Current period / renewal date
- "Upgrade" button (if not on Team) → opens tier selection modal → Stripe Checkout
- "Manage Billing" → Stripe Customer Portal (for paid users)
- "Cancel Subscription" → confirmation dialog → Stripe portal

**Machines section:**
- List of all machines (same as dashboard but with management actions)
- Tap machine → Machine Settings:
  - Rename machine
  - View public key (for verification)
  - Approval timeout setting
  - Auto-approve rules (P1)
  - "Revoke Machine" → confirmation → DELETE /machines/{id}
- "Add Machine" → pairing flow

**Notifications section:**
- "Approval Requests" toggle (default: on)
- "Session Completions" toggle (default: on)
- "Session Errors" toggle (default: on)
- Per-machine notification settings (list of machines with individual toggles)

**Security section:**
- "Require Biometrics" toggle (default: off)
- "App Lock Timeout" → picker: 1 min / 5 min / 15 min / 30 min / Never
- "Clear Local Data" → clears AsyncStorage cache; retains credentials

**Advanced section:**
- "Custom Relay URL" → text field, validates WebSocket URL format
- "Custom Cloud URL" → text field, for self-hosters
- "Debug Mode" → toggle; shows connection status, latency, message counts in debug overlay (default: off)
- App version display

**Team section (Team tier only):**
- Team name display
- Member list with remove buttons (admin only)
- "Invite Member" → email input → POST /team/invite
- "Transfer Ownership" → member picker → confirmation

#### 3.4.9 Security Implementation

**Biometric auth:**
Use `expo-local-authentication`:
- `hasHardwareAsync()` to check availability
- `authenticateAsync()` with `promptMessage: "Unlock Relix"`
- Fallback to device PIN: `fallbackLabel: "Use PIN"`
- If auth fails 3 times: lock app, show "Too many attempts. Use your device PIN."

**App lock timer:**
- Record timestamp in `AppState` `change` event handler when state transitions to `background`
- On `active` state: compute elapsed time; if > timeout AND biometrics enabled → show lock screen
- Lock screen: full-screen blur overlay with biometric prompt
- In-memory data cleared after lock: zustand session store cleared; relay WebSocket closed (will reconnect after unlock)

**Data cleared on lock:**
- `sessionStore.clearSessions()` — clears in-memory session event history
- WebSocket connection closed
- Retained: AsyncStorage cache (non-sensitive machine/session metadata), SecureStore JWT

**Data cleared on logout:**
- SecureStore: JWT deleted
- AsyncStorage: all `relix.*` keys deleted
- zustand stores: reset to initial state
- Push token unregistered: DELETE /push/token/{token}
- WebSocket closed

---

## 4. Non-Functional Requirements

### 4.1 Performance Targets

| Metric | Target | Notes |
|---|---|---|
| Approval push notification delivery | < 5 seconds end-to-end | From agent event to phone notification |
| Relay message latency (p50) | < 50ms | Agent to mobile or mobile to agent |
| Relay message latency (p99) | < 500ms | |
| App cold start to Dashboard | < 2 seconds | On mid-range Android |
| Agent memory (idle) | < 15MB | Daemon with no active sessions |
| Agent memory (active, 5 sessions) | < 50MB | |
| Agent CPU (idle) | < 0.1% | Polling + keepalive only |
| Relay per-connection memory | < 8KB | Excluding message buffer |
| Relay max connections (single instance) | 10,000 | Go goroutine per connection |
| Cloud API response time (p50) | < 100ms | All endpoints except Stripe calls |
| Dashboard load (machine list) | < 1 second | Cached locally, fresh from WebSocket |

### 4.2 Security Requirements

**Encryption in transit:**
- All WebSocket connections over WSS (TLS 1.2+)
- All HTTP API calls over HTTPS (TLS 1.2+)
- HSTS headers on all Cloud endpoints
- Certificate pinning in mobile app (optional, P2 — adds complexity, breaks self-hosting)

**Encryption at rest:**
- PostgreSQL: enable encryption at rest via Fly.io volume encryption
- Redis: no sensitive data stored (only ephemeral sessions); use Fly.io private network (not public)
- Mobile: Expo SecureStore uses iOS Keychain / Android Keystore (hardware-backed)
- Agent credentials: chmod 0600, never world-readable

**E2E encryption:**
- All `payload` fields in session events encrypted with NaCl box (X25519 + XSalsa20-Poly1305)
- Nonces: random 24 bytes per message (never reused)
- Each mobile device has its own keypair; multiple devices paired to same machine each have separate encrypted channels (not shared key)
- No key escrow; Relix cannot decrypt session content. This is a product feature, not a bug.

**Key rotation:**
- Agent initiates rotation every 30 days
- New X25519 keypair generated on machine
- Old key accepted for 48-hour grace period
- If offline >30 days: require re-pairing
- Emergency revocation: user can revoke machine from app → relay drops connection → agent gets 403 on next connect

**Authentication security:**
- JWT HS256 with 256-bit secret minimum
- JWT expiry: 30 days; refresh window: last 7 days before expiry
- Bcrypt cost 12 for password hashing
- Rate limiting on auth endpoints:
  - `POST /auth/login`: 5 attempts per email per 15 minutes (IP + email combined limit)
  - `POST /auth/register`: 3 accounts per IP per hour
  - `GET /auth/github`: 10 per IP per hour

**Relay security:**
- No user data stored (stateless buffer in memory)
- User IDs in logs are hashed (SHA256 first 8 chars) for pseudonymization
- Relay cannot decrypt payloads
- Auth required within 5 seconds of WebSocket open; else close connection
- Message size limit: 1MB per message (to prevent memory exhaustion)
- Connection limit per user: 10 simultaneous (to prevent abuse)

**Supply chain:**
- Go modules: use `go.sum` with checksums
- Expo: `npm audit` in CI
- No telemetry/analytics SDKs that phone home with user data
- Dependency updates: Dependabot configured for all repos

### 4.3 Scalability

**Relay (stateless horizontal scaling):**
- Each relay instance is independent; no shared state between instances
- WebSocket sticky sessions required at load balancer (Fly.io: use session affinity via `fly-prefer-region` header)
- Current design: all state in-memory per instance. Scale by running multiple instances with sticky LB.
- Future: if needed, relay state can be moved to Redis for true stateless scaling
- Target: each instance handles 10K connections; scale to 3 instances for 30K connections at launch

**Cloud (stateless application, stateful database):**
- Cloud API is stateless; scale horizontally behind load balancer
- PostgreSQL: single primary at launch (Fly.io Postgres); add read replicas when read load requires
- Redis: single instance for ephemeral data (device code cache, rate limiting)
- Target: 100 req/sec per instance; scale instances as needed

**Mobile:**
- WebSocket reconnects gracefully after relay restart
- Exponential backoff: 1s, 2s, 4s, 8s, 16s, 32s, 60s max
- Offline mode: cached dashboard data shown; approval buttons disabled; "No connection" banner

### 4.4 Availability Targets

| Component | Target Uptime | Notes |
|---|---|---|
| Relay | 99.5% | WebSocket hub; downtime = users can't receive events |
| Cloud | 99.9% | Auth + billing; downtime = can't login or process payments |
| Push (APNs/FCM) | Best effort | Third-party dependency; degraded gracefully |

**Recovery behavior:**
- Relay restart: agents and mobile reconnect automatically via backoff; buffered messages lost on restart (acceptable per design)
- Cloud restart: agents have cached JWTs; mobile has cached data; no immediate user impact for <5 minute outages
- DB failure: Cloud returns 503; agents continue operating with cached auth; mobile shows "Unable to sync" banner

**Fly.io deployment:**
- Relay: 2 instances minimum (rolling deploy, one always up)
- Cloud: 1 instance minimum (acceptable given stateless + fast restart)
- Both in `ord` (Chicago) region at launch; add regions based on user geography

### 4.5 Compliance

**Data privacy:**
- Session content is E2E encrypted; Relix cannot read it. This is the primary privacy guarantee.
- Metadata collected: email, machine names, session start/end times, session counts. Documented in Privacy Policy.
- No third-party analytics SDKs at launch. Server-side analytics only (aggregate metrics from Cloud logs).
- User data deletion: `DELETE /users/me` removes all user data from DB within 30 days. Push tokens deleted immediately.

**GDPR (European users):**
- Privacy Policy must include: data controller identity (Relix), data processed, legal basis (contract), retention periods, user rights (access, erasure, portability)
- Data stored in US (Fly.io ord region) at launch. Add EU region if significant EU user base develops.
- `GET /users/me` returns all user metadata (data access right)
- `DELETE /users/me` implements right to erasure
- No data sold to third parties. Document in Privacy Policy.

**App Store compliance:**
- iOS: no use of private APIs; privacy manifest required (Expo handles); data usage disclosure in App Store Connect
- Android: Google Play data safety form required; document data collection accurately
- Both: no subscription dark patterns; clear cancellation path in-app (Apple/Google requirement)

**Payment compliance:**
- Stripe handles PCI compliance for payment card data; Relix never sees card numbers
- Apple in-app purchase: NOT used. Stripe Checkout is used (web browser flow). This is allowed for non-digital-goods subscriptions per App Store guidelines. However, Apple may require IAP for in-app upgrade flows — monitor Apple policy. If forced to use IAP: Apple takes 15-30% cut, add to pricing model.

---

## 5. Monetization Requirements

### 5.1 Tier Definitions (Authoritative)

These are the enforceable tier limits. All enforcement happens in Cloud API (`relixctl` respects them but Cloud is the authority).

**Free:**
- Max machines: 3 (enforced at `POST /machines`)
- Max concurrent sessions: 2 (enforced by relay: when mobile requests session list, Cloud returns which sessions are viewable; relay limits event forwarding to 2 sessions per connection)
- Session history: 7 days (enforced at `GET /sessions`: filter `started_at > now-7days`)
- Session recording: disabled (Cloud rejects recording uploads from Free users)
- Priority relay: no (Free connections use standard queue on relay; Pro/Team flagged in JWT and get priority in relay message queue — future implementation)
- Shared sessions: no

**Plus ($4.99/mo or $49.90/yr):**
- Max machines: 10
- Max concurrent sessions: 5
- Session history: 30 days
- Session recording: enabled (recordings stored as E2E encrypted blobs in Fly.io storage; 5GB storage limit)
- Priority relay: no
- Shared sessions: no

**Pro ($14.99/mo or $149.90/yr):**
- Max machines: unlimited (no enforcement check)
- Max concurrent sessions: unlimited
- Session history: 90 days
- Session recording: enabled (20GB storage limit)
- Priority relay: yes (JWT claim `"priority": true` recognized by relay)
- Shared sessions: no

**Team ($24.99/user/mo or $249.90/user/yr):**
- Per-seat pricing: billed monthly based on active member count (synced with Stripe metered billing or manual seat count update)
- All Pro limits, plus:
  - Shared sessions: team members can view each other's sessions (opt-in per machine)
  - Admin dashboard (web, P2): manage machines, view team session activity
  - SSO (P3): SAML/OIDC integration
  - Audit trail: all approval decisions logged with timestamp, user, decision (stored 90 days)
  - Team-wide notification policies (P2)
- Minimum: 2 seats

**Tier enforcement summary:**
- Machine limit: Cloud `POST /machines` checks current count
- Session limit: Relay checks sessions-in-progress count for user (tracked in relay in-memory store, synced from Cloud on connect)
- History limit: Cloud `GET /sessions` query filters by date based on tier
- Recording: Cloud `POST /sessions/{id}/recording` checks tier

### 5.2 Stripe Integration Requirements

**Required Stripe config (environment variables):**
```
STRIPE_PUBLISHABLE_KEY=pk_live_...
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRICE_PLUS_MONTHLY=price_...
STRIPE_PRICE_PLUS_ANNUAL=price_...
STRIPE_PRICE_PRO_MONTHLY=price_...
STRIPE_PRICE_PRO_ANNUAL=price_...
STRIPE_PRICE_TEAM_MONTHLY=price_...
STRIPE_PRICE_TEAM_ANNUAL=price_...
```

**Stripe Product setup (one-time in Stripe Dashboard):**
Create 3 products: "Relix Plus", "Relix Pro", "Relix Team"
For each: create monthly and annual prices
Team monthly price: use per-unit pricing with `billing_scheme: per_unit`, quantity = seat count

**Subscription lifecycle:**
1. User initiates upgrade → Cloud creates Stripe Customer (or retrieves existing), creates Checkout Session
2. User completes Checkout → Stripe fires `checkout.session.completed` webhook
3. Cloud receives webhook, updates `users.tier`, `users.stripe_subscription_id`, `users.tier_expires_at`
4. Cloud sends push notification "You're now on {Tier}! 🎉"
5. Monthly: Stripe fires `invoice.payment_succeeded` → Cloud updates `tier_expires_at`
6. User cancels → Stripe fires `customer.subscription.deleted` (at period end) → Cloud sets `tier = 'free'`
7. Payment fails → Stripe retries (smart dunning); Cloud receives `invoice.payment_failed` → sends push "Update your payment method"

**Annual billing:**
- Annual plan is 2 months free: Plus $49.90/yr (vs $59.88), Pro $149.90/yr (vs $179.88)
- Stripe: create annual prices at discounted rates
- UI: show "Save 2 months" label on annual option in upgrade modal

**Team seat billing:**
- At launch: use fixed-quantity subscriptions (admin sets seat count manually)
- Stripe Checkout quantity = number of seats
- If team grows beyond purchased seats: Cloud rejects new member invites until admin upgrades seats (PUT /team with new seat_count, which updates Stripe subscription quantity via API)
- Proration: use Stripe's default proration behavior (charge immediately for added seats)

**Grace period (subscription lapse):**
- On `customer.subscription.deleted`: do NOT immediately remove data or downgrade limits
- Set `tier_expires_at = current_period_end` (already paid for current period)
- On first API call after `tier_expires_at`: apply Free tier limits
- Data beyond Free tier limits (old history): retained for 7 additional days, then deleted by background job
- Push notification sent 3 days before and on the day limits apply: "Your {tier} access expires in 3 days"

### 5.3 Free Tier Conversion Strategy

**Conversion triggers (where to show upgrade prompts):**
1. When user tries to add 4th machine → 402 error → "Upgrade to Plus for up to 10 machines" in-app card
2. When user tries to start 3rd concurrent session → inline "Upgrade for more sessions" prompt
3. In Session view: show "Recording available on Plus+" banner at bottom of ended sessions (P1)
4. After 7 days of use: show "Your history is limited to 7 days. Upgrade to Plus to see 30 days." (max once per week, dismissible)
5. Dashboard: for users with 2-3 machines, show "Add more machines on Plus" small CTA (non-intrusive, once)

**What to never do:**
- Do NOT show upgrade prompts on first open
- Do NOT nag users who have dismissed a prompt (max 1 prompt per 7 days per category)
- Do NOT disable core functionality on Free tier — all core features (push notifications, chat mode, terminal mode) are available on Free
- The free tier must be genuinely useful on its own

**Conversion messaging:**
- Plus: "A coffee a month. 10 machines, 5 sessions, 30 days history."
- Pro: "For the serious developer. Unlimited everything."
- Team: "Give your team visibility into every agent." (target: EMs, not devs)

### 5.4 Revenue Targets

**Year 1 targets (not enforced by OpenClaw, but inform prioritization):**
- Month 1 (launch): 100 free users, 10 paid
- Month 3: 1,000 free users, 100 paid (~$800 MRR)
- Month 6: 5,000 free users, 500 paid (~$5,000 MRR)
- Month 12: 20,000 free users, 2,000 paid ($20,000+ MRR)

**Primary acquisition channel:** Developer word-of-mouth, GitHub, Hacker News, r/programming
**Secondary:** The open source relay generates GitHub stars which funnel to paid mobile app

---

## 6. Success Metrics

### 6.1 Launch KPIs

| Metric | Target at Launch | Target at 30 days |
|---|---|---|
| App Store rating | — | ≥ 4.5 |
| Free signups | — | 500 |
| Paid conversions | — | 50 |
| Daily Active Users | — | 100 |
| Push notification opt-in rate | — | ≥ 70% |
| Approval response rate (from notification) | — | ≥ 60% |
| Crash-free rate (mobile) | — | ≥ 99.5% |
| Relay uptime | 99.5% | 99.5% |

### 6.2 Analytics Events to Track

All events tracked server-side (Cloud logs → aggregation). No client-side analytics SDKs.

**User lifecycle:**
- `user.signup` — `{method: "github|email"}`
- `user.login` — `{method: "github|email"}`
- `user.logout`
- `user.deleted`

**Agent:**
- `agent.installed` — `{platform: "macos|linux", arch: "amd64|arm64", method: "brew|curl"}`
- `agent.login`
- `agent.paired` — `{sas_verified: true|false}`
- `agent.daemon_started`
- `agent.session_discovered` — `{tool: "claude-code"}`

**Sessions:**
- `session.started` — `{tool: "claude-code", source: "terminal|mobile"}`
- `session.event_forwarded` — `{kind: "assistant_message|tool_use|..."}`
- `session.approval_requested`
- `session.approval_responded` — `{decision: "allow|deny", source: "app|notification|timeout"}`
- `session.ended` — `{duration_seconds: N, messages: N}`

**Mobile engagement:**
- `screen.viewed` — `{screen_name: "Dashboard|SessionView|Settings|..."}`
- `notification.received` — `{type: "approval_needed|session_complete|..."}`
- `notification.actioned` — `{action: "allow|deny|open", in_app: false}`

**Monetization:**
- `billing.upgrade_prompted` — `{trigger: "machine_limit|session_limit|history_cta"}`
- `billing.checkout_opened` — `{tier: "plus|pro|team", interval: "monthly|annual"}`
- `billing.checkout_completed` — `{tier: "plus|pro|team"}`
- `billing.cancelled`
- `billing.payment_failed`

### 6.3 Funnel Definitions

**Activation funnel:**
1. App downloaded → 2. Account created → 3. Agent installed → 4. Machine paired → 5. First session viewed → 6. First approval actioned

Target: ≥ 40% of signups reach step 5 within 7 days.

**Conversion funnel:**
1. Free user → 2. Hits tier limit → 3. Sees upgrade prompt → 4. Opens Checkout → 5. Completes payment

Target: ≥ 20% conversion from "hits tier limit" to "completes payment".

**Retention metric:**
- D7 retention (returns to app within 7 days): target ≥ 50%
- D30 retention: target ≥ 30%

---

## 7. Open Questions & Risks

### 7.1 Technical Risks

**Risk 1: Claude Code stream-json protocol changes (HIGH)**
- Anthropic does not publish a stability guarantee for `--input-format stream-json`
- Mitigation: The format is used by Anthropic's own Agent SDK; likely to remain stable. Monitor Claude Code changelog on release. Add integration tests that run against the real Claude Code binary on each Relix release. Have a fallback plan: tail the .jsonl files directly (read-only mode) if the stdin protocol breaks.
- Decision: If stream-json breaks without warning and we have no fallback, relixctl falls back to .jsonl tailing (read-only) and shows "Limited mode: upgrade Claude Code to restore full control"

**Risk 2: Claude Code session file path encoding (MEDIUM)**
- The exact encoding of project paths in `~/.claude/projects/` directory names is not documented in any public spec
- Mitigation: Test empirically against real Claude Code sessions on multiple platforms before launch. Document the encoding in a comment with a test case. If encoding changes, session discovery breaks silently (no crash, just sessions not shown).
- Decision: Treat the directory name encoding as an implementation detail we reverse-engineered; add a test that creates a session with a known path and verifies discovery.

**Risk 3: NaCl box key management on mobile (MEDIUM)**
- libsodium-wrappers (WASM) in Expo adds ~1.5MB to bundle; WASM initialization is async
- Private key stored in expo-secure-store (limited to ~2KB on some Android devices)
- Mitigation: X25519 private key is 32 bytes (64 bytes base64) — well within SecureStore limits. WASM init on app launch (~200ms) — accept the cost. Test on low-end Android (Samsung Galaxy A series).

**Risk 4: Apple Review rejection (MEDIUM)**
- Apple may reject the app for using Stripe Checkout (web flow) instead of Apple IAP for subscription upgrades
- Rule: Apple requires IAP for in-app purchases of digital goods. However, subscriptions that are initiated on web (outside the app) are allowed. The app will only show a "Manage Billing" link that opens Stripe in Safari — this pattern is used by many B2B SaaS apps.
- Mitigation: Structure the app to not have an in-app "Buy Now" button. Instead, show a "Visit relix.sh to upgrade" CTA, or open the Stripe Checkout in the system browser (not a webview). If Apple still rejects: implement Apple IAP as a parallel path (with higher prices to cover the 15-30% cut).

**Risk 5: Background notification actions on iOS (MEDIUM)**
- iOS background fetch for notification action handling (Allow/Deny from notification) requires careful entitlement configuration
- expo-notifications handles most of this, but testing on physical iOS devices before App Store submission is required
- Mitigation: Test on iPhone with TestFlight early. Background app refresh must be enabled. Document the required Expo config.

**Risk 6: Claude Code hooks may be changed or removed (LOW)**
- The hooks system is documented but could change
- Mitigation: The PreToolUse hook is the core of our approval system. If it's removed, we lose in-process approval and must fall back to a less reliable pattern (detecting stdin wait state). The hook is used widely by the community — risk of removal is low.

**Risk 7: relixctl daemon stability on macOS (LOW)**
- launchd can kill long-running daemons unpredictably; macOS power management may suspend network connections
- Mitigation: launchd plist uses `KeepAlive: true` so daemon is restarted on crash. Network change detection using SCNetworkReachability callbacks to force reconnect after network changes.

### 7.2 Business Risks

**Risk 8: Anthropic improves Remote Control to match Relix (HIGH)**
- Anthropic adding multi-machine dashboard, push notifications, and E2E encryption would significantly undermine Relix's value prop
- Mitigation: Focus on multi-tool value prop (Aider, Cline) immediately after launch. Anthropic will not support competing tools. The multi-tool differentiation is durable.
- If Anthropic adds push notifications: Relix's advantage narrows to multi-tool + privacy + self-hosting. These are still real differentiators for power users and enterprises.

**Risk 9: Small total addressable market (MEDIUM)**
- The market for "mobile control of AI coding agents" may be smaller than hoped
- Signal: Anthropic Remote Control exists and has users (exact numbers unknown). Demand is real.
- Mitigation: Free tier generates word-of-mouth with no marginal cost. Even 1,000 paid users at $10 ARPU = $10K MRR — a viable solo product.

**Risk 10: Competitor clones the product (MEDIUM)**
- Any developer could build a similar product. Anthropic could build it natively.
- Mitigation: Speed to market, multi-tool support, and the open source relay/agent create switching costs. The mobile app experience is a moat.

### 7.3 Third-Party Dependencies

**Anthropic / Claude Code:**
- Claude Code binary is the P0 adapter target. No SLA or backward compatibility guarantee from Anthropic for stream-json CLI format.
- Risk: Protocol change without notice. Mitigation: integration tests, fallback to read-only mode.
- Claude Code license: check whether CLI automation via stdin/stdout is permitted under Anthropic's terms. Assumption: yes (it's a CLI tool). Verify before launch.

**Apple (APNs + App Store):**
- APNs: 99.9% uptime historically. P12 certificate or token-based auth (prefer token-based; no expiry issues)
- App Store review: 1-3 day review time. Submit TestFlight first for faster iteration.
- iOS push in killed app: iOS allows push notifications and background fetch to work when app is killed. Confirmed.

**Google (FCM + Play Store):**
- FCM: highly reliable. Use FCM HTTP v1 API (legacy API deprecated 2024)
- Play Store: faster review (hours vs days) than App Store

**Stripe:**
- Payment reliability: 99.99% uptime SLA
- Webhook delivery: at-least-once; implement idempotency keys (`stripe_idempotency_key` = `{webhook_event_id}`) to handle duplicate webhook delivery
- Test mode: use `stripe listen --forward-to localhost:8080/api/v1/billing/webhook` for local development

**Fly.io:**
- Infrastructure for relay + cloud. 99.9% uptime SLA.
- Private networking: relay and cloud communicate via Fly.io private network; Redis not exposed publicly
- Volumes: PostgreSQL data on Fly.io volumes. Set up daily snapshots.

### 7.4 What Happens if Claude Code Changes Its Protocol

**Scenario A: `--input-format stream-json` flag removed or changed**
- Detection: relixctl will fail to parse stdout events; log parse errors; send `protocol_error` event to mobile
- Mobile shows: "Claude Code protocol mismatch. Update relixctl: brew upgrade relixctl"
- Recovery: release updated relixctl with new protocol within 24 hours of detecting the change
- Fallback mode: tail .jsonl files directly; read-only bridge mode

**Scenario B: `~/.claude/projects/` directory structure changes**
- Detection: session discovery returns 0 sessions; agent logs "no sessions found"
- Recovery: update discovery logic in relixctl; release update
- User impact: sessions not shown in app until relixctl is updated; Claude Code itself continues to work

**Scenario C: Hooks system changed or removed**
- Detection: PreToolUse hook not called; approval requests never arrive at agent
- Mobile shows: "Approvals unavailable. Update relixctl."
- Recovery: find alternative approval interception mechanism (inspect Claude Code source, check for new hook types, or use stdout event monitoring to detect Claude Code pausing for input)

**Protocol monitoring strategy:**
- Subscribe to Claude Code GitHub releases (ghcr.io/anthropics/claude-code or npm package)
- Add a `claude_code_version` field to session metadata; log when version changes
- Integration test suite that spawns real Claude Code binary; run on CI weekly

---

## Appendix A: Configuration File Locations

| File | Location | Permissions | Purpose |
|---|---|---|---|
| Credentials | `~/.config/relixctl/credentials.json` | 0600 | JWT + user info |
| Keys | `~/.config/relixctl/keys.json` | 0600 | X25519 keypair |
| Peers | `~/.config/relixctl/peers.json` | 0600 | Mobile public keys |
| Config | `~/.config/relixctl/config.json` | 0644 | User configuration |
| Daemon PID | `~/.config/relixctl/daemon.pid` | 0644 | Daemon process ID |
| Daemon socket | `~/.config/relixctl/daemon.sock` | 0600 | Hook handler IPC |
| Daemon log | `~/.config/relixctl/daemon.log` | 0644 | Daemon logs (rotated) |
| launchd plist | `~/Library/LaunchAgents/sh.relix.relixctl.plist` | 0644 | macOS auto-start |
| systemd service | `~/.config/systemd/user/relixctl.service` | 0644 | Linux auto-start |

## Appendix B: Environment Variables (Cloud)

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `REDIS_URL` | Yes | Redis connection string |
| `JWT_SECRET` | Yes | Shared with relay for JWT validation |
| `GITHUB_CLIENT_ID` | Yes | GitHub OAuth app client ID |
| `GITHUB_CLIENT_SECRET` | Yes | GitHub OAuth app client secret |
| `STRIPE_SECRET_KEY` | Yes | Stripe secret key |
| `STRIPE_PUBLISHABLE_KEY` | Yes | Stripe publishable key |
| `STRIPE_WEBHOOK_SECRET` | Yes | Stripe webhook signing secret |
| `STRIPE_PRICE_PLUS_MONTHLY` | Yes | Stripe price ID |
| `STRIPE_PRICE_PLUS_ANNUAL` | Yes | Stripe price ID |
| `STRIPE_PRICE_PRO_MONTHLY` | Yes | Stripe price ID |
| `STRIPE_PRICE_PRO_ANNUAL` | Yes | Stripe price ID |
| `STRIPE_PRICE_TEAM_MONTHLY` | Yes | Stripe price ID |
| `STRIPE_PRICE_TEAM_ANNUAL` | Yes | Stripe price ID |
| `APNS_KEY_ID` | Yes | APNs token key ID |
| `APNS_TEAM_ID` | Yes | Apple Developer Team ID |
| `APNS_PRIVATE_KEY` | Yes | APNs private key (PEM, base64 encoded) |
| `APNS_BUNDLE_ID` | Yes | iOS app bundle ID (sh.relix.app) |
| `FCM_SERVICE_ACCOUNT_JSON` | Yes | FCM service account JSON (base64 encoded) |
| `SERVICE_TOKEN` | Yes | Internal token for relay→cloud calls |
| `PORT` | No | HTTP port (default 8080) |
| `LOG_LEVEL` | No | `debug|info|warn|error` (default info) |
| `BASE_URL` | Yes | Public URL (https://cloud.relix.sh) |

## Appendix C: Environment Variables (Relay)

| Variable | Required | Description |
|---|---|---|
| `JWT_SECRET` | Yes | Must match Cloud's JWT_SECRET |
| `CLOUD_URL` | Yes | Cloud API URL for machine status updates |
| `SERVICE_TOKEN` | Yes | Token for authenticating relay→cloud calls |
| `PORT` | No | HTTP port (default 8080) |
| `LOG_LEVEL` | No | default info |
| `MAX_CONNECTIONS_PER_USER` | No | default 10 |
| `BUFFER_MAX_MESSAGES` | No | default 1000 |
| `BUFFER_MAX_BYTES` | No | default 10485760 (10MB) |
| `BUFFER_TTL_SECONDS` | No | default 86400 (24h) |

## Appendix D: Mobile App Constants

| Constant | Value | Notes |
|---|---|---|
| Bundle ID (iOS) | `sh.relix.app` | |
| Package name (Android) | `sh.relix.app` | |
| Deep link scheme | `relix://` | |
| App lock default timeout | 5 minutes | |
| Session history poll interval | 30 seconds | |
| WebSocket ping interval | 30 seconds | |
| WebSocket pong timeout | 10 seconds | |
| Reconnect max backoff | 60 seconds | |
| Approval auto-deny timeout | 30 minutes | Configurable per machine |
| Push token refresh interval | 24 hours | Re-register on app open if >24h |
| Min password length | 8 chars | |
| Biometric max failures | 3 | Falls back to PIN |

## Appendix E: SAS Emoji Pool (64 emoji)

The SAS (Short Authentication String) uses exactly these 64 visually distinct emoji, indexed 0-63:

🐶 🐱 🐭 🐹 🐰 🦊 🐻 🐼 🐨 🐯 🦁 🐮 🐷 🐸 🐵 🐔 🐧 🐦 🦆 🦅 🦉 🦇 🐝 🪲 🐛 🦋 🐌 🐞 🐜 🦟 🦗 🕷️ 🐢 🐍 🦎 🦖 🐙 🦑 🦀 🐡 🐬 🐳 🦈 🌊 🔥 ⚡ 🌈 ❄️ 🌸 🍎 🍕 🚀 🎸 🔑 💎 🎯 🏆 ⚽ 🎭 🎪 🎨 🎬 🔮 🗝️

SAS derivation:
```
sas_bytes = HKDF-SHA256(
  salt = "relix-sas-v1",
  ikm = agent_public_key || mobile_public_key,
  info = "",
  length = 4
)
emoji[0] = sas_bytes[0] % 64
emoji[1] = sas_bytes[1] % 64
emoji[2] = sas_bytes[2] % 64
emoji[3] = sas_bytes[3] % 64
```

Both agent and mobile compute this independently. If the same emoji appear on both devices, the public keys are identical and no MITM occurred.
