# Relix — Project Summary

## What It Is

A universal command center for all your AI coding agents. One mobile app, one dashboard, every tool, every machine. Control Claude Code, Aider, Cline, and more from your phone with E2E encryption.

**Company:** Relix | **Domain:** relix.sh | **GitHub:** github.com/relixdev | **CLI:** `relixctl`

## The Problem

Developers run AI coding agents across multiple tools and machines. Walk away from your laptop and your agents are stuck — waiting for approvals, invisible to you. Even Anthropic's built-in Remote Control only works for Claude Code and routes through their servers.

## Differentiation vs Anthropic Remote Control

| | Remote Control | Relix |
|---|---|---|
| Tools supported | Claude Code only | Claude Code, Aider, Cline, Cursor, Continue.dev |
| Multi-machine | Per-session | All machines, one dashboard |
| Encryption | Decrypted at Anthropic | E2E — relay sees only ciphertext |
| Self-hostable | No | Yes (open source relay) |
| Push notifications | No | Yes (APNs + FCM) |
| Team features | No | Shared sessions, audit trail (Team tier) |
| Automation rules | No | Auto-approve reads, block destructive ops |
| Offline queue | No | Buffers messages, delivers on reconnect |

## Architecture

```
relixctl (agent)  →  Relay (WebSocket hub)  ←  Mobile App (React Native)
                           ↑
                      Cloud (auth/billing/push)
```

- All connections outbound (no port forwarding needed)
- NaCl box encryption (X25519 + XSalsa20-Poly1305)
- Open source relay + agent (MIT), proprietary mobile + cloud

## Adapter Roadmap

| Priority | Tool | Market Signal | Timeline |
|----------|------|--------------|----------|
| P0 | Claude Code | 40.8% SO survey, ~$2B ARR | Launch (built) |
| P1 | Aider | 42K GitHub stars, 4.1M installs | +2 weeks |
| P1 | Cline | 59K GitHub stars, 3.3M VS Code installs | +4 weeks |
| P2 | Cursor | 360K paying users, $2B ARR | +8 weeks |
| P2 | Continue.dev | 32K stars, 2.3M VS Code installs | +8 weeks |
| P3 | GitHub Copilot | 4.7M paid subs, 42% market share | TBD |

## Monetization

| | Free | Plus $4.99/mo | Pro $14.99/mo | Team $24.99/user/mo |
|---|---|---|---|---|
| Machines | 3 | 10 | Unlimited | Unlimited |
| Sessions | 2 | 5 | Unlimited | Unlimited |
| History | 7 days | 30 days | 90 days | 90 days |
| Recording | No | Yes | Yes | Yes |
| Priority relay | No | No | Yes | Yes |
| Shared sessions | No | No | No | Yes |

---

## Codebase

| Component | Tech | Lines | Tests | Files |
|-----------|------|-------|-------|-------|
| **protocol/** | Go | ~600 | 23 | 10 |
| **relixctl/** | Go | ~5,000 | 90+ | 48 |
| **relay/** | Go | ~2,800 | 50+ | 30 |
| **cloud/** | Go | ~2,000 | 15+ | 29 |
| **mobile/** | TypeScript/RN | ~4,200 | 15 | 45 |
| **docs/** | Markdown | — | — | 7 |
| **Total** | | **~14,600** | **193** | **175** |

### Protocol (Go module, MIT)
- Wire format: JSON envelopes over WebSocket
- Crypto: X25519 key generation, NaCl box encrypt/decrypt, SealPayload/OpenPayload
- Types: Envelope, Payload, Session, Event, UserInput, CopilotAdapter interface

### Agent — `relixctl` (Go, MIT)
- Claude Code adapter: session discovery via `~/.claude/projects/`, bidirectional streaming via `--input-format stream-json`
- Relay client: WebSocket with auth, reconnecting with exponential backoff, E2E encryption bridge
- Daemon: discover sessions → attach → forward events → dispatch approvals
- Auth: GitHub OAuth, device code flow, 4-emoji SAS pairing
- Services: launchd (macOS), systemd (Linux)
- Commands: login, pair, start, stop, status, sessions, config, uninstall

### Relay (Go, MIT, Docker)
- Hub: connection registry, agent↔mobile routing
- Buffering: per-user queue (1000 msg / 10MB / 24h TTL), drain on reconnect
- Pairing: one-time codes with 5-min TTL, per-IP rate limiting
- Observability: Prometheus metrics, docker-compose with Grafana
- Deployment: distroless multi-stage Docker image

### Cloud (Go, proprietary)
- Auth: GitHub OAuth, email/bcrypt, JWT HS256
- Machine registry with per-tier limit enforcement
- Billing: tier definitions, Stripe stub
- Push: APNs + FCM stubs
- API: net/http with auth middleware

### Mobile (Expo/TypeScript, proprietary)
- Auth: GitHub OAuth + email login, secure token storage
- Dashboard: machine list, inline approval cards, pull-to-refresh
- Session view: chat mode (bubbles) + terminal mode (ANSI, dark theme)
- Pairing: 6-digit code display + 4-emoji SAS verification
- Push: registration, deep linking, iOS action buttons
- Security: biometric auth, app lock after 5 min backgrounded
- Settings: account, machine management, preferences

---

## Repo Structure

```
github.com/relixdev/
├── protocol/       # Shared protocol definitions (MIT)
├── relixctl/       # Agent CLI (MIT)
├── relay/          # WebSocket relay server (MIT)
├── cloud/          # Auth, billing, push backend (proprietary)
├── mobile/         # React Native app (proprietary)
└── docs/
    ├── specs/      # Design specifications
    ├── plans/      # Implementation plans
    └── PROJECT.md  # This file
```

## Next Steps to Production

1. Push to GitHub (github.com/relixdev or personal account)
2. Replace stubs: Stripe, APNs, FCM, PostgreSQL
3. Deploy relay + cloud to Fly.io
4. EAS Build → TestFlight + Play Store internal track
5. End-to-end test: real Claude Code session from phone
6. Build Aider adapter (P1)
7. Register relix.sh domain, landing page
8. Beta launch
