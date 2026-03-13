# Shadow Memory System

## Philosophy

Shadow operates across multiple projects and Discord channels. Memory must be:
- **Persistent** — survives across conversations and sessions
- **Structured** — not a dump of raw notes, but organized by type and project
- **Searchable** — Shadow can find what it needs in seconds, not minutes
- **Self-maintaining** — Shadow updates memories as things change, prunes stale ones
- **Contextual** — each project/channel gets its own memory space, with a global layer on top

## Directory Structure

```
memory/
├── global/                          # Cross-project knowledge
│   ├── zach.md                      # Who Zach is, preferences, communication style
│   ├── decisions.md                 # Major decisions with date + reasoning
│   ├── accounts.md                  # Service accounts, what's set up, what's not
│   ├── budget.md                    # Spending tracker, approved budgets
│   └── learnings.md                 # Hard-won lessons (things that failed, things that worked)
│
├── relix/                           # Relix project memory
│   ├── status.md                    # SINGLE SOURCE OF TRUTH for current project state
│   ├── decisions.md                 # Architecture and product decisions with reasoning
│   ├── blockers.md                  # Current blockers and what's needed to unblock
│   ├── users.md                     # Beta users, feedback, quotes, feature requests
│   ├── metrics.md                   # KPIs, conversion rates, revenue, costs
│   ├── infrastructure.md            # What's deployed where, URLs, connection strings
│   ├── releases.md                  # Version history, what shipped when
│   ├── incidents.md                 # Outages, bugs, postmortems
│   └── backlog.md                   # Prioritized list of what to build next
│
├── [other-project]/                 # Same structure per project
│   ├── status.md
│   ├── decisions.md
│   └── ...
│
└── discord/                         # Discord-channel-specific context
    ├── general.md                   # What gets discussed in #general
    ├── relix-dev.md                 # Dev channel context, ongoing threads
    ├── relix-support.md             # User support patterns, common issues
    └── standup.md                   # Daily standup history (last 7 days)
```

## File Formats

### status.md (most important file — read this every session)

```markdown
# [Project] Status — Updated [date]

## Current Phase
[One line: what phase are we in]

## This Week's Focus
- [ ] Task 1 — [status]
- [ ] Task 2 — [status]
- [x] Task 3 — done [date]

## Blockers
- [Blocker] — waiting on [what] from [who] since [date]

## Recently Completed
- [date] — [what was done]
- [date] — [what was done]

## Key Numbers
- Users: X
- Revenue: $X MRR
- Uptime: X%

## Next Actions (when current focus is done)
1. [Next thing]
2. [Next thing]
```

### decisions.md

```markdown
# [Project] Decisions Log

## [date] — [Decision title]
**Context:** [Why this came up]
**Options considered:** [What alternatives existed]
**Decision:** [What was decided]
**Reasoning:** [Why]
**Approved by:** [Zach / Shadow autonomously]
**Status:** [Active / Superseded by X]
```

### learnings.md

```markdown
# Learnings

## [date] — [Title]
**What happened:** [Description]
**What we learned:** [Lesson]
**What to do differently:** [Action]
```

## Daily Routine

Every day (or every new session), Shadow should:

### 1. Morning Read (30 seconds)
```
Read: memory/relix/status.md
Read: memory/relix/blockers.md
Read: memory/global/budget.md (if spending today)
```

### 2. Work
Execute against the current focus in status.md.

### 3. End-of-Day Write (60 seconds)
Update these files with anything that changed:
- `status.md` — update task checkboxes, add completed items
- `blockers.md` — add new blockers, remove resolved ones
- `decisions.md` — log any decisions made today
- `learnings.md` — anything surprising or worth remembering
- `metrics.md` — any new numbers

### 4. Weekly Review (Fridays)
- Archive completed tasks from status.md
- Review blockers — escalate anything stuck > 3 days
- Update metrics.md with weekly numbers
- Post standup summary to Discord
- Prune any stale memories

## Memory Rules

### What to Remember
- Decisions and their reasoning (you WILL forget why you did something)
- User feedback (exact quotes, not summaries)
- Things that broke and how they were fixed
- Approval grants from Zach (so you never ask twice)
- Infrastructure details (URLs, connection strings, service configs)
- Spending (what costs what, running total)

### What NOT to Remember
- Raw code (it's in git)
- Anything that can be derived by reading the codebase
- Conversation transcripts (too noisy)
- Temporary debugging notes
- Implementation details (they change; the decisions behind them don't)

### Memory Hygiene
- **status.md is always current** — if it's wrong, nothing else matters
- **Decisions are permanent** — never delete a decision, mark it superseded
- **Metrics are append-only** — add new rows, don't overwrite old ones
- **Prune weekly** — remove resolved blockers, archive old standup entries
- **Date everything** — every entry gets a date so you know how fresh it is

## Bootstrap Instructions

When starting for the first time on a new project:

1. Create the memory directory structure above
2. Read all project documentation (CLAUDE.md, HANDOFF.md, etc.)
3. Populate `status.md` with current state from HANDOFF.md
4. Populate `decisions.md` with architecture decisions from the design spec
5. Populate `infrastructure.md` with what exists (nothing deployed yet for Relix)
6. Populate `accounts.md` with what services are set up
7. Set your first week's focus in `status.md`
8. Start executing

## Integration with Discord

Each Discord channel maps to a memory context:

| Channel | Memory File | What Shadow Tracks |
|---------|------------|-------------------|
| #general | discord/general.md | Zach's mood, priorities, random context |
| #relix-dev | discord/relix-dev.md | Technical discussions, decisions made in chat |
| #relix-support | discord/relix-support.md | User issues, patterns, common fixes |
| #standup | discord/standup.md | What Shadow did today, plans for tomorrow |

When Zach says something important in Discord, Shadow should:
1. Extract the key information
2. Write it to the appropriate memory file
3. If it's a decision or approval: add to decisions.md
4. If it changes priorities: update status.md
5. Acknowledge in Discord: "Got it, updated [file]"

## Example: Relix Bootstrap

After reading this document, Shadow's first actions should be:

```
1. Create memory/relix/status.md:
   - Phase: Pre-deployment
   - Focus: Steps 1-5 of critical path (domain, DNS, PostgreSQL, deploy)
   - Blockers: Need Cloudflare/Fly.io credentials from Zach

2. Create memory/relix/decisions.md:
   - Copy key decisions from design spec (encryption, protocol, Go, Expo, etc.)

3. Create memory/relix/infrastructure.md:
   - GitHub: github.com/relixdev/relix ✅
   - Cloudflare: ❌ not set up
   - Fly.io: ❌ not set up
   - Stripe: ❌ not set up
   - App Store: ❌ not set up

4. Create memory/global/accounts.md:
   - List all services needed with setup status

5. Message Zach on Discord:
   "Memory bootstrapped. Ready to start. I need Cloudflare and Fly.io
   credentials to begin the critical path. Can you set those up?"

6. While waiting: work on items that don't need credentials
   (PostgreSQL migration code, Stripe integration code, tests)
```
