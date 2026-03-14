# Relix — Launch Content Drafts

## Hacker News — Show HN

**Title:** Show HN: Relix – Control Claude Code, Aider, and Cline from your phone (E2E encrypted)

**Body:**
Hey HN, I built Relix because I kept walking away from my laptop while Claude Code was running, only to come back 20 minutes later and find it stuck waiting for approval.

Relix is a universal mobile command center for AI coding agents. Install a CLI on your machine, pair with the mobile app, and get push notifications + one-tap approvals from your phone.

Key features:
- Works with Claude Code, Aider, Cline (not locked to one tool)
- E2E encrypted (X25519 + NaCl box) — zero knowledge, we can't read your code
- Multi-machine dashboard (laptop, server, CI)
- Self-hostable relay server (MIT licensed)
- Push notifications for approvals, errors, completions

Tech stack: Go (relay + cloud + CLI), React Native (mobile), PostgreSQL, WebSocket.

"But doesn't Anthropic have Remote Control?" — Yes, but it only works with Claude Code, routes through Anthropic's servers (no E2E encryption), and doesn't support multiple tools or machines.

Free tier: 1 machine. Plus: $4.99/mo for 3. Self-host option is completely free.

GitHub (relay + CLI are MIT): https://github.com/relixdev
Website: https://relix.sh

Would love feedback, especially from anyone running AI coding agents regularly.

---

## Reddit — r/programming

**Title:** I built a mobile command center for AI coding agents (Claude Code, Aider, Cline) with E2E encryption

**Body:**
I kept walking away from Claude Code while it was running, only to come back and find it stuck waiting for approval. So I built Relix — a universal mobile dashboard for AI coding agents.

**What it does:**
- Push notifications when your AI agent needs attention
- One-tap approvals for file writes and tool use
- Real-time session monitoring from your phone
- Works with Claude Code, Aider, Cline (multi-tool)
- E2E encrypted — your code never leaves your control
- Self-hostable relay (MIT license)

**How it works:**
```
brew install relixdev/tap/relixctl
relixctl login
relixctl pair  # scan QR or enter 6-digit code
```

**Tech:** Go backend, React Native mobile, WebSocket relay, X25519 + NaCl encryption.

Free tier available. Relay + CLI are open source (MIT).

GitHub: https://github.com/relixdev
Website: https://relix.sh

---

## Reddit — r/ClaudeAI

**Title:** Built a mobile app to control Claude Code from your phone (with push notifications and E2E encryption)

**Body:**
I love Claude Code but hate having to sit at my desk while it works. Every time I walk away, it gets stuck on an approval and sits there until I come back.

So I built Relix — a mobile command center that connects to Claude Code (and other AI tools) via WebSocket with end-to-end encryption.

You get:
- Push notifications when Claude needs approval
- One-tap approve/deny from your phone
- Real-time view of what Claude is doing
- Works across multiple machines
- E2E encrypted (we can't read your code, unlike Anthropic's Remote Control which routes through their servers)

Install relixctl, pair with the app, and you're done. Free tier includes 1 machine.

Also supports Aider and Cline — one app for all your AI coding tools.

Website: https://relix.sh
GitHub: https://github.com/relixdev

---

## Twitter/X — Launch Thread

**Tweet 1:**
🚀 Introducing Relix — the command center for AI coding agents.

Control Claude Code, Aider, and Cline from your phone.

Push notifications. One-tap approvals. E2E encrypted.

Your AI agents don't stop when you walk away anymore.

→ relix.sh

**Tweet 2:**
How it works:

1. `brew install relixdev/tap/relixctl`
2. `relixctl pair`
3. Enter code in the app
4. Get push notifications when your agent needs you

30 seconds to set up. Works on macOS and Linux.

**Tweet 3:**
Why not just use Anthropic's Remote Control?

✅ Relix: multi-tool (Claude + Aider + Cline)
✅ Relix: E2E encrypted (zero knowledge)
✅ Relix: multi-machine dashboard
✅ Relix: self-hostable (MIT)

Remote Control only works with Claude Code and routes through Anthropic's servers.

**Tweet 4:**
Tech stack for the nerds:

- Go relay server (WebSocket, 4.7MB Docker image)
- Go API server (JWT, Stripe, PostgreSQL)
- Go CLI (launchd/systemd daemon)
- React Native mobile (Expo)
- X25519 + NaCl box encryption
- MIT licensed (relay + CLI)

**Tweet 5:**
Free tier: 1 machine, no credit card.

Plus ($4.99/mo): 3 machines
Pro ($14.99/mo): 10 machines
Team ($24.99/user/mo): unlimited

Or self-host the relay for free. Your call.

→ relix.sh

---

## Product Hunt

**Tagline:** Command center for all your AI coding agents

**Description:**
Relix connects your AI coding tools to your phone. Get push notifications when Claude Code needs approval, monitor sessions in real-time, and control everything from a single dashboard — with end-to-end encryption.

Works with Claude Code, Aider, Cline, and more. One app for every AI tool, across every machine.

**Topics:** Developer Tools, Artificial Intelligence, Productivity, Mobile Apps

**Makers:** @zachforsyth

**First Comment:**
"Hey Product Hunt! 👋

I built Relix because I got tired of running back to my laptop every time Claude Code got stuck on an approval.

The idea is simple: your AI coding agents should be as accessible as your text messages. Push notification → one tap → back to what you were doing.

What makes Relix different:
1. Multi-tool — not locked to one AI (Claude, Aider, Cline)
2. E2E encrypted — we can't read your code
3. Self-hostable — the relay is MIT licensed

Would love your feedback. What AI coding tools are you using? What would make this more useful for you?"
