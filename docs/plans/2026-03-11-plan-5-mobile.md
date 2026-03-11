# Plan 5: Relix Mobile

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the React Native mobile app (iOS + Android) with dashboard home, chat/terminal session views, push notifications, and biometric auth.

**Architecture:** React Native with Expo (managed workflow). TypeScript throughout. Zustand for state management. React Navigation for routing. libsodium-wrappers for E2E decryption. WebSocket connection to relay for real-time events.

**Tech Stack:** React Native (Expo SDK 52+), TypeScript, zustand, react-navigation, expo-notifications, expo-local-authentication, libsodium-wrappers

**Spec:** `docs/specs/2026-03-11-relix-design.md`
**Depends on:** Plan 1 (protocol — JSON schemas), Plan 3 (relay running), Plan 4 (cloud running)

---

## File Structure

```
mobile/
├── app.json                   # Expo config
├── package.json
├── tsconfig.json
├── app/
│   ├── _layout.tsx            # Root layout (navigation, auth provider)
│   ├── (auth)/
│   │   ├── login.tsx          # Login screen (GitHub OAuth + email)
│   │   └── onboarding.tsx     # Post-signup: install agent instructions
│   ├── (tabs)/
│   │   ├── _layout.tsx        # Tab navigator (Home, New, Settings)
│   │   ├── index.tsx          # Dashboard home screen
│   │   ├── new.tsx            # Start new session
│   │   └── settings.tsx       # Settings (account, machines, preferences)
│   ├── session/
│   │   └── [id].tsx           # Session view (chat/terminal toggle)
│   └── pairing.tsx            # Pairing flow (show code, verify SAS)
├── components/
│   ├── ApprovalCard.tsx       # Inline approve/deny card
│   ├── ChatMessage.tsx        # Chat bubble (user + assistant)
│   ├── TerminalView.tsx       # Monospace terminal-style renderer
│   ├── MachineCard.tsx        # Machine row in dashboard
│   ├── ToolCallCard.tsx       # Expandable file edit / command card
│   ├── SessionToggle.tsx      # Chat ↔ Terminal mode toggle
│   └── SASVerification.tsx    # 4-emoji visual verification
├── lib/
│   ├── api.ts                 # Relix Cloud API client (REST)
│   ├── relay.ts               # WebSocket connection to relay
│   ├── crypto.ts              # libsodium NaCl box encrypt/decrypt
│   ├── protocol.ts            # TypeScript types matching Go protocol module
│   ├── auth.ts                # Token storage (SecureStore), refresh logic
│   └── notifications.ts       # Push notification registration + handling
├── stores/
│   ├── authStore.ts           # User auth state
│   ├── machineStore.ts        # Connected machines + status
│   └── sessionStore.ts        # Active session events + messages
└── eas.json                   # EAS Build config
```

---

## Task Outline

### Chunk 1: Project Setup & Protocol Types (Tasks 1-4)

- [ ] **Task 1:** Init Expo project with TypeScript template, add dependencies
- [ ] **Task 2:** TypeScript protocol types — mirror Go protocol module (Envelope, Payload, Session, Event, MessageType, etc.)
- [ ] **Task 3:** Crypto module — libsodium-wrappers: generateKeyPair, encrypt, decrypt, sealPayload, openPayload
- [ ] **Task 4:** Crypto tests — round-trip encrypt/decrypt, verify interop with Go (test vectors)

### Chunk 2: Auth & Onboarding (Tasks 5-8)

- [ ] **Task 5:** Auth store — zustand store for user, JWT, refresh token (expo-secure-store)
- [ ] **Task 6:** API client — login (GitHub OAuth via AuthSession), token refresh, machine list
- [ ] **Task 7:** Login screen — GitHub OAuth button, email login form
- [ ] **Task 8:** Onboarding screen — show `brew install` / `curl` instructions, poll for first machine connection

### Chunk 3: Dashboard (Tasks 9-12)

- [ ] **Task 9:** Machine store — zustand store for machines list, online/offline status, active sessions
- [ ] **Task 10:** WebSocket relay connection — connect, authenticate, handle reconnection, parse Envelopes
- [ ] **Task 11:** Dashboard screen — machine list with status indicators, pending approval cards at top
- [ ] **Task 12:** Inline approval — Allow/Deny buttons on dashboard, send approval_response via relay

### Chunk 4: Session View — Chat Mode (Tasks 13-16)

- [ ] **Task 13:** Session store — zustand store for events in current session, decrypt payloads
- [ ] **Task 14:** Chat message components — user bubble, assistant bubble, tool call card, approval card
- [ ] **Task 15:** Chat session screen — FlatList of messages, auto-scroll, text input at bottom
- [ ] **Task 16:** Send message — encrypt user input, send via relay, optimistic UI update

### Chunk 5: Session View — Terminal Mode (Tasks 17-19)

- [ ] **Task 17:** Terminal renderer — monospace text view, ANSI-to-styled-text conversion (basic subset)
- [ ] **Task 18:** Terminal session screen — scrollable terminal output, approve/deny as terminal-style buttons
- [ ] **Task 19:** Chat/Terminal toggle — SessionToggle component in header, persists preference

### Chunk 6: Pairing & Push (Tasks 20-23)

- [ ] **Task 20:** Pairing screen — generate code via Cloud API, display 6-digit code, poll for completion
- [ ] **Task 21:** SAS verification — display 4-emoji string, confirm/reject
- [ ] **Task 22:** Push notification registration — expo-notifications, send device token to Cloud API
- [ ] **Task 23:** Push notification handling — tap notification opens relevant session, deep linking

### Chunk 7: Settings & Security (Tasks 24-27)

- [ ] **Task 24:** Settings screen — account info, subscription tier, manage machines, session preferences
- [ ] **Task 25:** Biometric auth — expo-local-authentication, require Face ID/fingerprint on app open
- [ ] **Task 26:** App lock — auto-lock after 5 minutes backgrounded, clear session data from memory
- [ ] **Task 27:** Machine management — rename, revoke (re-pair required), remove

### Chunk 8: Polish & App Store (Tasks 28-31)

- [ ] **Task 28:** Pull-to-refresh on dashboard, loading states, error handling, empty states
- [ ] **Task 29:** App icon, splash screen, App Store screenshots
- [ ] **Task 30:** EAS Build config — iOS + Android builds, signing
- [ ] **Task 31:** TestFlight / Play Store internal testing track deployment
