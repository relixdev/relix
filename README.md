# Relix

**Command center for all your AI coding agents.**

Control Claude Code, Aider, Cline, and more from your phone. One app, every tool, every machine. E2E encrypted.

## How It Works

```
Your Machine                    Cloud                     Your Phone
┌──────────┐              ┌───────────┐              ┌──────────┐
│ relixctl │──WebSocket──▶│   Relay   │◀──WebSocket──│  Mobile  │
│ (agent)  │              │  (router) │              │  (app)   │
└──────────┘              └───────────┘              └──────────┘
     │                         ▲                          │
     │                    ┌────┴────┐                     │
     ▼                    │  Cloud  │                     ▼
  Claude Code             │  (auth) │               Dashboard
  Aider, Cline...        └─────────┘            Sessions, Approvals
```

- **All connections outbound** — works behind firewalls, on cellular, everywhere
- **E2E encrypted** — relay routes opaque blobs, never sees your code
- **Open source relay + agent** — self-host if you want

## Quick Start

### On your phone
1. Download Relix from the App Store / Play Store
2. Sign up and go to "Add Machine"

### On your dev machine
```bash
# Install the agent
curl -fsSL relix.sh/install | sh

# Login
relixctl login

# Pair with your phone
relixctl pair <code-from-app>
```

That's it. Walk away from your laptop — approve tool calls, send messages, and monitor sessions from your phone.

## Features

- **Multi-tool support** — Claude Code today, Aider and Cline coming soon
- **Multi-machine dashboard** — all your dev machines in one view
- **Push notifications** — "Approval needed: Edit src/auth.ts"
- **Chat + Terminal modes** — view sessions your way
- **Biometric lock** — Face ID / fingerprint to open
- **Smart approvals** — approve from notification without opening the app
- **Self-hostable relay** — `docker run ghcr.io/relixdev/relay`

## Architecture

| Component | Description | License |
|-----------|-------------|---------|
| `protocol/` | Wire protocol, crypto primitives, shared types | MIT |
| `relixctl/` | CLI agent — session discovery, bridging, daemon | MIT |
| `relay/` | WebSocket router with buffering and metrics | MIT |
| `cloud/` | Auth, billing, push notifications | Proprietary |
| `mobile/` | React Native app (iOS + Android) | Proprietary |

### Tech Stack

- **Backend:** Go 1.23+
- **Mobile:** React Native (Expo), TypeScript
- **Encryption:** X25519 + XSalsa20-Poly1305 (NaCl box)
- **Transport:** WebSocket (JSON envelopes)
- **Infrastructure:** Fly.io, Stripe, APNs/FCM

## Development

```bash
# Run all Go tests
cd protocol && go test -race ./... && cd ..
cd relixctl && go test -race ./... && cd ..
cd relay && go test -race ./... && cd ..
cd cloud && go test -race ./... && cd ..

# Run mobile tests
cd mobile && npx jest

# Build the agent
cd relixctl && make build

# Build the relay
cd relay && make build

# Run relay locally
cd relay && docker-compose up
```

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for the full development guide.

## Self-Hosting

Run your own relay:

```bash
docker run -d \
  -p 8080:8080 \
  -e RELAY_JWT_SECRET=your-secret \
  ghcr.io/relixdev/relay
```

Point your agent at it:
```bash
relixctl config set relay_url wss://relay.yourdomain.com
```

Point the mobile app at it in Settings → Relay URL.

## Pricing

| | Free | Plus $4.99/mo | Pro $14.99/mo | Team $24.99/user/mo |
|---|---|---|---|---|
| Machines | 3 | 10 | Unlimited | Unlimited |
| Sessions | 2 | 5 | Unlimited | Unlimited |
| History | 7 days | 30 days | 90 days | 90 days |

## Security

- **E2E encryption:** All session data encrypted with NaCl box (X25519 + XSalsa20-Poly1305). The relay never sees plaintext.
- **Key pairing:** 6-digit code + 4-emoji Short Authentication String for MITM protection.
- **Key rotation:** Automatic every 30 days.
- **Mobile security:** Biometric auth, auto-lock after 5 minutes, session data cleared from memory on background.
- **Self-hostable:** Run the relay on your own infrastructure for full control.

## Adapter Roadmap

| Status | Tool | Integration |
|--------|------|-------------|
| ✅ Built | Claude Code | Headless mode, bidirectional streaming |
| 🔜 Next | Aider | Open source CLI, Python API |
| 🔜 Next | Cline | Open source VS Code extension |
| 📋 Planned | Cursor | VS Code fork, extension surface |
| 📋 Planned | Continue.dev | Open source, pluggable |
| 📋 Planned | GitHub Copilot | Largest market, most closed |

## Documentation

- [Project Summary](docs/PROJECT.md)
- [Design Specification](docs/specs/2026-03-11-relix-design.md)
- [Product Requirements](docs/PRD.md)
- [API Reference](docs/API.md)
- [Deployment Guide](docs/DEPLOYMENT.md)
- [Development Guide](docs/DEVELOPMENT.md)
- [Go-to-Market](docs/GO-TO-MARKET.md)

## License

- `protocol/`, `relixctl/`, `relay/` — MIT License
- `cloud/`, `mobile/` — Proprietary

## Links

- **Website:** [relix.sh](https://relix.sh)
- **GitHub:** [github.com/relixdev](https://github.com/relixdev)
- **Discord:** Coming soon
