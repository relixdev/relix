# Relix — Developer Guide

This document explains how to work on the Relix codebase. It is written for an AI agent that has read the codebase and needs to make changes autonomously without asking anyone questions.

---

## Table of Contents

1. [Repository Structure](#1-repository-structure)
2. [Local Development Setup](#2-local-development-setup)
3. [Running Services Locally](#3-running-services-locally)
4. [Running Tests](#4-running-tests)
5. [Go Module Relationships](#5-go-module-relationships)
6. [How to Add a New CopilotAdapter](#6-how-to-add-a-new-copilotadapter)
7. [How to Add a New API Endpoint to Cloud](#7-how-to-add-a-new-api-endpoint)
8. [How to Add a New Screen to Mobile](#8-how-to-add-a-new-screen)
9. [Code Conventions](#9-code-conventions)
10. [Git Workflow](#10-git-workflow)
11. [CI/CD](#11-cicd)

---

## 1. Repository Structure

```
github.com/relixdev/
├── protocol/                    # Shared types and crypto (MIT, Go module)
│   ├── protocol.go              # CopilotAdapter interface, Envelope, Session, Event, UserInput
│   ├── crypto.go                # X25519 key generation, NaCl box encrypt/decrypt
│   └── go.mod                   # module github.com/relixdev/protocol
│
├── relixctl/                    # Agent CLI installed on developer machines (MIT)
│   ├── main.go                  # Entry point — calls cmd.New("").Execute()
│   ├── go.mod                   # module github.com/relixdev/relix/relixctl
│   ├── cmd/                     # Cobra command handlers
│   │   ├── root.go              # Root command, config subcommand, stubCmd helper
│   │   ├── start.go             # relixctl start — installs and starts daemon
│   │   ├── daemonrun.go         # relixctl daemon-run — actual daemon loop (internal)
│   │   ├── status.go            # relixctl status — prints connection state
│   │   ├── sessions.go          # relixctl sessions — lists Claude Code sessions
│   │   └── uninstall.go         # relixctl uninstall — removes service + config
│   └── internal/
│       ├── adapter/             # CopilotAdapter implementations
│       │   ├── claudecode.go    # Claude Code adapter (subprocess + stream-json)
│       │   ├── discovery.go     # Scans ~/.claude/projects/ for session files
│       │   └── integration_test.go
│       ├── auth/                # Agent auth flows
│       │   ├── oauth.go         # GitHub OAuth browser flow
│       │   ├── devicecode.go    # Device code flow (headless servers)
│       │   ├── pairing.go       # 6-digit code pairing with relay
│       │   └── refresh.go       # JWT refresh
│       ├── config/              # relixctl config file (~/.relixctl/config.json)
│       ├── crypto/              # X25519 keystore (~/.relixctl/keys/)
│       ├── daemon/              # Daemon supervisor loop
│       ├── relay/               # Relay WebSocket client, bridge, reconnect logic
│       └── service/             # launchd (macOS) and systemd (Linux) integration
│
├── relay/                       # WebSocket relay server (MIT, Docker)
│   ├── go.mod                   # module github.com/relixdev/relix/relay
│   ├── cmd/relay/main.go        # Entry point
│   └── internal/
│       ├── auth/                # JWT validation for relay connections
│       ├── config/              # Env-driven config (RELAY_JWT_SECRET, etc.)
│       ├── conn/                # WebSocket connection wrapper
│       ├── hub/                 # Core routing logic
│       │   ├── hub.go           # Connection registry, route()
│       │   ├── router.go        # Message dispatch (agent→mobile, mobile→agent)
│       │   ├── buffer.go        # Per-machine offline message buffer
│       │   ├── pairing.go       # 6-digit code generation and validation
│       │   ├── ratelimit.go     # Per-IP rate limiting for pairing
│       │   ├── status.go        # Machine online/offline status tracking
│       │   └── lifecycle.go     # Connection open/close handlers
│       ├── metrics/             # Prometheus metric definitions
│       └── server/              # HTTP server, WebSocket upgrade, pairing HTTP endpoints
│
├── cloud/                       # Auth, billing, push backend (proprietary)
│   ├── go.mod                   # module github.com/relixdev/relix/cloud
│   ├── cmd/cloud/main.go        # Entry point — wires all services together
│   └── internal/
│       ├── api/                 # HTTP handlers and server wiring
│       │   ├── server.go        # Route registration, writeJSON/writeError helpers
│       │   ├── auth_handlers.go # GitHub OAuth, email register/login, refresh
│       │   ├── billing_handlers.go  # GET /billing/plan, POST /billing/checkout
│       │   ├── machine_handlers.go  # CRUD for machine registry
│       │   └── push_handlers.go     # Device token registration + send
│       ├── auth/                # JWT issuance/validation, middleware, email/bcrypt, GitHub OAuth
│       ├── billing/             # Plan definitions, Stripe stub
│       ├── config/              # Env-driven config (JWT_SECRET, DATABASE_URL, etc.)
│       ├── idgen/               # ID generation (prefix + random bytes)
│       ├── machine/             # Machine registry with tier-limit enforcement
│       ├── push/                # APNs and FCM stubs
│       └── user/                # User model, in-memory store (replace with Postgres)
│
└── mobile/                      # React Native app (Expo, proprietary)
    ├── App.tsx                  # Root component (navigation setup)
    ├── index.ts                 # Expo entry point
    ├── app.json                 # Expo config (bundle IDs, icons, etc.)
    ├── package.json             # Dependencies
    ├── tsconfig.json            # Strict TypeScript
    └── src/                     # Application source
        ├── screens/             # One file per screen
        ├── components/          # Shared UI components
        ├── store/               # Zustand state stores
        ├── api/                 # REST client for cloud API
        ├── relay/               # WebSocket client for relay
        ├── crypto/              # libsodium-wrappers NaCl bridge
        └── navigation/          # react-navigation setup
```

---

## 2. Local Development Setup

### Go Services

```bash
# Install Go 1.22+
brew install go

# Verify
go version  # must be >= 1.22

# Install dependencies for each module (done automatically on build/test)
cd relixctl && go mod download
cd ../relay && go mod download
cd ../cloud && go mod download
```

### Mobile App

```bash
cd mobile

# Install Node dependencies
npm install

# Install Expo CLI
npm install -g expo-cli eas-cli

# Verify
npx expo --version
```

### Required Local Environment

Create `cloud/.env.local` for local cloud runs (not committed to git):

```bash
JWT_SECRET=dev-secret-change-in-production-minimum-32-chars
GITHUB_CLIENT_ID=your_dev_github_client_id
GITHUB_CLIENT_SECRET=your_dev_github_client_secret
PORT=8080
```

Create `relay/.env.local`:

```bash
RELAY_JWT_SECRET=dev-secret-change-in-production-minimum-32-chars
RELAY_PORT=8081
```

The JWT secrets in both files must match for agents to authenticate with the relay using tokens issued by cloud.

---

## 3. Running Services Locally

### Run the Relay

```bash
cd relay

# Export env vars
export RELAY_JWT_SECRET="dev-secret-change-in-production-minimum-32-chars"
export RELAY_PORT=8081

# Run
go run ./cmd/relay

# Output:
# relay listening on :8081
```

### Run the Cloud Service

```bash
cd cloud

export JWT_SECRET="dev-secret-change-in-production-minimum-32-chars"
export GITHUB_CLIENT_ID="your_client_id"
export GITHUB_CLIENT_SECRET="your_client_secret"

go run ./cmd/cloud

# Output:
# relix cloud listening on :8080
```

### Run the Agent (relixctl)

```bash
cd relixctl

# Build first (required — daemon-run is invoked as a subprocess)
go build -o ./relixctl .

# Point at local relay and cloud
./relixctl config set relay_url ws://localhost:8081

# Run commands
./relixctl status
./relixctl sessions
```

### Run the Mobile App

```bash
cd mobile

# Start Expo dev server
npx expo start

# Scan QR code with Expo Go app on your phone, or press:
# i — iOS simulator
# a — Android emulator
```

The mobile app's API client needs to point at your local cloud service. Set the API base URL in `src/api/client.ts` to `http://localhost:8080` for local development (or your machine's LAN IP if testing on a real device).

### Run All Services Together (docker-compose)

The relay ships with a `docker-compose.yml` that includes Grafana:

```bash
cd relay
docker compose up

# Prometheus: http://localhost:9090
# Grafana: http://localhost:3000 (admin/admin)
# Relay: ws://localhost:8081
```

---

## 4. Running Tests

### Run All Tests in a Module

```bash
# Protocol
cd protocol && go test ./...

# Relay (includes integration tests)
cd relay && go test ./...

# Cloud
cd cloud && go test ./...

# Agent
cd relixctl && go test ./...
```

### Run Tests with Race Detector

The relay and relixctl have concurrent code. Always run with `-race` to catch data races:

```bash
cd relay && go test -race ./...
cd relixctl && go test -race ./...
```

### Run a Specific Test

```bash
# Run a single test function
cd relay && go test -run TestHubRoute ./internal/hub/

# Run tests matching a pattern
cd cloud && go test -run TestBilling ./internal/billing/
```

### Run Integration Tests

Some packages have `integration_test.go` files that spin up real services:

```bash
# Cloud API integration tests
cd cloud && go test -run TestIntegration ./internal/api/

# Relay integration tests
cd relay && go test -run TestIntegration ./internal/server/

# relixctl adapter integration tests
cd relixctl && go test -run TestIntegration ./internal/adapter/
```

### Mobile Tests

```bash
cd mobile
npm test
# Runs Jest
```

### Test Coverage

```bash
# Generate coverage report for cloud
cd cloud && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

### Before Pushing

Run this checklist:

```bash
# In each Go module:
go vet ./...       # catches common mistakes
go test -race ./... # tests + race detector
staticcheck ./...  # install: go install honnef.co/go/tools/cmd/staticcheck@latest

# Mobile:
cd mobile && npx tsc --noEmit && npm test
```

---

## 5. Go Module Relationships

There are four Go modules in this repo. They are separate modules with `replace` directives pointing to local paths.

```
protocol/          ← no local dependencies
relixctl/          ← depends on protocol (via replace directive)
relay/             ← depends on protocol (via replace directive)
cloud/             ← does NOT depend on protocol (uses its own types)
```

### The `replace` Directive

Each module that depends on `protocol` has this in its `go.mod`:

```go
replace github.com/relixdev/protocol => ../protocol
```

This means `go build` resolves the `protocol` import from the local `../protocol` directory, not from the internet. This is how a monorepo works with multiple Go modules.

**When adding a dependency to `protocol`:** run `go mod tidy` in both `protocol/` and in any module that depends on it (`relixctl/`, `relay/`).

**When publishing protocol as a real module:** remove the `replace` directives and publish `github.com/relixdev/protocol` to GitHub. All dependents must then `go get github.com/relixdev/protocol@<version>`.

### Adding External Dependencies

```bash
# Add to a specific module
cd relay && go get nhooyr.io/websocket@latest

# Tidy after adding
go mod tidy
```

Never run `go get` from the repo root — there is no root `go.mod`.

---

## 6. How to Add a New CopilotAdapter

This is the most common extension point. Use this guide to add Aider, Cline, or any other AI coding tool.

### Step 1: Understand the Interface

The `CopilotAdapter` interface lives in `protocol/protocol.go`:

```go
type CopilotAdapter interface {
    Discover(ctx context.Context) ([]Session, error)
    Attach(ctx context.Context, sessionID string) (<-chan Event, error)
    Send(ctx context.Context, sessionID string, msg UserInput) error
    Detach(sessionID string) error
}
```

- `Discover` — scan the machine for active sessions of this tool
- `Attach` — connect to a specific session; return a channel of events
- `Send` — deliver user input or approval response into the session
- `Detach` — cleanly disconnect

### Step 2: Create the Adapter File

Create `relixctl/internal/adapter/aider.go`:

```go
package adapter

import (
    "context"
    "github.com/relixdev/protocol"
)

// AiderAdapter implements protocol.CopilotAdapter for Aider.
type AiderAdapter struct {
    // fields: working directory scanner, attached sessions, etc.
}

// NewAiderAdapter creates an adapter for Aider sessions.
func NewAiderAdapter() *AiderAdapter {
    return &AiderAdapter{}
}

func (a *AiderAdapter) Discover(ctx context.Context) ([]protocol.Session, error) {
    // Aider does not have session files like Claude Code.
    // Detect running aider processes via process table or PID files.
    // Return a []protocol.Session with Tool="aider".
    panic("not yet implemented")
}

func (a *AiderAdapter) Attach(ctx context.Context, sessionID string) (<-chan protocol.Event, error) {
    // Aider can be run with --input-file and --output-file, or via stdin/stdout.
    // Spawn: aider --input-format <format> and hold stdin open.
    // Read stdout and emit protocol.Events.
    panic("not yet implemented")
}

func (a *AiderAdapter) Send(ctx context.Context, sessionID string, msg protocol.UserInput) error {
    // Write to Aider's stdin pipe.
    panic("not yet implemented")
}

func (a *AiderAdapter) Detach(sessionID string) error {
    // Cancel the context, close stdin pipe.
    panic("not yet implemented")
}
```

### Step 3: Research the Target Tool's Integration Surface

For Aider:
- Run `aider --help` to find stdin/stdout options
- Aider supports `--input-history-file`, `--chat-mode`, and `--no-pretty` for scripting
- The Aider source is at https://github.com/paul-gauthier/aider — read `aider/main.py` and `aider/io.py`

For Cline (VS Code extension):
- Cline is a VS Code extension — integration surface is the VS Code extension API
- Look for: task output, message passing via VS Code commands, or a local HTTP server Cline might expose
- Check: https://github.com/cline/cline

### Step 4: Implement Discover

Look at `relixctl/internal/adapter/discovery.go` for the pattern used by Claude Code (scanning session files). Adapt for your tool:

```go
// For Aider: scan for .aider.chat.history.md files or running processes
func DiscoverAiderSessions() []protocol.Session {
    // Option 1: check for running aider processes
    // Option 2: check for .aider/ directories in common locations
    // Option 3: Aider creates .aider.chat.history.md in the working directory
}
```

### Step 5: Implement Attach

Model it on `ClaudeCodeAdapter.Attach` in `relixctl/internal/adapter/claudecode.go`:

1. Spawn the tool process with `exec.CommandContext`
2. Get stdin and stdout pipes
3. Start a goroutine that reads stdout line-by-line
4. Parse each line into a `protocol.Event`
5. Map the tool's event types to `protocol.PayloadKind` constants
6. Send events to the returned channel

### Step 6: Write Tests

Create `relixctl/internal/adapter/aider_test.go`. Use the same `CommandFactory` injection pattern as the Claude Code adapter to substitute a mock process in tests:

```go
func TestAiderDiscover(t *testing.T) { ... }
func TestAiderAttach(t *testing.T) { ... }
func TestAiderSend(t *testing.T) { ... }
func TestAiderDetach(t *testing.T) { ... }
```

### Step 7: Register the Adapter in the Daemon

In `relixctl/internal/daemon/daemon.go`, the daemon maintains a list of active adapters. Add your new adapter:

```go
adapters := []protocol.CopilotAdapter{
    adapter.NewClaudeCodeAdapter(claudeDir, nil),
    adapter.NewAiderAdapter(),   // add this
}
```

### Step 8: Update the Tool field in Session

Set `Tool: "aider"` in the sessions returned by `Discover`. The mobile app displays this as the session type. The string is passed verbatim to the mobile — the mobile uses it to show an icon or label.

### Adapter Reference: Claude Code

The complete, working Claude Code adapter is at `relixctl/internal/adapter/claudecode.go`. Read it in full before starting — your adapter should follow the same structure:

- `CommandFactory` function type for testability
- `attachedSession` struct for per-session state
- `mapKind()` to translate tool-specific event types to `protocol.PayloadKind`
- Goroutine per session that reads stdout and closes the channel on process exit
- Mutex protecting the `attached` map

---

## 7. How to Add a New API Endpoint to Cloud

Follow this exact pattern — every endpoint in `cloud/` uses it consistently.

### Step 1: Add the Route

Open `cloud/internal/api/server.go`, find the `routes()` method, add your route:

```go
func (s *Server) routes() {
    authMW := auth.Middleware(s.tokens)

    // ... existing routes ...

    // New authenticated endpoint
    s.mux.Handle("PATCH /machines/{id}", authMW(http.HandlerFunc(s.handleRenameMachine)))
}
```

Use Go 1.22 method+path pattern syntax (`"PATCH /machines/{id}"`). Path variables are extracted with `r.PathValue("id")`.

### Step 2: Create the Handler

Add the handler to the appropriate `_handlers.go` file. If no good fit exists, create a new file following the naming pattern `<noun>_handlers.go`.

```go
func (s *Server) handleRenameMachine(w http.ResponseWriter, r *http.Request) {
    // 1. Extract auth context
    userID := auth.UserIDFromContext(r.Context())
    machineID := r.PathValue("id")

    // 2. Parse request body
    var req struct {
        Name string `json:"name"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }
    if req.Name == "" {
        writeError(w, http.StatusBadRequest, "name is required")
        return
    }

    // 3. Call the business logic layer
    m, err := s.registry.Rename(r.Context(), userID, machineID, req.Name)
    if err != nil {
        writeError(w, http.StatusNotFound, err.Error())
        return
    }

    // 4. Write response
    writeJSON(w, http.StatusOK, m)
}
```

### Step 3: Add Business Logic

If the handler needs a new method on an existing service (e.g., `registry.Rename`), add it to the appropriate internal package. Follow the same error-wrapping pattern:

```go
// In cloud/internal/machine/registry.go
func (r *Registry) Rename(ctx context.Context, userID, machineID, name string) (*user.Machine, error) {
    r.mu.Lock()
    defer r.mu.Unlock()

    m, ok := r.machines[machineID]
    if !ok {
        return nil, fmt.Errorf("machine: %q not found", machineID)
    }
    if m.UserID != userID {
        return nil, fmt.Errorf("machine: %q does not belong to user %q", machineID, userID)
    }
    m.Name = name
    return m, nil
}
```

### Step 4: Write Tests

Add a test to `cloud/internal/api/server_test.go` or the relevant `_test.go` file. Use the existing test helpers (look at `server_test.go` for how tests wire up a test server):

```go
func TestHandleRenameMachine(t *testing.T) {
    srv := newTestServer(t)
    token := mustLogin(t, srv)

    // Register a machine first
    m := mustRegisterMachine(t, srv, token, "my-machine")

    // Rename it
    body := `{"name":"renamed-machine"}`
    req := httptest.NewRequest("PATCH", "/machines/"+m.ID, strings.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()
    srv.ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
    }
}
```

### Step 5: Update the API Documentation

Add the new endpoint to `docs/API.md`. Follow the exact format used for existing endpoints.

---

## 8. How to Add a New Screen to Mobile

The mobile app uses Expo Router or react-navigation (check `mobile/src/navigation/`). Follow this pattern.

### Step 1: Create the Screen File

Create `mobile/src/screens/NewScreen.tsx`:

```typescript
import React from 'react';
import { View, Text, StyleSheet } from 'react-native';

interface Props {
  // navigation props injected by react-navigation
}

export function NewScreen({ }: Props) {
  return (
    <View style={styles.container}>
      <Text>New Screen</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#000',
  },
});
```

### Step 2: Register in Navigation

Open `mobile/src/navigation/` and add the screen to the appropriate navigator:

```typescript
// In the relevant Stack.Navigator or Tab.Navigator
<Stack.Screen name="NewScreen" component={NewScreen} />
```

### Step 3: Add Navigation Types

Add the new screen to the navigation type definitions (usually in `mobile/src/navigation/types.ts`):

```typescript
export type RootStackParamList = {
  // ... existing screens ...
  NewScreen: { someParam: string };
};
```

### Step 4: Connect to State

If the screen needs data from the API or relay, use the existing zustand stores in `mobile/src/store/`. Do not fetch data directly in components — go through a store action.

### Step 5: Test

```bash
cd mobile && npm test
```

Also manually test in Expo Go or a simulator before considering it done.

---

## 9. Code Conventions

### Go

**Error handling:** Always wrap errors with context using `fmt.Errorf("package: operation: %w", err)`. The cloud follows `"machine: operation"`, `"auth: operation"` etc. as prefixes. This makes log output self-documenting.

**HTTP handlers:** All handlers follow: decode → validate → call business logic → write response. Never call `writeJSON` before checking errors. Never write a partial response.

**Exported functions:** Have doc comments. Unexported functions do not need comments unless the logic is non-obvious.

**Tests:** Table-driven where there are multiple cases. Use `t.Helper()` in test helpers. Test file names match `*_test.go` in the same package (white-box testing).

**Logging:** Use `log.Printf` in main packages. Internal packages return errors — they do not log. The relay and cloud use `log` from the standard library (no structured logging yet).

**Concurrency:** Protect shared state with `sync.Mutex` or `sync.RWMutex`. Use `sync.RWMutex` when reads are far more frequent than writes (e.g., the machine registry, the relay hub). Use `atomic` types for counters (see `ClaudeCodeAdapter.seq`).

**Config:** All configuration comes from environment variables. No config files for services. Use the `config.Load()` pattern established in each module.

**Stubs:** Stubs are named `Stub<Service>` and implement the same interface as the real implementation. They are in the same package as the interface. They log calls with `[stub]` prefix so you can see them firing in dev.

### TypeScript (Mobile)

**Strict mode:** `tsconfig.json` has `"strict": true`. No `any` types. No non-null assertions (`!`) without a comment explaining why it's safe.

**Component files:** One component per file. File name matches the exported component name. PascalCase for component files, camelCase for everything else.

**State:** Use zustand stores for all shared state. Local component state (`useState`) is fine for UI-only state (e.g., input field value, modal visibility).

**API calls:** All API calls go through `src/api/client.ts`. Never call `fetch` directly from a component.

### General

**No secret values in code or git.** All secrets via environment variables or Fly.io secrets.

**No dead code.** Remove unused functions, imports, and variables before committing.

**No console.log or fmt.Println in production paths.** Use proper logging with `log.Printf`.

---

## 10. Git Workflow

### Branch Strategy

```
main          ← always deployable
feature/*     ← new features (branch from main, merge to main)
fix/*         ← bug fixes (branch from main)
release/vX.Y  ← release preparation (tag from here)
```

### Commit Messages

Follow conventional commits format:

```
feat(adapter): add Aider adapter with stdin/stdout integration
fix(relay): prevent race condition in hub buffer drain
chore(deps): bump golang.org/x/crypto to 0.32.0
docs(api): add /machines/:id PATCH endpoint documentation
test(cloud): add integration test for billing checkout flow
refactor(relay): extract router logic from hub into separate file
```

Format: `type(scope): description`
- Types: `feat`, `fix`, `chore`, `docs`, `test`, `refactor`, `perf`
- Scope: the module or package name
- Description: imperative mood, lowercase, no period

### Tagging Releases

```bash
# Tag relixctl releases (used by Goreleaser for binary distribution)
git tag relixctl/v0.2.0
git push origin relixctl/v0.2.0

# Tag relay releases (used for Docker image tagging)
git tag relay/v0.2.0
git push origin relay/v0.2.0
```

### Before Merging

1. All tests pass: `go test -race ./...` in each module
2. No linter errors: `go vet ./...`
3. `go mod tidy` has been run in any module whose deps changed
4. New public APIs have doc comments
5. Stubs replaced with real implementations are tracked in HANDOFF.md

---

## 11. CI/CD

### GitHub Actions Setup

Create `.github/workflows/` with the following files.

#### `test.yml` — Run tests on every push and PR

```yaml
name: Test

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test-go:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        module: [protocol, relay, cloud, relixctl]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Test ${{ matrix.module }}
        working-directory: ${{ matrix.module }}
        run: go test -race ./...

  test-mobile:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - working-directory: mobile
        run: npm install && npm test
```

#### `deploy-relay.yml` — Deploy relay on push to main

```yaml
name: Deploy Relay

on:
  push:
    branches: [main]
    paths: ['relay/**']

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: superfly/flyctl-actions/setup-flyctl@master
      - run: fly deploy --app relix-relay
        working-directory: relay
        env:
          FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}
```

#### `deploy-cloud.yml` — Deploy cloud on push to main

```yaml
name: Deploy Cloud

on:
  push:
    branches: [main]
    paths: ['cloud/**']

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: superfly/flyctl-actions/setup-flyctl@master
      - run: fly deploy --app relix-cloud
        working-directory: cloud
        env:
          FLY_API_TOKEN: ${{ secrets.FLY_API_TOKEN }}
```

#### `release-relixctl.yml` — Release relixctl binaries with Goreleaser

```yaml
name: Release relixctl

on:
  push:
    tags: ['relixctl/v*']

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - uses: goreleaser/goreleaser-action@v5
        with:
          workdir: relixctl
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

#### Goreleaser Config for relixctl

Create `relixctl/.goreleaser.yaml`:

```yaml
project_name: relixctl
builds:
  - id: relixctl
    main: ./main.go
    binary: relixctl
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w
      - -X main.version={{.Version}}

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip

brews:
  - name: relixctl
    repository:
      owner: relixdev
      name: homebrew-tap
    homepage: https://relix.sh
    description: Relix agent CLI — control AI coding agents from your phone
```

### GitHub Actions Secrets to Configure

Go to repository Settings → Secrets and variables → Actions → New repository secret:

| Secret | Value |
|--------|-------|
| `FLY_API_TOKEN` | From `fly tokens create deploy -x 999999h` |
| `GITHUB_TOKEN` | Automatically available — no action needed |

### EAS Build (Mobile)

EAS builds are triggered manually or via `eas build` command. To automate:

```yaml
# .github/workflows/mobile-build.yml
name: EAS Build

on:
  push:
    branches: [main]
    paths: ['mobile/**']

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - uses: expo/expo-github-action@v8
        with:
          eas-version: latest
          token: ${{ secrets.EXPO_TOKEN }}
      - working-directory: mobile
        run: eas build --platform all --profile production --non-interactive
```

Add `EXPO_TOKEN` secret: create at https://expo.dev/accounts/[username]/settings/access-tokens.
