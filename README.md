<p align="center">
  <h1 align="center">⚡ Relix</h1>
  <p align="center"><strong>Command center for all your AI coding agents.</strong></p>
  <p align="center">
    One app to control Claude Code, Aider, Cline, and every AI tool on every machine — with E2E encryption and push notifications.
  </p>
</p>

<p align="center">
  <a href="https://github.com/relixdev/relix/actions"><img src="https://img.shields.io/github/actions/workflow/status/relixdev/relix/test.yml?branch=main&label=tests&style=flat-square" alt="Tests"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue?style=flat-square" alt="License: MIT"></a>
  <a href="https://discord.gg/relix"><img src="https://img.shields.io/badge/discord-join-7289da?style=flat-square&logo=discord&logoColor=white" alt="Discord"></a>
  <a href="https://github.com/relixdev/relix/releases/latest"><img src="https://img.shields.io/github/v/release/relixdev/relix?style=flat-square&color=orange" alt="Latest Release"></a>
</p>

---

<p align="center">
  <em><!-- TODO: Replace with demo GIF --></em><br>
  <code>relixctl pair</code> → push notification → approve from phone → agent continues
</p>

---

## Why Relix

Your AI coding agents are stuck when you walk away. Sessions block for hours waiting for approvals you can't see. You have multiple machines, multiple tools, multiple terminal windows — and no single place to manage them.

Relix fixes that.

## Features

- **🔌 Multi-tool** — Claude Code, Aider, Cline, and more in one dashboard
- **🖥️ Multi-machine** — Every dev machine, server, and CI runner in one view
- **🔐 E2E Encrypted** — NaCl box encryption. The relay routes ciphertext — it never sees your code
- **🔔 Push Notifications** — "Approval needed: Edit src/auth.ts" straight to your phone
- **🏠 Self-hostable** — MIT-licensed relay. Run it on your own infrastructure

## Quick Start

### Install the CLI

```bash
# Homebrew
brew install relixdev/tap/relixctl

# or curl
curl -fsSL relix.sh/install | sh

# or from source
git clone https://github.com/relixdev/relix.git
cd relix/relixctl && make build
```

### Pair with your phone

```bash
# 1. Download Relix from the App Store / Play Store
# 2. Sign up and tap "Add Machine"

# 3. On your dev machine:
relixctl login
relixctl pair <code-from-app>
```

### Start coding

Walk away from your laptop. When Claude Code needs approval, your phone buzzes. Tap **Allow**. The agent continues. That's it.

## Architecture

```
Your Machines                     Cloud                      Your Phone
┌──────────────┐            ┌─────────────┐            ┌──────────────┐
│   relixctl   │──outbound──▶    Relay    ◀──outbound──│    Mobile    │
│              │  WebSocket  │  (router)  │  WebSocket  │     App      │
│  ┌─────────┐ │            └──────┬──────┘            │ ┌──────────┐ │
│  │Claude   │ │                   │                   │ │Dashboard │ │
│  │Code     │ │            ┌──────┴──────┐            │ │Sessions  │ │
│  ├─────────┤ │            │    Cloud    │            │ │Approvals │ │
│  │Aider    │ │            │  (auth +   │            │ │Chat      │ │
│  ├─────────┤ │            │  billing)  │            │ └──────────┘ │
│  │Cline    │ │            └─────────────┘            └──────────────┘
│  └─────────┘ │
└──────────────┘

  All connections outbound — works behind firewalls, NATs, and on cellular.
  Relay sees only encrypted blobs — zero-knowledge by design.
```

## Supported Tools

| Tool | Status | Integration |
|------|--------|-------------|
| **Claude Code** | ✅ Supported | Headless mode, bidirectional streaming |
| **Aider** | ✅ Supported | CLI adapter, Python API |
| **Cline** | 🔜 Coming soon | VS Code extension bridge |
| **Cursor** | 📋 Planned | VS Code fork surface |
| **Continue.dev** | 📋 Planned | Open plugin API |
| **GitHub Copilot** | 📋 Planned | Extension surface |

## Pricing

|  | **Free** | **Plus** $4.99/mo | **Pro** $14.99/mo | **Team** $24.99/user/mo |
|--|----------|-------------------|-------------------|-------------------------|
| Machines | 3 | 10 | Unlimited | Unlimited |
| Concurrent sessions | 2 | 5 | Unlimited | Unlimited |
| Session history | 7 days | 30 days | 90 days | 90 days |
| Priority relay |  |  | ✅ | ✅ |
| Shared sessions |  |  |  | ✅ |
| SSO / Admin |  |  |  | ✅ |

The free tier is permanent. No credit card required. [Upgrade →](https://relix.sh/pricing)

## Self-Hosting

Run your own relay for full control over your data:

```bash
docker run -d \
  -p 8080:8080 \
  -e RELAY_JWT_SECRET=your-secret \
  ghcr.io/relixdev/relay
```

Point your agent and mobile app at it:

```bash
relixctl config set relay_url wss://relay.yourdomain.com
# Mobile: Settings → Relay URL → wss://relay.yourdomain.com
```

The relay is MIT-licensed. Inspect every byte that touches your infrastructure.

## Project Structure

| Component | Description | License |
|-----------|-------------|---------|
| [`protocol/`](protocol/) | Wire protocol, crypto primitives, shared types | MIT |
| [`relixctl/`](relixctl/) | CLI agent — session discovery, bridging, daemon | MIT |
| [`relay/`](relay/) | WebSocket router with buffering and metrics | MIT |
| `cloud/` | Auth, billing, push notifications | Proprietary |
| `mobile/` | React Native app (iOS + Android) | Proprietary |

**Tech stack:** Go 1.23+ · React Native (Expo) · X25519 + XSalsa20-Poly1305 · WebSocket · Fly.io

## Contributing

We welcome contributions to the open-source components (`protocol/`, `relixctl/`, `relay/`).

1. Fork the repo
2. Create a feature branch (`git checkout -b feat/my-feature`)
3. Write tests for your changes
4. Ensure all tests pass: `go test -race ./...`
5. Open a pull request

See [DEVELOPMENT.md](docs/DEVELOPMENT.md) for the full development guide.

## Security

Relix uses **NaCl box** (X25519 + XSalsa20-Poly1305) for end-to-end encryption. The relay is zero-knowledge — it routes encrypted blobs and never sees plaintext session data.

- **Key exchange:** 6-digit pairing code + 4-emoji Short Authentication String (SAS) for MITM protection
- **Key rotation:** Automatic every 30 days
- **Mobile:** Biometric auth, auto-lock, session data cleared on background

**Found a vulnerability?** Please report it responsibly via [security@relix.sh](mailto:security@relix.sh). Do not open a public issue for security vulnerabilities.

## License

The open-source components (`protocol/`, `relixctl/`, `relay/`) are licensed under the [MIT License](LICENSE).

`cloud/` and `mobile/` are proprietary.

## Links

[Website](https://relix.sh) · [Documentation](https://relix.sh/docs) · [Discord](https://discord.gg/relix) · [Twitter](https://twitter.com/relixdev) · [Blog](https://relix.sh/blog) · [GitHub](https://github.com/relixdev)
