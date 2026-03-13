# Relix — Executive Handoff

**Read this first.** This document tells you exactly where the project stands, what works, what is stubbed, what to do next, and what not to touch. If you read only one document, read this one.

---

## Table of Contents

1. [Current State of the Project](#1-current-state)
2. [Critical Path to First Paying Customer](#2-critical-path)
3. [Accounts and Services to Create](#3-accounts-and-services)
4. [Secrets and Credentials to Generate](#4-secrets-and-credentials)
5. [Known Issues and Technical Debt](#5-known-issues)
6. [Architecture Decisions and Rationale](#6-architecture-decisions)
7. [What NOT to Change](#7-what-not-to-change)
8. [What SHOULD Be Changed](#8-what-should-be-changed)
9. [Estimated Timeline to Beta](#9-timeline-to-beta)
10. [Estimated Timeline to Public Launch](#10-timeline-to-public-launch)

---

## 1. Current State of the Project

### What Is Built and Works

The following has been implemented and has tests:

**Protocol (`protocol/`)**
- Wire format: JSON envelopes over WebSocket, fully specified
- Crypto: X25519 key generation, NaCl box encrypt/decrypt, SealPayload/OpenPayload
- Types: Envelope, Payload, Session, Event, UserInput, CopilotAdapter interface
- Status: Complete. Do not modify.

**Relay (`relay/`)**
- WebSocket hub with connection registry
- Agent↔mobile message routing
- Per-machine offline message buffering (1000 messages, 10MB, 24h TTL)
- Drain buffered messages on mobile reconnect, ordered by seq
- Pairing: 6-digit code generation, per-IP rate limiting, 5-minute TTL
- Prometheus metrics
- JWT authentication on connect
- Docker image (distroless, multi-stage)
- 50+ tests including integration tests
- Status: **Production-ready**. Deploy this as-is.

**Agent (`relixctl/`)**
- Claude Code adapter: subprocess management, stream-json parsing, event mapping
- Session discovery: scans `~/.claude/projects/` for session `.jsonl` files
- Relay client: WebSocket connect, reconnect with exponential backoff
- E2E encryption bridge
- Daemon supervisor with launchd (macOS) and systemd (Linux) integration
- Config management: `~/.relixctl/config.json`
- Keystore: `~/.relixctl/keys/`
- Commands: config, status, sessions, start, stop, uninstall, daemon-run
- Cobra CLI structure
- 90+ tests
- **Stubs**: `login` and `pair` commands print "not yet implemented" — they are registered but not wired to real auth flows
- Status: Core infrastructure complete. Auth commands need real implementation.

**Cloud (`cloud/`)**
- HTTP server with all routes registered
- Auth: GitHub OAuth exchange, email/bcrypt register/login, JWT issue/validate, middleware
- Machine registry: register, list, delete with per-tier limit enforcement
- Billing: plan definitions, tier limits, checkout handler
- Push: register and send handlers
- **Stubs**: Stripe (returns fake URLs), APNs (logs to stdout), FCM (does not exist — only APNs stub), PostgreSQL (in-memory store), Redis (not used)
- 15+ tests
- Status: API shape is correct. All business logic stubs need real implementations.

**Mobile (`mobile/`)**
- React Native (Expo) project bootstrapped
- tsconfig with strict mode
- Assets: icon, splash, android adaptive icons
- Status: Scaffolded. All screens, navigation, and API calls need to be built or completed.

### What Is Not Built

1. Mobile app screens: dashboard, session view, pairing, settings, onboarding — these need to be built out (TypeScript/React Native)
2. `relixctl login` — stub, needs real GitHub OAuth browser flow + device code flow
3. `relixctl pair` — stub, needs to call the relay's pairing API and exchange public keys
4. PostgreSQL user store — in-memory only, does not survive restarts
5. APNs push — stub that logs; real `apns2` HTTP/2 client not implemented
6. FCM push — not implemented at all
7. Stripe checkout — stub returns fake URLs
8. Stripe webhook handler — not implemented (no route registered)
9. Stripe billing portal — not implemented
10. User tier upgrade path — no webhook processing to update user tier after payment
11. PATCH /machines/:id (rename) — not implemented
12. POST /billing/portal — not implemented
13. Aider adapter — not implemented
14. Cline adapter — not implemented
15. Homebrew tap — not published
16. Install script (`relix.sh/install`) — not created
17. Landing page (`relix.sh`) — not created
18. CI/CD GitHub Actions — not configured

---

## 2. Critical Path to First Paying Customer

Execute these steps in order. Each step is atomic — complete it fully before moving to the next.

### Step 1: Push the Code to GitHub

```bash
# Create GitHub org: github.com/relixdev
# Create repos: protocol, relay, relixctl (public, MIT), cloud, mobile (private)

git remote add origin git@github.com:relixdev/relix.git
git push -u origin main

# Or if using separate repos per component:
# cd relay && git remote add origin git@github.com:relixdev/relay.git && git push -u origin main
# cd relixctl && git remote add origin git@github.com:relixdev/relixctl.git && git push -u origin main
```

### Step 2: Register relix.sh Domain

- Register at Cloudflare (https://www.cloudflare.com/products/registrar/) — cheapest and best DNS
- Set nameservers to Cloudflare if registering elsewhere
- Do not configure DNS records yet — wait until Fly.io is set up

### Step 3: Deploy Relay to Fly.io

See `docs/DEPLOYMENT.md` section 5. This takes ~30 minutes.

Done when: `curl https://relay.relix.sh/` returns a WebSocket upgrade response (101).

### Step 4: Deploy Cloud to Fly.io

See `docs/DEPLOYMENT.md` section 6. This takes ~30 minutes.

Done when: `curl https://api.relix.sh/health` returns `{"status":"ok"}`.

### Step 5: Replace the In-Memory User Store with PostgreSQL

This is the most critical stub to replace. The in-memory store loses all users on every restart.

1. Provision Fly Postgres (`fly postgres create --name relix-postgres --region iad`)
2. Run the migrations from DEPLOYMENT.md section 3
3. Implement `PostgresStore` in `cloud/internal/user/` that satisfies the `user.Store` interface
4. Wire it in `cloud/cmd/cloud/main.go` instead of `user.NewMemoryStore()`
5. Deploy the updated cloud

The `user.Store` interface is defined in `cloud/internal/user/`. Implement: `CreateUser`, `GetUserByID`, `GetUserByEmail`, `GetUserByGitHubID`. All four methods are called by existing handlers.

Done when: Creating a user via `POST /auth/email/register`, restarting the cloud service, then calling `POST /auth/email/login` with the same credentials succeeds.

### Step 6: Implement relixctl login and pair

The `login` and `pair` commands are stubs (`cmd/root.go` registers them with `stubCmd()`). Implement them:

**login:**
- `relixctl login` — open `https://api.relix.sh/auth/github` in the browser, receive the code on callback, exchange for JWT, store in `~/.relixctl/config.json`
- `relixctl login --code` — device code flow: display a URL and code, poll for completion

Reference implementation: `relixctl/internal/auth/oauth.go` and `relixctl/internal/auth/devicecode.go` are already written — they just need to be wired to the `login` command.

**pair:**
- Call `POST https://relay.relix.sh/pair/code` with the user ID and mobile public key
- Display the 6-digit code
- Poll `GET https://relay.relix.sh/pair/status/:code` until completed
- Exchange public keys
- Register the machine via `POST https://api.relix.sh/machines`
- Store the machine ID and keys in `~/.relixctl/`
- Display the 4-emoji SAS for verification

Reference: `relixctl/internal/auth/pairing.go` is already written — wire it to the `pair` command in `cmd/`.

Done when: Running `relixctl login` on a dev machine authenticates, `relixctl pair <code>` pairs with the mobile app, and the machine appears in the mobile dashboard.

### Step 7: Build the Mobile App Core Screens

Minimum viable mobile app for beta:

1. **Onboarding / Login screen** — GitHub OAuth button + email/password form
2. **Pair screen** — displays 6-digit code, polls for completion, shows SAS
3. **Dashboard** — lists machines with status, shows pending approval cards
4. **Session view** — chat mode showing events, approve/deny buttons

Each screen must connect to the actual API (cloud) and relay (WebSocket). Reference the API docs in `docs/API.md`.

Done when: A developer can install the app, log in, pair a machine, start Claude Code, and approve a tool use from the app.

### Step 8: Implement Real Push Notifications

1. APNs: get the `.p8` key from Apple Developer Portal, implement `apns2` client in `cloud/internal/push/apns.go`
2. FCM: get the Firebase server key, implement FCM client in `cloud/internal/push/fcm.go`
3. Implement database persistence of device tokens in `POST /push/register`
4. Wire relay → cloud push: when the relay receives an approval-needed event with no mobile client connected, call `POST /api.relix.sh/push/send` with the user's stored device tokens

Done when: Walking away from a Claude Code session with an active approval request triggers a push notification on the test device within 30 seconds.

### Step 9: End-to-End Test

Perform this test manually before any beta announcement:

1. Fresh phone, fresh laptop — install app and relixctl from scratch
2. Sign up with email (not GitHub — test email path)
3. Pair the machine
4. Start a Claude Code session that will make a file edit
5. Walk away from the laptop (or lock the screen)
6. Receive the push notification on the phone
7. Tap the notification
8. Approve the tool use from the approval card
9. Confirm on the laptop that Claude Code continued
10. Send a follow-up message from the phone
11. Confirm Claude Code received and responded to the message

If any step fails, fix it before announcing beta.

Done when: All 11 steps complete without issues on two separate test devices.

### Step 10: Implement Stripe Billing

1. Create Stripe products and prices (see DEPLOYMENT.md section 11)
2. Replace `StubStripe` with real Stripe implementation using `stripe-go`
3. Implement `POST /billing/webhook` to update user tier on subscription events
4. Implement `POST /billing/portal` for subscription management
5. Test with Stripe test mode: upgrade from free to Plus, verify tier updates in `GET /billing/plan`, verify machine limit increases

Done when: A test payment upgrades the user from free to Plus, the machine limit increases to 10, and the billing portal allows cancellation.

### Step 11: App Store Submission

1. Complete all app store metadata (see DEPLOYMENT.md section 9)
2. Build production iOS binary: `eas build --platform ios --profile production`
3. Submit to TestFlight
4. Wait for Apple review (typically 1-3 days for initial review)
5. Build production Android binary and submit to Play Store internal track

Done when: App is available on TestFlight and can be installed on a test device via TestFlight.

### Step 12: Invite First Beta Users

With the above steps complete, you have a working product. Follow the beta strategy in `docs/GO-TO-MARKET.md` section 4, Phase 1.

Target: 50 users, 70% pairing completion rate, at least 1 paying customer within 2 weeks.

---

## 3. Accounts and Services to Create

Create all of these before beginning Step 3 of the critical path. Keep all credentials in a password manager.

| Service | URL | Purpose | Notes |
|---------|-----|---------|-------|
| Fly.io | https://fly.io | Hosts relay and cloud | Free plan is sufficient for launch |
| GitHub (org) | https://github.com/relixdev | Source code hosting, OAuth | Create org `relixdev`, create OAuth App (see DEPLOYMENT.md) |
| Stripe | https://stripe.com | Billing | Start in test mode, switch to live for launch |
| Apple Developer | https://developer.apple.com | iOS app + APNs | $99/year — required for any iOS distribution |
| Google Play Console | https://play.google.com/console | Android app | $25 one-time |
| Firebase | https://console.firebase.google.com | FCM push (Android) | Free tier is sufficient |
| Expo / EAS | https://expo.dev | Mobile build pipeline | Free tier: 30 builds/month |
| Cloudflare | https://cloudflare.com | DNS, domain, CDN | Register `relix.sh` here |
| Posthog | https://posthog.com | Product analytics | Free up to 1M events/month |
| Better Uptime | https://betteruptime.com | Uptime monitoring | Free tier is sufficient |

---

## 4. Secrets and Credentials to Generate

Generate these yourself. Store every one in 1Password or Bitwarden immediately after generating. Do not store in code, git, `.env` files committed to git, or anywhere that could be exposed.

| Secret | How to Generate | Where to Use |
|--------|----------------|--------------|
| `JWT_SECRET` | `openssl rand -hex 32` | Fly.io secret for `relix-cloud`: `fly secrets set JWT_SECRET=...` |
| `RELAY_JWT_SECRET` | Use the **same value** as `JWT_SECRET` | Fly.io secret for `relix-relay`: `fly secrets set RELAY_JWT_SECRET=...` |
| GitHub OAuth Client ID | From GitHub OAuth App settings | Fly.io secret: `GITHUB_CLIENT_ID` |
| GitHub OAuth Client Secret | From GitHub OAuth App settings | Fly.io secret: `GITHUB_CLIENT_SECRET` |
| Stripe Secret Key (test) | From Stripe dashboard → Developers → API keys | Fly.io secret: `STRIPE_SECRET_KEY` (use `sk_test_...` initially) |
| Stripe Secret Key (live) | From Stripe dashboard | Replace test key when ready for real payments |
| Stripe Webhook Secret | From Stripe dashboard → Webhooks | Fly.io secret: `STRIPE_WEBHOOK_SECRET` |
| APNs Private Key (.p8) | From Apple Developer → Keys | Fly.io secret: `APNS_PRIVATE_KEY` (paste file contents) |
| APNs Key ID | From Apple Developer → Keys | Fly.io secret: `APNS_KEY_ID` |
| Apple Team ID | From Apple Developer → Membership | Fly.io secret: `APNS_TEAM_ID` |
| FCM Server Key | From Firebase → Project Settings → Cloud Messaging | Fly.io secret: `FCM_SERVER_KEY` |
| Expo Token | From expo.dev → Access Tokens | GitHub Actions secret: `EXPO_TOKEN` |
| Fly.io API Token | `fly tokens create deploy -x 999999h` | GitHub Actions secret: `FLY_API_TOKEN` |

**Critical:** `JWT_SECRET` and `RELAY_JWT_SECRET` must be the same value. Agents authenticate with the relay using JWTs issued by the cloud service. If the secrets differ, agents cannot connect to the relay.

---

## 5. Known Issues and Technical Debt

### Critical (blocks production)

1. **In-memory user store does not persist.** Every cloud restart loses all users and machines. Fix: implement PostgreSQL store (Step 5 of critical path). Estimated effort: 4-6 hours.

2. **relixctl login and pair are stubs.** The code to implement them exists in `relixctl/internal/auth/` but is not wired to the CLI commands. Fix: wire `oauth.go`, `devicecode.go`, and `pairing.go` to the cobra commands. Estimated effort: 2-3 hours.

3. **APNs and FCM are stubs.** Push notifications do not fire. This breaks the core value proposition — the product is significantly less useful without push. Fix: implement `apns2` and FCM clients (see DEPLOYMENT.md section 10). Estimated effort: 3-4 hours.

4. **Stripe is a stub.** No payments can be processed. Fix: implement `stripe-go` checkout and webhook handler. Estimated effort: 4-6 hours.

5. **Mobile app screens are not built.** The Expo project is bootstrapped but no screens are implemented. This requires the most work. Estimated effort: 40-60 hours for MVP screens.

### Non-Critical (defer until after beta)

6. **No email verification.** Accounts are created with unverified emails. This allows spam registrations. Defer to after beta when scale justifies it.

7. **No rate limiting on auth endpoints.** A brute-force attack on `POST /auth/email/login` is possible. Add per-IP rate limiting before public launch. Redis is already wired in config for this purpose.

8. **JWT TTL is 24 hours with no refresh enforcement.** The mobile app should refresh tokens proactively. If a token expires, the user is logged out. Acceptable for beta; add proactive refresh before v1.

9. **Machine Registry is in-memory.** Same issue as user store — machines are lost on restart. Fixed automatically when PostgreSQL store is implemented (same store interface).

10. **Relay connection state is in-memory.** Expected and documented. See architecture decision #3 below. Not a bug — a deliberate tradeoff.

11. **No PATCH /machines/:id (rename).** Users cannot rename machines from the app. The spec calls for it. Deferred from initial implementation. Estimated effort: 1 hour.

12. **No CORS configuration on cloud service.** The API does not return CORS headers. This is fine for mobile apps (CORS is browser-only), but will block any web dashboard you build. Add CORS middleware before building a web app.

13. **go.mod version inconsistency.** `relay` and `relixctl` use `go 1.26.1` while `cloud` uses `go 1.22`. Standardize on `go 1.22` (the minimum version you actually need) to avoid accidentally depending on future language features.

14. **No structured logging.** Uses `log.Printf` throughout. For production observability, switch to `log/slog` (stdlib, Go 1.21+). Deferred — plain text logs are searchable and sufficient for launch.

15. **No request IDs.** HTTP requests have no correlation ID, making it hard to trace a specific request through logs. Add a middleware that injects `X-Request-ID` before scaling the team.

---

## 6. Architecture Decisions and Rationale

These decisions are load-bearing. Changing them requires carefully auditing all dependents.

### Decision 1: Single JWT Secret Shared Between Relay and Cloud

The relay and cloud use the same `JWT_SECRET`/`RELAY_JWT_SECRET`. This means:
- Cloud issues tokens, relay validates them
- Agents authenticate with the relay using cloud-issued tokens
- One secret rotation affects both services simultaneously

**Why:** Simplest possible architecture. No inter-service token exchange. The relay is a dumb pipe — it only needs to validate "is this a real Relix user?" not "what are their permissions?"

**Tradeoff:** Rotating the JWT secret logs out all users and agents simultaneously (JWTs expire in 24h anyway). Acceptable.

**Alternative considered:** Separate secrets, cloud issues relay-specific tokens. Rejected as over-engineered for current scale.

### Decision 2: Relay Is Stateless by Design

The relay holds connection state in memory. On restart, all WebSocket connections drop. Clients reconnect automatically (exponential backoff in the agent, and the mobile app reconnects on next foreground). Buffered messages are lost.

**Why:** Durable connection state requires external storage (Redis pub/sub or similar). This adds operational complexity, cost, and latency. A restart of the relay is rare and the reconnection is automatic within 60 seconds. The data loss (buffered messages not yet delivered) is acceptable — Claude Code was already waiting for the user.

**Tradeoff:** Cannot scale relay to multiple instances without Redis-backed state. For launch with <10K concurrent connections, one instance is sufficient.

**When to revisit:** When you need >1 relay instance for capacity or availability. At that point, add Redis pub/sub for cross-instance message routing.

### Decision 3: All Connections Outbound

Both agents and mobile clients dial out to the relay. Neither is a server. This means:
- Agents work behind corporate firewalls, NAT, VPNs
- No port forwarding required
- Works on cellular (mobile switches IPs, reconnects)

**Why:** The target users run agents on laptops and corporate servers. Port forwarding is a non-starter in most corporate environments. Outbound WebSocket to port 443 works everywhere.

**What it requires:** The relay must be publicly accessible and always-on. This is why we deploy relay to Fly.io with `min_machines_running = 1`.

### Decision 4: E2E Encryption with X25519 + NaCl Box

All session content is encrypted before it reaches the relay. The relay routes opaque blobs. The relay cannot read session content.

**Why:**
1. Enterprise compliance — "your code goes through our servers" is a dealbreaker for many companies
2. Trust building — open source relay + E2E encryption lets security-conscious users verify the claims
3. Competitive differentiation — Anthropic Remote Control decrypts at Anthropic's servers

**Implementation:** `protocol/crypto.go` handles all crypto. X25519 key exchange via the pairing flow, NaCl box for per-message encryption. The `seq` field inside the encrypted payload prevents replay attacks.

**What this means operationally:** Relix Cloud cannot provide "what did this user's agent do" logs. This is intentional and a feature, not a limitation.

### Decision 5: Go for Everything Server-Side

Agent, relay, and cloud are all Go.

**Why:**
- Single language: one set of tooling, one mental model
- Trivial cross-compilation: `GOOS=linux GOARCH=arm64 go build` produces a static binary for any target
- Excellent WebSocket and concurrency support
- Tiny binaries: the relay is ~10MB, the agent is ~15MB (after `ldflags="-s -w"`)
- Strong standard library: the cloud uses only `net/http` with no framework

**No external HTTP frameworks.** The server uses Go 1.22's enhanced `http.ServeMux` with method+path patterns (`"POST /auth/github"`). This is deliberate — frameworks add complexity without adding value at this scale.

### Decision 6: CopilotAdapter Interface as Extension Point

The `CopilotAdapter` interface in `protocol/protocol.go` is the primary extension point for adding new tools. It was designed before any adapters were built, then the Claude Code adapter was built to satisfy it.

**Why:** This enforces that every adapter implements the same four operations: Discover, Attach, Send, Detach. The daemon is adapter-agnostic — it loops over all adapters and calls the same methods regardless of which tool is being bridged.

**What this means:** Adding a new tool (Aider, Cline) requires only implementing a new struct that satisfies `CopilotAdapter`. No changes to the relay, cloud, mobile, or protocol. See DEVELOPMENT.md section 6 for the step-by-step guide.

### Decision 7: Expo / React Native for Mobile

**Why:**
- EAS Build: no Xcode or Android Studio required for production builds
- OTA updates: JS-layer changes deploy without App Store review (Expo Updates)
- Push notifications: `expo-notifications` handles APNs and FCM token registration
- Single codebase for iOS and Android

**Tradeoff:** If Expo cannot provide a required native capability (unlikely for this app), you would need to eject to a bare workflow. The risk is low — the app uses: WebSocket, push notifications, biometrics, secure storage, and navigation. All are well-supported by Expo packages.

---

## 7. What NOT to Change

These things are stable, correct, and changes to them will break production.

### The Wire Protocol

`protocol/protocol.go` defines the Envelope format, message types, and PayloadKind constants. The relay routes based on the `type` and `machine_id` fields in the envelope. The agent and mobile app parse the `type` field to dispatch messages.

**Do not rename fields.** `machine_id`, `session_id`, `v`, `type`, `timestamp`, `payload` — any rename breaks existing agents and mobile clients in production.

**Do not remove message types.** Even if you add new message types, keep old ones for backward compatibility. The `v` field enables protocol versioning — use it.

**Do not change the `payload` encryption.** The NaCl box scheme in `protocol/crypto.go` is the security foundation. Changing it requires coordinated deploy of all three components (relay, agent, mobile). If you must change the crypto (e.g., to add authenticated encryption metadata), increment `v` and negotiate on connect.

### The Encryption Scheme

`golang.org/x/crypto/nacl/box` for Go, `libsodium-wrappers` for React Native. NaCl box is a well-audited, high-performance AEAD cipher. Do not replace it with a custom scheme. Do not "simplify" it.

The SAS (Short Authentication String) is derived from the X25519 shared secret after key exchange. It is the user-visible proof that the pairing completed correctly. Do not change the derivation algorithm — existing paired devices would show different SAS values.

### The CopilotAdapter Interface

Adding methods to `CopilotAdapter` in `protocol/protocol.go` breaks all existing adapter implementations. If you need to extend the interface, consider:
1. Adding an optional second interface (e.g., `CopilotAdapterV2`) that embeds `CopilotAdapter`
2. Using functional options on the adapter struct

The daemon currently calls `Discover`, `Attach`, `Send`, and `Detach`. These four methods are stable.

### The ID Format

Machine IDs use the format `mch_<random>` (from `idgen.New("mch")`). These IDs are stored in agent configs (`~/.relixctl/config.json`) on developer machines. If you change the format, existing deployed agents will have stale machine IDs that the new cloud does not recognize.

User IDs use `usr_<random>`. Same concern — they are embedded in JWTs.

### JWT Claims Structure

The JWT `Claims` struct in `cloud/internal/auth/jwt.go` has `Subject` (user ID) and `Role`. The relay validates these claims to authenticate WebSocket connections. If you change the claims structure (rename fields, change types), you must update both cloud and relay simultaneously and all in-flight tokens will be invalid.

---

## 8. What SHOULD Be Changed

These are explicit known-incomplete items. Changing them is required, not optional.

### Highest Priority (Required Before Beta)

| Item | Location | What to Do |
|------|----------|------------|
| In-memory user store | `cloud/cmd/cloud/main.go` line 23 | Replace `user.NewMemoryStore()` with `user.NewPostgresStore(db)` |
| relixctl login stub | `relixctl/cmd/root.go` line 34 | Replace `stubCmd("login", ...)` with a real cobra command wired to `relixctl/internal/auth/oauth.go` |
| relixctl pair stub | `relixctl/cmd/root.go` line 35 | Replace `stubCmd("pair", ...)` with a real cobra command wired to `relixctl/internal/auth/pairing.go` |
| APNs stub | `cloud/internal/push/apns.go` | Replace log statement with real `apns2` HTTP/2 client |
| FCM (missing) | `cloud/internal/push/` | Create `fcm.go` implementing `push.Service` using the FCM v1 HTTP API |
| Push token persistence | `cloud/internal/api/push_handlers.go` line 24 | Remove the `// TODO` and implement database storage of device tokens |
| Mobile app screens | `mobile/` | Build onboarding, pair, dashboard, session view screens |

### Medium Priority (Required Before Public Launch)

| Item | Location | What to Do |
|------|----------|------------|
| Stripe real implementation | `cloud/internal/billing/stripe.go` | Replace `StubStripe` with real implementation using `github.com/stripe/stripe-go/v76` |
| Stripe webhook handler | `cloud/internal/api/server.go` | Add `POST /billing/webhook` route and handler that updates user tier |
| Stripe billing portal | `cloud/internal/billing/stripe.go` | Add `CreatePortalSession()` to `StripeService` interface and implement |
| PATCH /machines/:id | `cloud/internal/api/` | Implement rename handler (see DEVELOPMENT.md section 7) |
| Rate limiting on auth | `cloud/internal/api/server.go` | Add per-IP rate limit middleware for auth endpoints |
| go.mod version alignment | `relay/go.mod`, `relixctl/go.mod`, `cloud/go.mod` | Standardize on `go 1.22` |

### Lower Priority (Deferred)

| Item | What to Do |
|------|------------|
| Structured logging | Replace `log.Printf` with `log/slog` throughout |
| Request IDs | Add middleware that generates and propagates `X-Request-ID` |
| Email verification | Add verification flow for email signups |
| Aider adapter | Implement `relixctl/internal/adapter/aider.go` |
| Cline adapter | Implement `relixctl/internal/adapter/cline.go` |
| Cursor adapter | Research integration surface, implement |
| Web dashboard | Build `app.relix.sh` for browser-based session viewing |
| CORS middleware | Required before any browser-facing API calls |

---

## 9. Estimated Timeline to Beta

Assumes one AI agent working full-time on this codebase, with no blockers from external reviews (App Store review is the only external dependency).

| Step | Work | Calendar Days |
|------|------|--------------|
| Push code to GitHub | Set up repos, push all code | Day 1 |
| Register domain | Register relix.sh, configure Cloudflare | Day 1 |
| Deploy relay + cloud | Fly.io setup, Docker images, DNS | Day 2 |
| PostgreSQL store | Implement, deploy, test | Day 3-4 |
| relixctl login + pair | Wire existing auth code to CLI | Day 4-5 |
| APNs + FCM | Implement real push clients | Day 5-6 |
| Mobile app MVP screens | Onboarding, pair, dashboard, session view | Day 7-16 |
| End-to-end test | Full user journey test | Day 17 |
| App Store submission | TestFlight + Play Store | Day 17-18 |
| Apple review | External dependency — typically 1-3 days | Day 19-22 |
| Beta user invites | First 50 users | Day 22-23 |

**Earliest beta launch: ~23 days from start.**

This assumes the mobile app is the longest pole. The core Go infrastructure (relay, cloud, agent) is largely complete; it's the mobile screens and the stubs-to-real-implementations that drive the timeline.

---

## 10. Estimated Timeline to Public Launch

Public launch = landing page live, app publicly available on App Store and Play Store, Product Hunt submission, first Reddit/HN post.

| Milestone | After Beta Start |
|-----------|-----------------|
| Beta feedback incorporated (top 3 bugs fixed) | Week 2 |
| Stripe billing working (first paying customer possible) | Week 3 |
| Aider adapter shipped (multi-tool narrative activated) | Week 4 |
| Landing page live with demo video | Week 4 |
| Homebrew tap + install script | Week 4 |
| App Store production release (vs TestFlight only) | Week 5 |
| Product Hunt launch | Week 5 (Tuesday/Wednesday) |
| Reddit / HN posts | Week 5 |
| Cline adapter | Week 6 |

**Earliest public launch: ~5-6 weeks from beta start (7-8 weeks from today).**

The gating factors:
1. Apple App Store review for the production build (not TestFlight) — typically 3-7 days
2. First paying customer validation — want at least 1 before public launch to confirm Stripe end-to-end
3. Multi-tool story — launching with only Claude Code is weaker than launching with Claude Code + Aider; worth the extra 2 weeks

**Recommendation:** Do not rush the public launch. Beta → fix top bugs → add Aider → launch. A single HN post with "works with Claude Code and Aider" will outperform a rushed launch with "works with Claude Code only" when Anthropic already offers that for free.

---

## Appendix: File Index for Key Stubs

Quick reference for every stub and its file location:

| Stub | File | Line | Status |
|------|------|------|--------|
| Stripe | `cloud/internal/billing/stripe.go` | Full file | Replace with real Stripe |
| APNs | `cloud/internal/push/apns.go` | Full file | Replace with apns2 |
| FCM | Does not exist | — | Create `cloud/internal/push/fcm.go` |
| Push token persistence | `cloud/internal/api/push_handlers.go` | Line 24 | Add database insert |
| In-memory user store | `cloud/cmd/cloud/main.go` | Line 23 | Replace with PostgreSQL |
| relixctl login | `relixctl/cmd/root.go` | Line 34 | Replace stubCmd with real impl |
| relixctl pair | `relixctl/cmd/root.go` | Line 35 | Replace stubCmd with real impl |
| Billing portal | Not implemented | — | Add to billing_handlers.go |
| Stripe webhook | Not implemented | — | Add route + handler to server.go |
| PATCH /machines/:id | Not implemented | — | Add to machine_handlers.go |
