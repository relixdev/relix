# Relix — Go-to-Market Playbook

This document is a complete, executable GTM playbook for a one-person AI-run company. Every action is specified. No human judgment calls are left open.

---

## Table of Contents

1. [Positioning](#1-positioning)
2. [Pre-Launch Checklist](#2-pre-launch-checklist)
3. [Landing Page](#3-landing-page)
4. [Launch Strategy](#4-launch-strategy)
5. [Content Strategy](#5-content-strategy)
6. [Growth Levers](#6-growth-levers)
7. [Conversion Funnel](#7-conversion-funnel)
8. [Pricing Psychology](#8-pricing-psychology)
9. [Metrics to Track](#9-metrics-to-track)

---

## 1. Positioning

### What Not to Say

Do NOT lead with "control Claude Code from your phone." Anthropic does this for free with their Remote Control feature. Competing on a feature Anthropic gives away for free is a losing strategy.

### What to Say

**Tagline:** "Command center for all your AI coding agents."

**One-liner:** "One app to control Claude Code, Aider, Cline, and every AI tool on every machine — with E2E encryption and push notifications."

**Elevator pitch:** "AI coding agents are stuck when you walk away from your laptop. Relix fixes that. It's a universal mobile dashboard for every AI coding tool — Claude Code, Aider, Cline, and more — with real-time approvals, push notifications, and E2E encryption so your code never leaves your control."

### Why Developers Will Use It

1. **Multi-tool:** The only tool that supports Claude Code *and* Aider *and* Cline in one place
2. **Multi-machine:** See all your machines (laptop, server, CI) in one dashboard
3. **E2E encrypted:** Enterprise developers cannot use Anthropic's Remote Control because it routes through Anthropic's servers; Relix is self-hostable with zero-knowledge encryption
4. **Push notifications:** The only solution that pushes "approval needed" to your phone

### Target Persona

**Primary:** Professional developer running AI coding agents frequently. Has multiple machines. Values productivity. Pays for tools. Not price-sensitive to $4.99/month.

**Secondary:** Engineering team lead who wants to give remote visibility to a team. Inbound from Primary persona.

---

## 2. Pre-Launch Checklist

Complete all items before announcing to anyone. Each item has a concrete done state.

### Infrastructure (Done When Deployed and Healthy)

- [ ] `relix.sh` domain registered and DNS configured
- [ ] Relay deployed to `relay.relix.sh` — WebSocket connects successfully
- [ ] Cloud deployed to `api.relix.sh` — `/health` returns `{"status":"ok"}`
- [ ] PostgreSQL provisioned and connected (replace in-memory store)
- [ ] APNs credentials configured — real push notifications fire on iOS
- [ ] FCM configured — real push notifications fire on Android
- [ ] Stripe configured with real products, prices, and webhook
- [ ] End-to-end test passes: install relixctl → pair → start Claude Code → send message from phone → receive response → approve tool use from phone

### App Store Presence (Done When Listed)

- [ ] iOS app submitted to TestFlight with all required metadata
- [ ] Android app submitted to Play Store internal testing track
- [ ] Both stores have: icon, screenshots, description, privacy policy URL
- [ ] Privacy policy page live at `https://relix.sh/privacy`
- [ ] Support email configured: `support@relix.sh`

### Web Presence (Done When Live)

- [ ] Landing page live at `relix.sh` (see section 3)
- [ ] GitHub org created: `github.com/relixdev`
- [ ] `relixctl` and `relay` repos public (MIT license)
- [ ] README files complete on both public repos
- [ ] Homebrew tap published: `relixdev/homebrew-tap`
- [ ] Install script live at `relix.sh/install` (for `curl -fsSL relix.sh/install | sh`)

### Social Presence (Done When Handles Secured)

- [ ] Twitter/X: `@relixdev` (or `@relix_sh`)
- [ ] Reddit account: `u/relixdev`
- [ ] GitHub org: `relixdev`
- [ ] Hacker News account: `relixdev` (create at news.ycombinator.com)
- [ ] Discord server created with `#announcements`, `#support`, `#beta-feedback` channels

### Content (Done When Published)

- [ ] Demo video recorded and uploaded (see section 5 for script)
- [ ] Product Hunt page drafted (not submitted yet)
- [ ] At least 2 blog posts drafted and ready to publish

---

## 3. Landing Page

The landing page at `relix.sh` must answer: "What is this, why do I want it, how do I get it, what does it cost, can I trust you?" in under 10 seconds of reading.

### Page Structure

**Above the fold:**
- Headline: "Command center for all your AI coding agents"
- Subheadline: "Control Claude Code, Aider, Cline, and more from your phone. Push notifications, instant approvals, E2E encrypted."
- Primary CTA: "Download for iOS" + "Get it on Google Play"
- Secondary CTA: "See how it works ↓" (scrolls to demo)
- Hero: Animated phone mockup showing the dashboard with approval card

**Section 2 — The Problem:**
- "Your AI agents are stuck when you walk away"
- Three pain points in cards:
  1. "Claude Code waiting for approval since 3pm — you're at dinner"
  2. "Seven machines, seven terminal windows you can't see"
  3. "Your session is going through Anthropic's servers"

**Section 3 — The Solution (demo video or GIF):**
- Show: phone buzz → approval card → tap Approve → Claude Code continues
- 30-second screen recording is more powerful than any text here

**Section 4 — Features:**
Three columns:
- "Universal" — Claude Code, Aider, Cline, Cursor (coming soon). Show tool logos.
- "Private" — E2E encrypted. Relay sees only ciphertext. Self-hostable.
- "Smart" — Auto-approve reads. Block destructive operations. Configurable rules.

**Section 5 — Multi-machine dashboard screenshot**
- Full-width iPhone screenshot showing 3+ machines with session states

**Section 6 — Pricing:**
Four-tier table matching the product (Free / Plus $4.99 / Pro $14.99 / Team $24.99). Highlight Plus as "Most Popular".

**Section 7 — Open source trust:**
- "The relay and agent are MIT-licensed and on GitHub. Inspect every byte that touches your machine."
- Links to `github.com/relixdev/relay` and `github.com/relixdev/relixctl`

**Section 8 — Install:**
```bash
brew install relixdev/tap/relixctl
# or
curl -fsSL relix.sh/install | sh
```

**Section 9 — FAQ:**

Q: Does Relix see my code?
A: No. The relay routes encrypted blobs. We can't read your session content even if compelled.

Q: Does it work with Claude's built-in Remote Control?
A: It's an alternative for users who want multi-tool support, multi-machine dashboards, and E2E encryption.

Q: What tools are supported?
A: Claude Code at launch. Aider and Cline within 2 weeks. Cursor and Continue.dev within 8 weeks.

Q: Can I self-host?
A: Yes. Run `docker run ghcr.io/relixdev/relay` and point your agent at it with `relixctl config set relay_url wss://your.server.com`.

**Footer:**
- Links: Pricing, Docs, GitHub, Privacy, Terms
- Copyright: Relix — relix.sh

### Copy Don'ts

- Don't say "like Anthropic Remote Control but better" — this invites comparison to a free product
- Don't use "revolutionary" or "game-changing"
- Don't say "AI-powered" — the product facilitates AI agents, it is not itself AI-powered
- Don't use more than 3 features in any bullet list — pick the best 3

---

## 4. Launch Strategy

### Phase 0: Silent Beta (Before Any Announcement)

Objective: Find 10 people who will give brutally honest feedback. Not cheerleaders.

Where to find them:
- Personal network: developers you've worked with who use AI coding agents
- Twitter/X DMs to active AI coding tool users (people who post about Claude Code, Aider, Cline)
- No public posts yet

What to ask:
- "I built a tool that lets you control AI coding agents from your phone. Can I get 15 minutes of your time to watch you use it?"
- Give them a TestFlight link, sit with them (virtually), watch where they get stuck
- Record every point of confusion — these are your onboarding problems

Done when: 10 people have successfully paired a machine and sent a message from their phone.

### Phase 1: Closed Beta (50-100 Developers)

Objective: Get enough users to validate retention (do people come back day 2, day 7?) and find the top 3 bugs.

**Week 1-2: Reddit**

Target subreddits in this order:
1. r/ClaudeAI (180K members) — highest overlap with Claude Code users
2. r/cursor (growing fast) — Cursor users will want Aider/Cline support
3. r/LocalLLaMA — power users, self-hosters
4. r/MachineLearning — developers who run agents on servers

Post title formula: `[Show HN style] I built X because Y — here's how it works`

Example Reddit post for r/ClaudeAI:

```
Title: I built a mobile dashboard for Claude Code because I kept missing approvals

Been running Claude Code sessions for 8+ hours on my dev servers and kept
getting stuck when I walked away. Sessions would block for hours waiting for
approval.

Built Relix — it's a mobile app that bridges your Claude Code sessions to your
phone. When Claude needs approval, you get a push notification. Tap Allow/Deny,
session continues.

A few things that differentiate it from Anthropic's built-in remote control:
- Works with Aider and Cline too (shipping in 2 weeks), not just Claude Code
- E2E encrypted — the relay only sees ciphertext, not your session content
- All machines in one dashboard (I run 4 machines)
- Push notifications (Anthropic's version doesn't have these yet)

The relay and CLI agent are MIT licensed: github.com/relixdev/relay

Free tier: 3 machines, 2 sessions — enough to run it for a week and see if you
like it.

TestFlight: [link]
Android beta: [link]

Happy to answer questions. Looking for feedback on the pairing flow specifically —
it works but feels clunky.
```

**What to NOT do on Reddit:**
- Don't post the same message to multiple subreddits simultaneously — it looks spammy
- Don't respond defensively to criticism
- Don't promise features without a ship date you're confident in
- Do respond to every comment within 2 hours for the first 24 hours

**Week 3-4: Hacker News**

Hacker News has a monthly "Ask HN: Who's hiring?" and "Show HN" thread. Show HN is for builders showing real products.

Show HN title: `Show HN: Relix – mobile command center for Claude Code, Aider, Cline`

Post body (max 300 words):
```
I run AI coding agents (Claude Code, Aider) on 4 machines — laptop, two servers,
and a spare desktop. The problem: when I walk away, sessions block waiting for
approvals. And each machine is a separate terminal window.

Relix is a mobile app + relay that fixes this. Install relixctl on your machines
(brew tap or curl install), pair with the app, and you get:

- Dashboard showing all machines and sessions in one view
- Push notifications when any session needs approval
- Inline approve/deny from the notification or in the app
- E2E encrypted (relay sees only ciphertext — self-hostable too)

Claude Code at launch. Aider adapter shipping this week. Cline in 2 weeks.

Architecture: all connections outbound (no port forwarding). X25519 + NaCl box
encryption. The relay and agent are MIT licensed on GitHub.

Free tier: 3 machines, 2 concurrent sessions. Plus is $4.99/month.

Tech: Go for the relay/cloud/agent, React Native (Expo) for the mobile app.

GitHub: github.com/relixdev/relay and github.com/relixdev/relixctl

Happy to answer technical questions about the architecture or encryption.
```

**Twitter/X Strategy**

Post a thread, not a single tweet. Threads get more reach.

Thread format:
1. Hook tweet: "I spent $0 on ads but kept missing Claude Code approvals. So I built something." [screenshot of approval notification on phone]
2. Tweet 2: "The problem: Claude Code sessions block for hours waiting for your input. Most people don't realize how often this kills their throughput."
3. Tweet 3: Show the architecture diagram. Explain that the relay sees only ciphertext.
4. Tweet 4: "It works with Aider and Cline too. Not just Claude Code." [short GIF]
5. Tweet 5: "The relay and CLI are MIT licensed." [GitHub link] "App is free to try — 3 machines, no credit card."
6. Final tweet: @mention a few relevant accounts (not spammy — only if genuinely relevant)

Target accounts to follow and engage with (not DM spam, just be in the conversation):
- @aiderproject
- @clinedev
- People who tweet about Claude Code workflows

**Discord Strategy**

Target servers where AI developers hang out:
1. Anthropic Discord (if they have one)
2. Aider Discord: https://discord.gg/aider
3. Cline Discord
4. Local LLaMA Discord
5. Your own Relix Discord (invite beta users here for support)

Do not post promotional messages. Engage in `#tools` or `#showcase` channels only when there's a dedicated place for it. Answer questions about Claude Code that naturally lead to mentioning Relix.

### Phase 2: Product Hunt Launch

Product Hunt launch should happen after Phase 1 — you want existing users who can upvote and leave genuine reviews.

**Pre-launch preparation:**
1. Create a Maker account on Product Hunt
2. "Coming soon" page at producthunt.com (collect followers pre-launch)
3. Reach out to 5 "hunters" (established PH users) to ask if they'll hunt the product
4. Prepare all assets: tagline, description, screenshots, demo video, first comment

**Timing:**
- Launch on Tuesday, Wednesday, or Thursday — highest traffic days
- Launch at 12:01 AM Pacific Time — gives maximum exposure window

**Product Hunt listing:**

Tagline: "Mobile command center for all your AI coding agents"

Description:
```
Relix lets you control Claude Code, Aider, Cline, and other AI coding agents
from your phone.

- Push notifications when any agent needs approval
- Dashboard showing all your machines in one view
- Approve/deny tool use directly from your phone
- E2E encrypted — relay sees only ciphertext

Free tier: 3 machines, 2 sessions. Plus at $4.99/month.

The relay and CLI agent are open source (MIT) on GitHub.
```

First comment (post immediately when it goes live):
```
Hi Product Hunt! I built Relix to solve a problem I kept hitting: running Claude
Code on multiple machines and missing approval requests when I stepped away.

Happy to answer questions about the architecture (it uses X25519 + NaCl box
encryption — the relay truly cannot read your session content), the multi-tool
support, or anything else.

The relay is self-hostable if you don't want to use our hosted version.

For the developers here: would love to know which AI coding tool you use most —
it helps us prioritize adapters.
```

**Day-of execution:**
- Post first comment immediately at launch
- Reply to every comment within 15 minutes for the first 3 hours
- Share on Twitter/X, Reddit, Discord at launch time
- Ask existing users to visit and leave an honest review (not just upvote)

---

## 5. Content Strategy

### SEO Keywords to Target

Primary (high intent, moderate competition):
- "control claude code from phone"
- "claude code remote control alternative"
- "ai coding agent mobile app"
- "aider mobile remote control"

Secondary (informational, build authority):
- "claude code hooks tutorial"
- "aider workflow tips"
- "ai coding agent approval workflow"
- "e2e encrypted developer tools"

Long-tail (low competition, high conversion):
- "control multiple claude code sessions"
- "claude code push notifications phone"
- "ai agent approval mobile"

### Blog Posts to Write

Post these on the Relix blog (a simple markdown blog at `relix.sh/blog`) in this order:

**Post 1 — Launch post (publish on launch day):**
Title: "We built a universal mobile dashboard for AI coding agents"
Content: The problem, the solution, the architecture (with diagrams), a brief comparison to Anthropic Remote Control. 1500 words. Include the architecture diagram. This is the post you share in Show HN.

**Post 2 — Technical deep dive (week 2):**
Title: "How we do E2E encryption for AI coding sessions (and why it matters)"
Content: Explain the X25519 + NaCl box scheme. Show the key exchange flow. Explain what the relay sees vs doesn't see. Explain the SAS (short authentication string) verification. Why this matters for enterprise users. 2000 words. This builds trust with security-conscious developers.

**Post 3 — Aider adapter deep dive (week 3, publish when Aider ships):**
Title: "Adding Aider to Relix: how we built the adapter in 2 days"
Content: Walkthrough of the `CopilotAdapter` interface. Show how Claude Code and Aider adapters differ. Explain the session discovery problem for each tool. 1200 words. Appeals to open-source contributors and Aider users.

**Post 4 — Case study (week 4-6):**
Title: "Running 4 AI coding agents simultaneously: a workflow"
Content: Show a real developer workflow using Relix. Be specific: "I start three Claude Code sessions, one Aider session, walk to a coffee shop, get two approval notifications on my commute." Include screenshots. 800 words.

**Post 5 — Self-hosting guide:**
Title: "How to self-host the Relix relay on your own server"
Content: Step-by-step with docker-compose. Targets enterprise developers who will not use a hosted relay. Drives word-of-mouth in the enterprise segment.

### Demo Video Script

Runtime target: 90 seconds. No narration needed — just action, overlaid text.

```
0:00 - 0:05  Black screen with text: "It's 6pm. Your AI agent is stuck."
             Show terminal: claude code waiting for approval, time ticking

0:05 - 0:15  Text: "Your phone buzzes."
             Show iPhone lock screen with notification: "Relix: Approval needed —
             Edit src/auth.ts"
             Swipe to open

0:15 - 0:30  Show the approval card in the Relix app
             Tool use details visible: "Edit file: src/auth.ts, lines 42-67"
             Tap "Allow"
             Show Claude Code on laptop continuing immediately

0:30 - 0:45  Text: "Works with every AI tool."
             Quick cuts: Claude Code → Aider → Cline
             Each showing a session in the Relix dashboard

0:45 - 0:60  Text: "All your machines, one dashboard."
             Show dashboard with 3 machines, different statuses
             One has "waiting for approval" badge
             Tap it, approve inline from dashboard

1:00 - 1:15  Text: "E2E encrypted. Open source relay."
             Show the GitHub page briefly
             Show the "relay sees only ciphertext" diagram

1:15 - 1:30  Text: "Free to start. 3 machines included."
             App Store badge + Play Store badge
             "relix.sh"
```

Upload to: YouTube (embed on landing page), X/Twitter (native video gets more reach), and as a GIF for Reddit posts.

### GitHub README for Open Source Repos

The README for `github.com/relixdev/relay` and `github.com/relixdev/relixctl` should:

1. Lead with "What this is" in one sentence
2. Show a quick start in the first 20 lines
3. Link to `relix.sh` prominently
4. Explain the architecture and security model
5. Have a "Self-hosting" section
6. Include the `CopilotAdapter` interface documentation (for relixctl)
7. License badge: MIT

A good README drives both trust and installs. Developers who star the relay repo are leads.

---

## 6. Growth Levers

### 1. Free Tier as Growth Engine

The free tier is not a compromise — it is a deliberate acquisition strategy. Three machines and two sessions is enough to:
- Install relixctl and pair a machine (15 minutes)
- Run Claude Code from the phone on a real project
- Experience the value proposition

Once a developer has approved a tool use from their phone while walking to a coffee machine, they will pay $4.99 without hesitation. The free tier creates the "aha moment" before asking for money.

**Do not reduce the free tier.** If conversion is low, the problem is the onboarding flow or the product, not the free tier being too generous.

### 2. Open Source Relay Builds Trust

Enterprises and security-conscious developers will not use a tool that routes their code through a third-party server they cannot inspect. The MIT-licensed relay removes this objection:

- Enterprise developer: "Can I audit the code?" → Yes, here's GitHub.
- Enterprise developer: "Can I run it on my own infra?" → Yes, `docker run ghcr.io/relixdev/relay`.

This is the Bitwarden model. It works. Open source relay + proprietary cloud drives both adoption and enterprise sales.

### 3. CLI Install Drives Organic Discovery

The Homebrew tap and curl install script will be discovered by developers searching for "relixctl install" or browsing the GitHub repo. Every `brew install relixdev/tap/relixctl` is an organic acquisition.

Make the install experience exceptional:
- `relixctl login` should work in under 60 seconds
- `relixctl pair <code>` should work in under 30 seconds
- The daemon should auto-start and stay running

If install is hard, word-of-mouth dies.

### 4. Word of Mouth via Shared Sessions

The Team tier's shared sessions feature lets multiple people see and approve sessions. This has a built-in network effect: one Team subscriber invites colleagues, colleagues experience the product, some become their own subscribers.

Amplify this: when an invited collaborator joins a session, show them a prompt to create their own account. This is the Slack/Figma "invited user" model.

### 5. Self-Hosted Option for Enterprise

Enterprise developers who self-host the relay still need the mobile app and cloud (for auth, billing, push notifications). They are paying customers. The self-host option is not revenue-loss — it is an enterprise acquisition channel.

When an enterprise self-hosts, the relay is free but the mobile app requires a cloud account. Offer an "Enterprise" tier (custom pricing, SSO, audit logs) for teams that want to self-host and keep all data on-premise.

---

## 7. Conversion Funnel

### Stages

```
Awareness     → discovers Relix (Reddit, HN, Twitter, word of mouth)
Install       → downloads app from App Store or Play Store
Pair          → installs relixctl and successfully pairs a machine
Daily Use     → opens app at least 3 days in a week
Hit Limit     → free tier limit reached (3 machines or 2 sessions)
Upgrade       → converts to Plus ($4.99)
```

### Reducing Friction at Each Stage

**Awareness → Install:**
- Landing page must load in <2 seconds
- Demo video must be above the fold
- App Store / Play Store badges must be visible without scrolling

**Install → Pair:**
This is the hardest step. The user must: install the app, install relixctl, run `relixctl login`, then `relixctl pair <code>`.

Reduce friction:
- Onboarding screen shows the exact commands to run, copy-button included
- Polling: app polls for pairing completion so user doesn't have to re-enter the code
- Error recovery: if pairing times out, show "Code expired — here's a new one" automatically

**Pair → Daily Use:**
- Immediately after pairing, show a "What to do next" screen with one action: "Start a Claude Code session and walk away"
- Push notifications are critical here — if the first notification fires correctly, the user has experienced the value proposition

**Daily Use → Hit Limit:**
- Track when a user registers machine 3 (approaching limit) or session 2
- Show a soft limit warning: "You're using 2/3 machines. Plus includes 10."

**Hit Limit → Upgrade:**
- When a 4th machine registration is attempted: intercept with an upgrade prompt
- Do not hard-block silently — explain the limit and show the upgrade option
- CTA: "Upgrade to Plus — $4.99/month" with one-tap Stripe checkout
- Offer 7-day free trial on first upgrade (reduces activation friction)

### Email Sequences

**Sequence 1: Welcome (trigger: account created)**

Email 1 (immediately): Subject: "Welcome to Relix — here's your first step"
```
You're in.

One thing to do right now:

  brew install relixdev/tap/relixctl
  relixctl login

Takes about 2 minutes. Once you pair your first machine, you'll get a push
notification the next time Claude Code needs your approval.

That's the moment.

— The Relix Team
```

Email 2 (day 3 if no pairing): Subject: "Did you get stuck? Here's what happens most often"
```
We noticed you haven't paired a machine yet.

The most common place people get stuck: relixctl login opens a browser and
then nothing happens.

Fix: if you're on a headless server, use:
  relixctl login --code

This gives you a URL + code to authenticate without a browser.

If you hit something else, reply to this email. We read everything.
```

Email 3 (day 7 if no pairing): Subject: "Quick question"
```
Haven't heard from you. Did Relix work?

If not — what happened? One sentence is enough. It helps us fix the thing
that's blocking you (and probably others).
```

**Sequence 2: Engaged User (trigger: first approval sent from phone)**

Email 1 (1 hour after first approval): Subject: "That just worked"
```
You just approved a tool use from your phone.

That means the agent is running, the relay is routing, the encryption is
working, and the push notification fired.

Worth knowing: you can also approve directly from the notification without
opening the app. Swipe left on the notification → Allow or Deny.

Next: try the session view. Tap the machine, then the session. You can read
the full conversation and send new messages from your phone.
```

**Sequence 3: Approaching Limit (trigger: 3rd machine registered)**

Email 1 (same day): Subject: "You're using Relix on 3 machines"
```
You've paired your third machine. That's the free tier limit.

When you need a fourth:

  Plus — $4.99/month — 10 machines, 5 sessions, 30-day history

If you've been using Relix for a week and it's been useful, Plus is probably
worth it. If you're still evaluating, keep the free tier as long as you need.

Upgrade: relix.sh/upgrade
```

---

## 8. Pricing Psychology

### Why $4.99 for Plus

- Sub-$5 is an impulse buy. No approval process. No "should I do this?" The mental threshold for "$4.99/month" is "a coffee." Developers who use AI coding agents are spending hundreds of dollars on API fees. $4.99 is noise.
- Stripe shows that conversion rates drop significantly above $5 for B2C developer tools. $4.99 captures the price-conscious segment. $5.00 triggers a comparison process.
- If the free tier creates the "aha moment," Plus closes at $4.99 without a sales conversation.

### Why the Free Tier is Generous

3 machines and 2 concurrent sessions is enough for:
- One developer laptop
- One server
- One spare machine

That is a representative setup for many developers. The free tier does not feel limited — it feels complete. This is intentional.

A stingy free tier (1 machine, 1 session) would reduce adoption without meaningfully increasing conversion. Developers talk to each other. "Free tier is too limited" spreads faster than "it's only $4.99."

### Why the Annual Discount Exists

Annual pricing at 10 months' cost (2 months free) accomplishes:
- Reduces churn (annual subscribers don't re-evaluate monthly)
- Improves cash flow
- Feels like a deal to value-conscious users

Annual Plus: $49.90/year (vs $59.88/year monthly = 2 months free)
Annual Pro: $149.90/year (vs $179.88/year monthly = 2 months free)

### Pro vs Plus Positioning

Pro ($14.99/month) is for:
- Developers with >10 machines (rare but intense users — data scientists, ML engineers with many servers)
- Developers who want 90-day history (important for long-running projects)
- Developers who value "priority relay" (lower latency routing for sessions on dedicated relay capacity)

The jump from $4.99 to $14.99 is 3x. This is intentional — Pro is for power users, not for developers who just hit the 10-machine limit. The positioning should emphasize unlimited machines, 90-day history, and priority relay rather than just machine count.

### Team Tier is Inbound Only

$24.99/user/month with shared sessions and admin/SSO is a B2B product. It will not sell via the app stores. It sells when an engineering manager at a company with 10+ developers sees that their team is all using Relix individually.

The path to Team sales:
1. Individual developer uses Relix
2. Shows it to their team in Slack
3. Engineering manager asks "can we get this for everyone with SSO?"
4. They find `relix.sh/enterprise` or email `sales@relix.sh`
5. Relix quotes Team tier pricing

Do not add a "contact sales" form until you have Team tier demand. The first few enterprise sales should come through direct email.

---

## 9. Metrics to Track

### Day 1 Metrics (Track from First User)

Set up these events in your analytics system (Posthog or Mixpanel — both have free tiers):

| Event | Trigger | Why It Matters |
|-------|---------|----------------|
| `account_created` | User completes signup | Top of funnel |
| `machine_paired` | First successful pairing | "Installed" milestone |
| `session_viewed` | User opens a session view | Engagement signal |
| `approval_sent` | User approves or denies a tool use | Core value delivered |
| `push_notification_received` | Push delivered to device | Infrastructure health |
| `push_notification_opened` | User opened app from push | Push effectiveness |
| `machine_limit_hit` | Registration of N+1 machine rejected | Upgrade trigger |
| `upgrade_shown` | Upgrade prompt displayed | Funnel stage |
| `checkout_started` | Stripe checkout initiated | Purchase intent |
| `subscription_created` | Stripe webhook confirmed | Revenue |
| `subscription_cancelled` | Stripe webhook confirmed | Churn |

### Key Metrics by Stage

**Growth:**
- Weekly installs (App Store Connect + Play Console analytics)
- Weekly new accounts
- Weekly new pairings (unique machines paired)

**Engagement:**
- DAU / WAU / MAU
- Sessions bridged per DAU (how often do users actually use it?)
- Approvals sent per session (are sessions reaching approval events?)
- D1 retention: % of users who open the app on day 2
- D7 retention: % of users still active at day 7
- D30 retention: % of users still active at day 30

**Conversion:**
- Free → Paid conversion rate (target: >5% within 30 days for active users)
- Upgrade prompt → Checkout started rate (measures prompt effectiveness)
- Checkout started → Subscription created rate (measures Stripe friction)

**Revenue:**
- MRR (monthly recurring revenue)
- ARPU (average revenue per user)
- Churn rate (% of paid subscribers who cancel per month — target: <5%)
- LTV (lifetime value = ARPU / churn rate)

**Infrastructure:**
- Relay: active WebSocket connections
- Relay: messages routed per minute
- Relay: buffer depth (how many messages are queued for offline clients)
- Cloud: `/health` latency p50/p99
- Push delivery rate (sent vs delivered — APNs and FCM report this)

### NPS

Send an NPS survey via email 14 days after first pairing (when the user has had time to form an opinion):

```
Subject: Quick question about Relix

How likely are you to recommend Relix to a developer friend?
[0-10 scale]

What's the main reason for your score?
[text field]
```

Detractors (0-6): reply personally, understand the specific problem, fix it if it's a bug
Passives (7-8): ask what one thing would make them a promoter
Promoters (9-10): ask them to share in a relevant community or leave an App Store review

Target NPS > 50 before scaling marketing. Below 50 means the product has fundamental issues that marketing budget will not fix.

### Weekly Review (For AI Operator)

Every Monday, review:
1. New installs vs last week (is the number going up?)
2. Pairing completion rate (installs that led to a paired machine — target: >60%)
3. D7 retention (did users return? Target: >30%)
4. Any support emails (common issues = product bugs)
5. Any significant Reddit/Twitter mentions
6. Revenue change (MRR delta)

If pairing completion rate drops below 50%, investigate the onboarding flow immediately — something is broken.

If D7 retention drops below 20%, investigate why users are not returning — either they are not running AI coding agents frequently, or the tool is not surfacing at the right moment.
