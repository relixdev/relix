# Relix — Frequently Asked Questions

## General

### What is Relix?
Relix is a universal command center for AI coding agents. It connects your AI tools (Claude Code, Aider, Cline) to your phone so you can monitor progress, approve tool use, and send messages — from anywhere.

### How is this different from Anthropic's Remote Control?
Anthropic's Remote Control only works with Claude Code and routes through Anthropic's servers. Relix supports multiple AI coding tools, works across multiple machines, uses end-to-end encryption (zero-knowledge), and sends real push notifications.

### Is my code safe?
Yes. Relix uses end-to-end encryption (X25519 + NaCl box). Your code and AI conversations are encrypted on your machine before they leave — we literally cannot read them. The relay server only sees encrypted blobs.

### Which AI tools does Relix support?
Currently: Claude Code. Coming soon: Aider, Cline, and more. The adapter architecture makes adding new tools straightforward.

### Does it work on Linux?
Yes. relixctl supports macOS (launchd) and Linux (systemd). Windows support is planned.

## Pricing

### Is there a free tier?
Yes. The free tier includes 1 machine and basic features. No credit card required.

### Can I cancel anytime?
Yes. Cancel from the app or email support@shadowscale.dev. We offer pro-rata refunds within 14 days.

### Do you offer annual billing?
Yes, with a discount. Annual plans save approximately 2 months compared to monthly.

## Technical

### How do I install relixctl?
```bash
# macOS (Homebrew)
brew install relixdev/tap/relixctl

# Linux / macOS (curl)
curl -fsSL https://relix.sh/install | sh
```

### How do I pair my phone?
1. Run `relixctl login` to authenticate
2. Run `relixctl pair` to get a 6-digit code
3. Enter the code in the Relix mobile app
4. Done — your phone is connected

### Can I use Relix with multiple machines?
Yes. Each machine runs its own relixctl instance. The mobile app shows all machines in one dashboard. Free tier: 1 machine. Plus: 3 machines. Pro: 10 machines. Team: unlimited.

### Is the relay server open source?
Yes. Both `relixctl` and the relay server are MIT licensed and available at github.com/relixdev.

### Can I self-host the relay?
Yes. The relay is a single Docker image. Run it anywhere you want. Self-hosting means your data never touches our servers at all.

## Support

### How do I get help?
- Email: support@shadowscale.dev
- Discord: discord.gg/relix (coming soon)
- GitHub Issues: github.com/relixdev/relix/issues

### How do I report a bug?
Open an issue at github.com/relixdev/relix/issues using the bug report template.

### How do I request a feature?
Open an issue at github.com/relixdev/relix/issues using the feature request template.
