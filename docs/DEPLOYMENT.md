# Relix — Deployment Guide

This document covers everything required to deploy Relix from zero to a production environment. Every step is actionable. No human assistance is assumed.

---

## Table of Contents

1. [Prerequisites — Accounts and Tools](#1-prerequisites)
2. [Environment Variables Reference](#2-environment-variables)
3. [PostgreSQL Setup](#3-postgresql-setup)
4. [Redis Setup](#4-redis-setup)
5. [Relay Deployment to Fly.io](#5-relay-deployment)
6. [Cloud Deployment to Fly.io](#6-cloud-deployment)
7. [Domain and DNS Setup](#7-domain-and-dns)
8. [SSL/TLS](#8-ssltls)
9. [Mobile App Deployment (EAS)](#9-mobile-app-deployment)
10. [Push Notification Setup](#10-push-notification-setup)
11. [Stripe Setup](#11-stripe-setup)
12. [Monitoring](#12-monitoring)
13. [Backup and Disaster Recovery](#13-backup-and-disaster-recovery)
14. [Cost Estimates](#14-cost-estimates)

---

## 1. Prerequisites

### Accounts Required

| Service | URL | Purpose |
|---------|-----|---------|
| Fly.io | https://fly.io | Hosts relay and cloud services |
| Stripe | https://stripe.com | Billing and subscriptions |
| Apple Developer | https://developer.apple.com | iOS app distribution + APNs |
| Google Play Console | https://play.google.com/console | Android app distribution + FCM |
| Firebase (Google) | https://console.firebase.google.com | FCM push notifications for Android |
| GitHub | https://github.com | OAuth provider, source hosting |
| Expo / EAS | https://expo.dev | Mobile build pipeline |
| Domain Registrar | (any — Cloudflare recommended) | relix.sh domain |

### Local Tooling Required

```bash
# Fly CLI
brew install flyctl
# or: curl -L https://fly.io/install.sh | sh

# EAS CLI
npm install -g eas-cli

# Go 1.22+ (for local builds and testing)
brew install go

# Docker (for local relay testing)
brew install --cask docker

# Verify
flyctl version
eas --version
go version
docker --version
```

### GitHub OAuth App Setup

You need a GitHub OAuth App to support "Login with GitHub" in the mobile app and agent.

1. Go to https://github.com/settings/developers → "OAuth Apps" → "New OAuth App"
2. Application name: `Relix`
3. Homepage URL: `https://relix.sh`
4. Authorization callback URL: `https://api.relix.sh/auth/github/callback`
5. Click "Register application"
6. Note down: **Client ID** and **Client Secret**
7. Create a second OAuth App for development with callback `http://localhost:8080/auth/github/callback`

---

## 2. Environment Variables

### Relay Service

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `RELAY_JWT_SECRET` | YES | — | HMAC secret used to validate JWT tokens from agents and mobile clients. Must match `JWT_SECRET` in cloud service. Generate with: `openssl rand -hex 32` |
| `RELAY_PORT` | no | `8080` | HTTP listen port. Fly.io handles TLS termination; the app listens on plain HTTP internally. |
| `RELAY_BUFFER_MAX_MESSAGES` | no | `1000` | Maximum messages to buffer per machine when mobile is disconnected |
| `RELAY_BUFFER_MAX_BYTES` | no | `10485760` | Maximum buffer size in bytes per machine (10MB default) |
| `RELAY_BUFFER_TTL` | no | `24h` | How long buffered messages are retained. Go duration format: `24h`, `30m`, etc. |

### Cloud Service

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | YES | — | HMAC secret for signing user JWTs. Must be at least 32 random bytes. Must match `RELAY_JWT_SECRET` so agents can authenticate with the relay using tokens issued by cloud. Generate with: `openssl rand -hex 32` |
| `DATABASE_URL` | no | — | PostgreSQL connection string. Format: `postgres://user:pass@host:5432/dbname?sslmode=require`. Currently the cloud uses an in-memory store (stub); set this when replacing with a real PostgreSQL store. |
| `REDIS_URL` | no | — | Redis connection string. Format: `redis://user:pass@host:6379`. Currently unused (stub); set when implementing rate limiting or session caching. |
| `GITHUB_CLIENT_ID` | no | — | GitHub OAuth App Client ID. Required for GitHub login to work. |
| `GITHUB_CLIENT_SECRET` | no | — | GitHub OAuth App Client Secret. |
| `STRIPE_SECRET_KEY` | no | — | Stripe secret key (`sk_live_...` for production, `sk_test_...` for testing). Required to replace the billing stub. |
| `PORT` | no | `8080` | HTTP listen port. |

### Generating Secrets

```bash
# JWT secret (use the same value for both RELAY_JWT_SECRET and JWT_SECRET)
openssl rand -hex 32

# Example output: a3f7b2c1d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1
```

Store all secrets in a password manager (1Password, Bitwarden) immediately after generating. They cannot be recovered if lost.

---

## 3. PostgreSQL Setup

### Option A: Fly Postgres (Recommended for Launch)

```bash
# Create a Postgres cluster on Fly.io
fly postgres create \
  --name relix-postgres \
  --region iad \
  --initial-cluster-size 1 \
  --vm-size shared-cpu-1x \
  --volume-size 10

# Note the connection string printed — save it immediately.
# Format: postgres://relix:<password>@relix-postgres.internal:5432/relix

# Attach to the cloud app (creates DATABASE_URL secret automatically)
fly postgres attach relix-postgres --app relix-cloud
```

### Option B: Supabase (Managed, No Ops)

1. Create account at https://supabase.com
2. New project → choose region closest to your Fly.io region
3. Settings → Database → "Connection string" (URI format)
4. Add `?sslmode=require` to the end
5. Set as `DATABASE_URL` secret in Fly.io cloud app

### Database Migrations

The cloud service currently uses an in-memory user store (`user.NewMemoryStore()`). When you replace it with PostgreSQL, you need to add a migration system. Use golang-migrate:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Create `cloud/migrations/` directory with numbered SQL files:

```sql
-- 001_create_users.up.sql
CREATE TABLE users (
    id          TEXT PRIMARY KEY,
    email       TEXT UNIQUE,
    github_id   TEXT UNIQUE,
    tier        TEXT NOT NULL DEFAULT 'free',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 002_create_machines.up.sql
CREATE TABLE machines (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    public_key  TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 003_create_device_tokens.up.sql
CREATE TABLE device_tokens (
    id           TEXT PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_token TEXT NOT NULL,
    platform     TEXT NOT NULL, -- 'apns' or 'fcm'
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, device_token)
);
```

Run migrations:
```bash
migrate -path cloud/migrations -database "$DATABASE_URL" up
```

---

## 4. Redis Setup

### Option A: Fly Redis (Upstash)

```bash
fly redis create \
  --name relix-redis \
  --region iad \
  --no-replicas

# Attach to cloud app
fly redis attach relix-redis --app relix-cloud
# This sets REDIS_URL secret automatically
```

### Option B: Upstash Directly

1. https://console.upstash.com → "Create Database"
2. Choose region matching your Fly.io deployment
3. Copy the Redis URL (format: `rediss://default:password@host:port`)
4. Set as `REDIS_URL` secret

Redis is not yet actively used in the codebase (the cloud config reads it but no code exercises it). It is wired in for future use as a rate-limiting and session-caching layer. You can defer this until you implement those features.

---

## 5. Relay Deployment

### Fly.io Setup

```bash
cd /path/to/relay

# Initialize Fly app (first time only)
fly launch --name relix-relay --region iad --no-deploy
```

### fly.toml for Relay

Create `/path/to/relay/fly.toml`:

```toml
app = "relix-relay"
primary_region = "iad"

[build]
  dockerfile = "Dockerfile"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = false
  auto_start_machines = true
  min_machines_running = 1

  [http_service.concurrency]
    type = "connections"
    hard_limit = 5000
    soft_limit = 4000

[[vm]]
  size = "shared-cpu-1x"
  memory = "256mb"

[metrics]
  port = 9091
  path = "/metrics"
```

### Relay Dockerfile

The relay ships with a multi-stage distroless Dockerfile. Verify it exists at `relay/Dockerfile`. If it does not, create it:

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o relay ./cmd/relay

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/relay /relay
EXPOSE 8080
ENTRYPOINT ["/relay"]
```

### Set Relay Secrets

```bash
# Generate and set JWT secret
JWT_SECRET=$(openssl rand -hex 32)
fly secrets set RELAY_JWT_SECRET="$JWT_SECRET" --app relix-relay

# Save this secret — you'll use the same value for the cloud service
echo "RELAY_JWT_SECRET=$JWT_SECRET"
```

### Deploy Relay

```bash
fly deploy --app relix-relay
fly status --app relix-relay
fly logs --app relix-relay

# Verify it's healthy
curl https://relix-relay.fly.dev/
# Should return 101 Switching Protocols (WebSocket upgrade expected)
```

### Scale Relay

The relay is stateful (in-memory connection registry). Scaling to multiple instances requires sticky sessions or a Redis-backed connection registry. For launch, run one instance:

```bash
# Keep exactly 1 machine running
fly scale count 1 --app relix-relay
```

When you need to scale beyond one instance (typically after 10K+ concurrent connections), you'll need to implement Redis-backed connection routing in `relay/internal/hub/hub.go`. The in-memory map in `Hub` is not process-safe across multiple instances.

---

## 6. Cloud Deployment

### fly.toml for Cloud

```toml
app = "relix-cloud"
primary_region = "iad"

[build]
  dockerfile = "Dockerfile"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = false
  auto_start_machines = true
  min_machines_running = 1

[[vm]]
  size = "shared-cpu-1x"
  memory = "256mb"
```

### Cloud Dockerfile

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o cloud ./cmd/cloud

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/cloud /cloud
EXPOSE 8080
ENTRYPOINT ["/cloud"]
```

### Set Cloud Secrets

Use the **same JWT secret** you set for the relay. Agents authenticate against the relay using tokens issued by cloud — they must share the secret.

```bash
# Same JWT secret as relay
fly secrets set JWT_SECRET="$JWT_SECRET" --app relix-cloud

# GitHub OAuth
fly secrets set GITHUB_CLIENT_ID="your_client_id" --app relix-cloud
fly secrets set GITHUB_CLIENT_SECRET="your_client_secret" --app relix-cloud

# Stripe (use test key until ready for real payments)
fly secrets set STRIPE_SECRET_KEY="sk_test_..." --app relix-cloud

# Database (after Fly Postgres attach, DATABASE_URL is set automatically)
# If using Supabase:
fly secrets set DATABASE_URL="postgres://..." --app relix-cloud
```

### Deploy Cloud

```bash
cd /path/to/cloud
fly launch --name relix-cloud --region iad --no-deploy
fly deploy --app relix-cloud
fly status --app relix-cloud

# Test health endpoint
curl https://relix-cloud.fly.dev/health
# Expected: {"status":"ok"}
```

---

## 7. Domain and DNS Setup

Register `relix.sh` at a domain registrar. Cloudflare is recommended (free DNS, good DDoS protection, easy certificate management).

### DNS Records

After adding the domain to Cloudflare:

| Type | Name | Value | Proxy | Purpose |
|------|------|-------|-------|---------|
| CNAME | `@` | `relix-landing.fly.dev` | Yes | Landing page (relix.sh) |
| CNAME | `www` | `relix-landing.fly.dev` | Yes | www redirect |
| CNAME | `relay` | `relix-relay.fly.dev` | No (DNS only) | WebSocket relay — Cloudflare proxy breaks WS |
| CNAME | `api` | `relix-cloud.fly.dev` | Yes | REST API |
| CNAME | `app` | `relix-app.fly.dev` | Yes | Web app (future) |

**Important:** `relay.relix.sh` must be set to "DNS only" (gray cloud) in Cloudflare. Cloudflare's proxy does not support WebSocket connections on the free plan. The relay uses `wss://` which is a raw WebSocket upgrade — it must bypass the proxy.

### Fly.io Custom Domain Setup

```bash
# Add custom domains to each Fly app
fly certs add relay.relix.sh --app relix-relay
fly certs add api.relix.sh --app relix-cloud

# Check certificate status
fly certs show relay.relix.sh --app relix-relay
fly certs show api.relix.sh --app relix-cloud
```

Fly.io will automatically provision Let's Encrypt certificates once the DNS records point to the Fly hostnames.

### Default Relay URL

The `relixctl` agent defaults to `wss://relay.relix.sh` (hardcoded in `relixctl/internal/config/config.go`). Once DNS is live, agents will connect automatically.

---

## 8. SSL/TLS

Fly.io handles TLS termination automatically for all HTTP services. The internal apps listen on plain HTTP on port 8080; Fly terminates TLS at the edge.

For the relay, because it is WebSocket:
- Clients connect to `wss://relay.relix.sh` (TLS)
- Fly terminates TLS, passes plain WebSocket frames to the relay on port 8080
- No changes needed in relay code

Certificate renewal is automatic (Let's Encrypt via Fly.io).

### Verifying TLS

```bash
curl -I https://api.relix.sh/health
# Should return HTTP/2 200

# Test WebSocket TLS
wscat -c wss://relay.relix.sh
# Should show WebSocket connection (then fail auth, which is expected)
```

---

## 9. Mobile App Deployment

### Prerequisites

- Expo account at https://expo.dev (free)
- EAS CLI: `npm install -g eas-cli`
- Apple Developer account ($99/year) — required for iOS
- Google Play Console account ($25 one-time) — required for Android

### EAS Setup

```bash
cd /path/to/mobile

# Log in
eas login

# Initialize EAS project (run once)
eas init

# This creates / updates eas.json
```

### eas.json Configuration

Create `mobile/eas.json`:

```json
{
  "cli": {
    "version": ">= 7.0.0"
  },
  "build": {
    "development": {
      "developmentClient": true,
      "distribution": "internal",
      "ios": {
        "simulator": true
      }
    },
    "preview": {
      "distribution": "internal",
      "ios": {
        "buildConfiguration": "Release"
      }
    },
    "production": {
      "autoIncrement": true,
      "ios": {
        "buildConfiguration": "Release"
      },
      "android": {
        "buildType": "app-bundle"
      }
    }
  },
  "submit": {
    "production": {
      "ios": {
        "appleId": "your@apple.id",
        "ascAppId": "YOUR_APP_STORE_CONNECT_APP_ID",
        "appleTeamId": "YOUR_TEAM_ID"
      },
      "android": {
        "serviceAccountKeyPath": "./google-play-service-account.json",
        "track": "internal"
      }
    }
  }
}
```

### app.json Configuration

Ensure `mobile/app.json` has correct bundle identifiers:

```json
{
  "expo": {
    "name": "Relix",
    "slug": "relix",
    "version": "1.0.0",
    "ios": {
      "bundleIdentifier": "sh.relix.app",
      "buildNumber": "1"
    },
    "android": {
      "package": "sh.relix.app",
      "versionCode": 1
    },
    "extra": {
      "eas": {
        "projectId": "YOUR_EAS_PROJECT_ID"
      }
    }
  }
}
```

### Building for iOS (TestFlight)

```bash
# Build for TestFlight
eas build --platform ios --profile production

# Submit to TestFlight (after build completes)
eas submit --platform ios --profile production

# Or submit a specific build
eas submit --platform ios --url "https://expo.dev/artifacts/..."
```

### Building for Android (Play Store)

```bash
# Build AAB for Play Store
eas build --platform android --profile production

# Submit to Play Store internal track
eas submit --platform android --profile production
```

### App Store Connect Setup (iOS)

1. https://appstoreconnect.apple.com → "My Apps" → "+" → "New App"
2. Platform: iOS, Name: Relix, Bundle ID: `sh.relix.app`
3. SKU: `relix-ios`
4. Complete all required metadata before first submission:
   - App description (see GO-TO-MARKET.md for copy)
   - Screenshots (6.7" iPhone required, 12.9" iPad optional)
   - App icon (already in `mobile/assets/icon.png` — must be 1024x1024 PNG)
   - Privacy policy URL: `https://relix.sh/privacy`
   - Support URL: `https://relix.sh/support`

### Google Play Console Setup (Android)

1. https://play.google.com/console → "Create app"
2. App name: Relix, Default language: English
3. App or game: App, Free or paid: Free
4. Complete store listing with same content as iOS
5. Create a service account for automated submissions:
   - IAM & Admin → Service accounts → Create
   - Grant "Release manager" role
   - Download JSON key → save as `mobile/google-play-service-account.json`
   - **Do not commit this file to git**

---

## 10. Push Notification Setup

### APNs (iOS) Setup

The current code in `cloud/internal/push/apns.go` is a stub. Replace it with a real APNs HTTP/2 implementation.

**Step 1: Create APNs Key**

1. https://developer.apple.com → Certificates, IDs & Profiles → Keys → "+"
2. Key name: `Relix Push Notifications`
3. Enable "Apple Push Notifications service (APNs)"
4. Download the `.p8` file — **this can only be downloaded once**
5. Note: Key ID (10 characters), Team ID (found in Membership)

**Step 2: Store APNs Credentials as Secrets**

```bash
fly secrets set APNS_KEY_ID="XXXXXXXXXX" --app relix-cloud
fly secrets set APNS_TEAM_ID="XXXXXXXXXX" --app relix-cloud
fly secrets set APNS_BUNDLE_ID="sh.relix.app" --app relix-cloud

# Store the .p8 key content
APNS_KEY=$(cat AuthKey_XXXXXXXXXX.p8)
fly secrets set APNS_PRIVATE_KEY="$APNS_KEY" --app relix-cloud
```

**Step 3: Replace the APNs Stub**

Install the `apns2` library:
```bash
cd cloud && go get github.com/sideshow/apns2
```

Replace `cloud/internal/push/apns.go` with a real implementation using `apns2.Client`.

### FCM (Android) Setup

**Step 1: Create Firebase Project**

1. https://console.firebase.google.com → "Add project" → Name: "Relix"
2. Add Android app → package name: `sh.relix.app`
3. Download `google-services.json` → place at `mobile/google-services.json`
   - **Do not commit to git** (contains API keys)

**Step 2: Get FCM Server Key**

1. Firebase Console → Project Settings → Cloud Messaging
2. Copy the "Server key" (starts with `AAAA...`)
3. Set as secret:

```bash
fly secrets set FCM_SERVER_KEY="AAAA..." --app relix-cloud
```

**Step 3: Replace the FCM Stub**

Install the FCM library:
```bash
cd cloud && go get github.com/appleboy/go-fcm
```

Replace the stub in `cloud/internal/push/` with a real FCM implementation.

### Unified Push Service

The current `push.Service` interface supports a single `Send(ctx, Notification)` method. For production, you'll want to:

1. Store device tokens per user with platform tag in the database (`device_tokens` table)
2. Look up all device tokens for a user when sending
3. Fan out to APNs for `platform=apns` and FCM for `platform=fcm`
4. Handle token expiry (remove stale tokens on 410 responses from APNs)

---

## 11. Stripe Setup

### Create Products and Prices

The billing stub in `cloud/internal/billing/stripe.go` needs to be replaced with real Stripe integration.

**Step 1: Create Products in Stripe Dashboard**

1. https://dashboard.stripe.com → Products → "Add product"
2. Create three products:

| Product | Monthly Price | Annual Price | Lookup Key |
|---------|--------------|--------------|------------|
| Relix Plus | $4.99 | $49.90 (2 months free) | `plus_monthly` / `plus_yearly` |
| Relix Pro | $14.99 | $149.90 | `pro_monthly` / `pro_yearly` |
| Relix Team | $24.99/user | $249.90/user | `team_monthly` / `team_yearly` |

3. Note the Price IDs (format: `price_xxxxx`) for each — you'll need them in code.

**Step 2: Configure Webhook**

1. Stripe Dashboard → Developers → Webhooks → "Add endpoint"
2. Endpoint URL: `https://api.relix.sh/billing/webhook`
3. Select events:
   - `checkout.session.completed`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`
   - `invoice.payment_failed`
4. Note the "Signing secret" (format: `whsec_xxxxx`)

```bash
fly secrets set STRIPE_WEBHOOK_SECRET="whsec_..." --app relix-cloud
```

**Step 3: Configure Billing Portal**

1. Stripe Dashboard → Settings → Billing → Customer Portal
2. Enable: Cancel subscriptions, Update payment methods, View invoices
3. Set return URL: `https://relix.sh/account`

**Step 4: Replace Stripe Stub**

```bash
cd cloud && go get github.com/stripe/stripe-go/v76
```

The stub is in `cloud/internal/billing/stripe.go`. Replace `StubStripe` with a real implementation:

```go
type RealStripe struct {
    secretKey string
    priceIDs  map[string]string // tier → Stripe price ID
}
```

Add webhook handler route `POST /billing/webhook` to `cloud/internal/api/server.go`.

**Step 5: Set Live vs Test Keys**

- Development: `sk_test_...` and `pk_test_...`
- Production: `sk_live_...` and `pk_live_...`

Never use live keys in development. Fly.io secrets can be updated per-app.

---

## 12. Monitoring

### Prometheus Metrics

The relay exposes Prometheus metrics at `:9091/metrics` (configured in `relay/fly.toml` under `[metrics]`). Fly.io scrapes this automatically if the metrics stanza is present.

Key relay metrics (from `relay/internal/metrics/metrics.go`):
- `relay_connections_active` — current active WebSocket connections
- `relay_messages_total` — total messages routed
- `relay_buffer_messages` — messages currently buffered for offline clients
- `relay_pairing_requests_total` — pairing code generations

### Grafana Dashboard on Fly.io

Fly.io provides a built-in metrics dashboard at https://fly.io/apps/relix-relay/metrics.

For a custom Grafana instance:

```bash
# Deploy Grafana as a Fly app
fly launch --name relix-grafana --image grafana/grafana:latest --region iad
fly secrets set GF_SECURITY_ADMIN_PASSWORD="$(openssl rand -base64 16)" --app relix-grafana
```

### Alerting

Set up Fly.io alerts at https://fly.io/apps/relix-relay/alerts:

| Alert | Threshold | Action |
|-------|-----------|--------|
| App restart | Any | PagerDuty / email |
| Memory > 90% | 90% | Scale up |
| CPU > 80% sustained | 5 min | Investigate |
| Health check failing | 2 consecutive | Page immediately |

### Uptime Monitoring

Use a free uptime monitor (Better Uptime, UptimeRobot, or Fly.io's built-in):

- Monitor `https://api.relix.sh/health` every 30 seconds
- Monitor `wss://relay.relix.sh` with WebSocket check
- Alert via email/SMS on downtime

### Log Aggregation

```bash
# Tail logs in real time
fly logs --app relix-relay
fly logs --app relix-cloud

# Fly.io ships logs to Logtail, Datadog, Papertrail automatically
# Configure at: https://fly.io/apps/relix-relay/monitoring
```

---

## 13. Backup and Disaster Recovery

### What Needs Backing Up

| Component | Data | Backup Strategy |
|-----------|------|----------------|
| PostgreSQL | Users, machines, subscriptions | Fly Postgres daily snapshots + Supabase built-in backups |
| Stripe | Billing data | Stripe is the source of truth — no separate backup needed |
| APNs key (`.p8`) | Push credential | Store in 1Password, never only on disk |
| JWT secret | Auth secret | Store in 1Password. Rotating requires re-login for all users. |
| GitHub OAuth secrets | Auth credential | Store in 1Password |

### Relay State

The relay is intentionally stateless — all connection state is in memory. On restart:
- Active connections are dropped; clients reconnect automatically (exponential backoff)
- Buffered messages are lost (24-hour buffer, acceptable loss on rare restart)
- No backup needed for relay state

### Database Backup (Fly Postgres)

```bash
# Manual snapshot
fly postgres backup create --app relix-postgres

# List backups
fly postgres backup list --app relix-postgres

# Restore
fly postgres backup restore <backup-id> --app relix-postgres
```

Fly Postgres takes daily snapshots automatically with 7-day retention on the free/paid tiers.

### Disaster Recovery Procedure

1. **Cloud service down:** `fly restart relix-cloud` — service is stateless, restarts in <30 seconds
2. **Relay down:** `fly restart relix-relay` — clients reconnect within 60 seconds automatically
3. **Database corrupted:** Restore from latest Fly Postgres snapshot, accept up to 24 hours data loss
4. **JWT secret compromised:** Generate new secret, set via `fly secrets set`, all users re-login on next request (JWTs expire after 24h)
5. **Fly.io region outage:** The `fly.toml` files specify `primary_region = "iad"`. If iad is down, Fly.io will not auto-failover for single-region apps. To add multi-region, add additional regions to `fly.toml` and ensure the relay uses Redis-backed state.

---

## 14. Cost Estimates

All prices as of March 2026.

### Zero Users (Development)

| Service | Cost/month |
|---------|-----------|
| Fly.io relay (shared-cpu-1x, 256MB) | ~$0 (free allowance) |
| Fly.io cloud (shared-cpu-1x, 256MB) | ~$0 (free allowance) |
| Fly.io Postgres (shared-cpu-1x, 1GB) | ~$0 (free allowance) |
| Fly.io Redis | ~$0 (free allowance) |
| Domain (relix.sh) | ~$2/month |
| **Total** | **~$2/month** |

### 1,000 Users

| Service | Cost/month |
|---------|-----------|
| Fly.io relay (dedicated-cpu-1x) | ~$20 |
| Fly.io cloud (shared-cpu-1x) | ~$5 |
| Fly.io Postgres (1GB SSD) | ~$10 |
| Fly.io Redis | ~$5 |
| EAS Build (production builds) | $0 (free tier: 30 builds/month) |
| Stripe fees (100 paying users × $4.99 avg) | ~$15 (2.9% + $0.30) |
| **Total infra** | **~$40/month** |
| **Revenue (assuming 10% paid conversion)** | **~$499/month** |

### 10,000 Users

| Service | Cost/month |
|---------|-----------|
| Fly.io relay (dedicated-cpu-2x, autoscale) | ~$100 |
| Fly.io cloud (dedicated-cpu-1x) | ~$40 |
| Fly.io Postgres (HA, 10GB) | ~$50 |
| Fly.io Redis | ~$20 |
| Monitoring (Grafana Cloud) | ~$0 (free tier) |
| Stripe fees (1,000 paying users) | ~$150 |
| **Total infra** | **~$210/month** |
| **Revenue (10% paid, avg $7 ARPU)** | **~$7,000/month** |

### Notes

- Fly.io free allowances: 3 shared VMs, 3GB total storage, 160GB outbound transfer
- WebSocket connections use persistent memory — each active connection costs ~10-20KB in the relay
- 10,000 concurrent WebSocket connections fit comfortably in 256MB RAM
- EAS Build free tier: 30 iOS + 30 Android builds/month. Paid plan ($29/month) for unlimited.
