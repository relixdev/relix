# Why We Built Relix

**Published:** 2026-03-14  
**Author:** Zach  
**Reading time:** 4 minutes

---

I've been running AI coding agents for months now. Claude Code, Aider, Cline — whatever's in the mood that day. They're incredible tools. They write code, refactor entire codebases, debug issues I've been staring at for hours.

But there's one problem that kept bugging me: **they get stuck when I walk away.**

## The Problem

You're deep in a refactor. Claude Code is churning through files, making changes, asking for approvals. You step away to grab coffee, answer the door, help your kid with homework. You come back 20 minutes later.

Claude Code is sitting there. Waiting. For an approval you never got.

It's not Claude's fault. It's not the AI's fault. It's that the whole setup assumes you're glued to your desk.

Anthropic's Remote Control helps — you can see Claude's output from your phone. But it only works with Claude Code. It routes through Anthropic's servers (no E2E encryption). And it doesn't support multiple tools or multiple machines.

## The Solution

I built Relix because I wanted:

1. **Multi-tool support** — One app for Claude Code, Aider, Cline, and whatever comes next
2. **Multi-machine** — See all my machines (laptop, server, CI runner) in one dashboard
3. **E2E encryption** — My code stays mine. Zero-knowledge architecture.
4. **Push notifications** — Real notifications when my agent needs attention, not just polling
5. **Self-hostable** — Run it on my own infrastructure if I want

## How It Works

Relix has three components:

**relixctl** — A CLI that runs on your machine. It connects to your AI coding tool (Claude Code, Aider, etc.) and bridges it to the relay server.

**Relay** — A WebSocket server that routes messages between your CLI and your phone. It's stateless, encrypted, and MIT licensed.

**Mobile app** — Your dashboard. See sessions, approve tool use, send messages — from anywhere.

The flow:

```
AI Tool ←→ relixctl ←→ Relay ←→ Mobile App
```

Everything between relixctl and the mobile app is end-to-end encrypted with X25519 key exchange and NaCl box encryption. The relay server only sees encrypted blobs.

## The Tech Stack

- **Go** — relay server, API server, CLI (14,600 lines, 193 tests)
- **React Native** — mobile app (Expo, TypeScript)
- **PostgreSQL** — user store, machine registry, subscriptions
- **Stripe** — billing (test mode, ready for production)
- **WebSocket** — real-time bidirectional communication
- **X25519 + NaCl** — E2E encryption

## What's Next

Relix is in beta. The core is done — CLI, relay, mobile app, encryption, billing. What's left:

- Aider adapter (in progress)
- Cline adapter
- iOS/Android app store submission
- Landing page
- Launch

## Try It

If you're running AI coding agents and you're tired of being tethered to your desk, give Relix a shot.

Free tier: 1 machine, no credit card.

Website: https://relix.sh  
GitHub: https://github.com/relixdev

Built by developers, for developers.
