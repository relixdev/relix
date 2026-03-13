# Relix Cloud — REST API Reference

**Base URL:** `https://api.relix.sh`

**Version:** v1 (all endpoints are unversioned at the path level; breaking changes will use a path prefix `/v2/`)

---

## Authentication

Most endpoints require a Bearer JWT token. Tokens are issued by the auth endpoints and expire after 24 hours. Refresh tokens before expiry using `POST /auth/refresh`.

```
Authorization: Bearer <token>
```

All tokens contain a `role` claim: `"mobile"` for app clients, `"agent"` for relixctl agents. Both roles use the same auth endpoints. The relay validates these same tokens when agents connect via WebSocket.

---

## Error Format

All errors use a consistent JSON envelope:

```json
{
  "error": "human-readable error message"
}
```

Common status codes:

| Code | Meaning |
|------|---------|
| `400` | Bad request — missing or invalid field |
| `401` | Unauthorized — missing or invalid token |
| `403` | Forbidden — valid token but insufficient permissions or tier limit |
| `404` | Not found — resource does not exist |
| `500` | Internal server error — something unexpected happened |

---

## Endpoints

### Authentication

---

#### `POST /auth/github`

Exchange a GitHub OAuth authorization code for a Relix JWT. Used by both the mobile app and the relixctl agent.

**Flow:**
1. App initiates GitHub OAuth, receives `code` in the callback URL
2. App calls this endpoint with that `code`
3. Relix exchanges the code with GitHub for a user profile
4. Relix creates the user account if it does not exist (first login)
5. Returns a JWT and the user object

**Request:**

```json
{
  "code": "github_oauth_authorization_code"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `code` | string | yes | GitHub OAuth authorization code from the callback |

**Response — `200 OK`:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "usr_a1b2c3d4",
    "email": "developer@example.com",
    "github_id": "12345678",
    "tier": "free",
    "created_at": "2026-03-13T00:00:00Z"
  }
}
```

**Error responses:**

| Status | Condition |
|--------|-----------|
| `400` | `code` is missing or empty |
| `400` | GitHub rejected the code (expired, already used, wrong client ID) |
| `500` | User creation failed (database error) |
| `500` | Token signing failed |

---

#### `POST /auth/email/register`

Create a new account with email and password. Returns a JWT immediately (no email verification at this time).

**Request:**

```json
{
  "email": "developer@example.com",
  "password": "secure-password-minimum-8-chars"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | yes | Valid email address |
| `password` | string | yes | Minimum 8 characters. Stored as bcrypt hash. |

**Response — `201 Created`:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "usr_a1b2c3d4",
    "email": "developer@example.com",
    "github_id": "",
    "tier": "free",
    "created_at": "2026-03-13T00:00:00Z"
  }
}
```

**Error responses:**

| Status | Condition |
|--------|-----------|
| `400` | Invalid request body (not valid JSON) |
| `400` | Email already registered |
| `400` | Password too short |
| `500` | Token signing failed |

---

#### `POST /auth/email/login`

Authenticate with email and password. Returns a fresh JWT.

**Request:**

```json
{
  "email": "developer@example.com",
  "password": "secure-password"
}
```

**Response — `200 OK`:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "usr_a1b2c3d4",
    "email": "developer@example.com",
    "github_id": "",
    "tier": "free",
    "created_at": "2026-03-11T00:00:00Z"
  }
}
```

**Error responses:**

| Status | Condition |
|--------|-----------|
| `400` | Invalid request body |
| `401` | Email not found or password does not match |
| `500` | Token signing failed |

**Note:** Both "email not found" and "wrong password" return `401` with message `"invalid credentials"`. This is intentional — it does not reveal whether an email address is registered.

---

#### `POST /auth/refresh`

Exchange a valid (not expired) JWT for a new JWT with a fresh 24-hour expiry. Requires authentication.

**Request:** No body required. The current token is read from the `Authorization` header.

**Response — `200 OK`:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "usr_a1b2c3d4",
    "email": "developer@example.com",
    "github_id": "",
    "tier": "free",
    "created_at": "2026-03-11T00:00:00Z"
  }
}
```

**Error responses:**

| Status | Condition |
|--------|-----------|
| `401` | Missing or invalid `Authorization` header |
| `401` | Token is expired — must re-authenticate |
| `401` | User no longer exists |
| `500` | Token signing failed |

**Recommended client behavior:** Call refresh when the token has less than 1 hour remaining. Store the new token immediately. If refresh fails with `401`, redirect to login.

---

### Machines

Machine objects represent developer machines (laptops, servers) that have relixctl installed.

**Machine object:**

```json
{
  "id": "mch_a1b2c3d4e5f6",
  "user_id": "usr_a1b2c3d4",
  "name": "my-macbook-pro",
  "public_key": "base64-encoded-x25519-public-key",
  "created_at": "2026-03-11T00:00:00Z"
}
```

---

#### `GET /machines`

List all machines registered to the authenticated user.

**Request:** No body. Auth required.

**Response — `200 OK`:**

```json
{
  "machines": [
    {
      "id": "mch_a1b2c3d4e5f6",
      "user_id": "usr_a1b2c3d4",
      "name": "my-macbook-pro",
      "public_key": "base64url-encoded-x25519-public-key",
      "created_at": "2026-03-11T00:00:00Z"
    },
    {
      "id": "mch_b2c3d4e5f6a1",
      "user_id": "usr_a1b2c3d4",
      "name": "dev-server-1",
      "public_key": "base64url-encoded-x25519-public-key",
      "created_at": "2026-03-12T00:00:00Z"
    }
  ]
}
```

Returns an empty array `[]` (never `null`) when no machines are registered.

**Error responses:**

| Status | Condition |
|--------|-----------|
| `401` | Missing or invalid token |
| `500` | Internal error |

---

#### `POST /machines`

Register a new machine. Called by `relixctl` during the pairing flow after the agent generates its X25519 key pair.

**Request:**

```json
{
  "name": "my-macbook-pro",
  "public_key": "base64url-encoded-x25519-public-key"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Human-readable name for the machine. Displayed in the mobile dashboard. |
| `public_key` | string | yes | Base64URL-encoded X25519 public key, generated by relixctl on first run. This is the key used for E2E encryption of all session traffic. |

**Response — `201 Created`:**

```json
{
  "id": "mch_a1b2c3d4e5f6",
  "user_id": "usr_a1b2c3d4",
  "name": "my-macbook-pro",
  "public_key": "base64url-encoded-x25519-public-key",
  "created_at": "2026-03-13T00:00:00Z"
}
```

**Error responses:**

| Status | Condition |
|--------|-----------|
| `400` | Missing or invalid request body |
| `400` | `name` or `public_key` is empty |
| `401` | Missing or invalid token |
| `403` | Tier machine limit reached. Message: `"machine: limit reached: Free tier allows 3 machines"` |
| `500` | Internal error |

**Note on limits:** Free tier allows 3 machines. Plus allows 10. Pro and Team are unlimited. Attempting to register beyond the limit returns `403` with a descriptive message. The mobile app should show an upgrade prompt when this occurs.

---

#### `DELETE /machines/:id`

Remove a machine. The machine must belong to the authenticated user. Called when the user removes a machine from the mobile app or runs `relixctl uninstall`.

**Request:** No body. Auth required. `:id` is the machine ID from the machine object.

**Response — `204 No Content`:** Empty body on success.

**Error responses:**

| Status | Condition |
|--------|-----------|
| `400` | Machine ID is empty |
| `401` | Missing or invalid token |
| `404` | Machine not found, or machine belongs to a different user |

---

#### `PATCH /machines/:id`

Rename a machine. Not yet implemented — this endpoint is specified here for completeness. Returns `501 Not Implemented` until built.

**Request:**

```json
{
  "name": "new-machine-name"
}
```

**Response — `200 OK`:**

```json
{
  "id": "mch_a1b2c3d4e5f6",
  "user_id": "usr_a1b2c3d4",
  "name": "new-machine-name",
  "public_key": "base64url-encoded-x25519-public-key",
  "created_at": "2026-03-11T00:00:00Z"
}
```

**Note:** To implement this, add `handleRenameMachine` in `cloud/internal/api/machine_handlers.go` and `registry.Rename()` in `cloud/internal/machine/registry.go`. See DEVELOPMENT.md for the step-by-step guide.

---

### Billing

---

#### `GET /billing/plan`

Get the current subscription tier and limits for the authenticated user.

**Request:** No body. Auth required.

**Response — `200 OK`:**

```json
{
  "tier": "free",
  "display_name": "Free",
  "monthly_price_cents": 0,
  "machine_limit": 3,
  "session_limit": 2
}
```

| Field | Type | Description |
|-------|------|-------------|
| `tier` | string | Internal tier name: `"free"`, `"plus"`, `"pro"`, `"team"` |
| `display_name` | string | Human-readable tier name: `"Free"`, `"Plus"`, `"Pro"`, `"Team"` |
| `monthly_price_cents` | integer | Monthly price in cents. `0` for free tier. `499` for Plus. `1499` for Pro. `2499` for Team. |
| `machine_limit` | integer | Maximum machines allowed. `-1` means unlimited. |
| `session_limit` | integer | Maximum concurrent sessions allowed. `-1` means unlimited. |

**Tier reference:**

| Tier | Price | `machine_limit` | `session_limit` |
|------|-------|-----------------|-----------------|
| `free` | $0 | 3 | 2 |
| `plus` | $4.99/mo (499 cents) | 10 | 5 |
| `pro` | $14.99/mo (1499 cents) | -1 (unlimited) | -1 (unlimited) |
| `team` | $24.99/user/mo (2499 cents) | -1 (unlimited) | -1 (unlimited) |

**Error responses:**

| Status | Condition |
|--------|-----------|
| `401` | Missing or invalid token |
| `404` | User not found |

---

#### `POST /billing/checkout`

Create a Stripe Checkout session to upgrade the user's subscription tier. Returns a `checkout_url` that the mobile app opens in a browser or WebView.

**Note:** Currently uses a stub (`StubStripe`). Returns a fake URL. Replace `cloud/internal/billing/stripe.go` with a real Stripe implementation before going live.

**Request:**

```json
{
  "tier": "plus"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `tier` | string | yes | Target tier: `"plus"`, `"pro"`, or `"team"` |

**Response — `200 OK`:**

```json
{
  "session_id": "cs_live_a1b2c3d4...",
  "checkout_url": "https://checkout.stripe.com/pay/cs_live_a1b2c3d4..."
}
```

| Field | Description |
|-------|-------------|
| `session_id` | Stripe Checkout session ID. Can be used with Stripe.js to redirect. |
| `checkout_url` | Full redirect URL for the Stripe checkout page. Open this in a browser. |

**Mobile client behavior:**
1. Call `POST /billing/checkout` with the desired tier
2. Open `checkout_url` in the in-app browser (`expo-web-browser`)
3. On return, call `GET /billing/plan` to check the updated tier
4. Stripe sends a webhook to `POST /billing/webhook` when payment completes

**Error responses:**

| Status | Condition |
|--------|-----------|
| `400` | `tier` is missing or empty |
| `401` | Missing or invalid token |
| `500` | Stripe API error |

---

#### `POST /billing/portal`

Create a Stripe Billing Portal session for managing an existing subscription. Not yet implemented — returns `501 Not Implemented`. Required before users can cancel or change payment methods.

**Request:**

```json
{
  "return_url": "https://relix.sh/account"
}
```

**Response — `200 OK`:**

```json
{
  "portal_url": "https://billing.stripe.com/session/..."
}
```

**Implementation note:** Add to `cloud/internal/api/billing_handlers.go` and register route `POST /billing/portal` in `server.go`. Requires a real Stripe implementation.

---

#### `POST /billing/webhook`

Internal endpoint called by Stripe when subscription events occur. This endpoint is **not** called by the mobile app or agents — it is called by Stripe's webhook system.

Not yet implemented. Required before real billing works.

**Request:** Raw Stripe webhook body with `Stripe-Signature` header for verification.

Key events to handle:

| Event | Action |
|-------|--------|
| `checkout.session.completed` | Upgrade user's tier in the database |
| `customer.subscription.updated` | Update user's tier (plan change) |
| `customer.subscription.deleted` | Downgrade user to free tier |
| `invoice.payment_failed` | Send email notification, optionally restrict features |

**Implementation note:** Register route `POST /billing/webhook` in `server.go`. Use `stripe-go` library's `webhook.ConstructEvent` with the `STRIPE_WEBHOOK_SECRET` to verify the payload before processing. After verifying, update the user's `tier` field in the database.

---

### Push Notifications

---

#### `POST /push/register`

Register a device push token for the authenticated user. Called by the mobile app after receiving a push token from the OS.

**Request:**

```json
{
  "device_token": "apns-or-fcm-device-token",
  "platform": "apns"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `device_token` | string | yes | The push token received from the OS. For iOS: APNs token (hex string). For Android: FCM registration token. |
| `platform` | string | yes | `"apns"` for iOS, `"fcm"` for Android. |

**Response — `200 OK`:**

```json
{
  "status": "registered"
}
```

**Note on current implementation:** The handler currently returns `"registered"` without persisting the token to a database (there is a `// TODO` comment in `push_handlers.go`). Before push notifications work in production, implement database persistence of device tokens in the `device_tokens` table. See DEPLOYMENT.md for the schema.

**Error responses:**

| Status | Condition |
|--------|-----------|
| `400` | Invalid request body |
| `400` | `device_token` is empty |
| `401` | Missing or invalid token |

---

#### `POST /push/send`

Send a push notification to a specific device. This is an **internal endpoint** — it is called by the relay or other cloud components when an agent emits an approval-needed event. It is **not** part of the public mobile app API.

Authentication is still required (agent JWT token with `role=agent`).

**Request:**

```json
{
  "device_token": "apns-or-fcm-device-token",
  "title": "Approval needed",
  "body": "Edit src/auth.ts — Claude Code is waiting",
  "data": {
    "machine_id": "mch_a1b2c3d4",
    "session_id": "s_e5f6a1b2",
    "type": "approval_request"
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `device_token` | string | yes | The device's push token |
| `title` | string | yes | Notification title (shown in the notification banner) |
| `body` | string | no | Notification body text |
| `data` | object | no | Key-value pairs sent as the notification payload. Used by the app to deep-link to the relevant session. |

**Response — `200 OK`:**

```json
{
  "status": "sent"
}
```

**Note:** Currently uses a stub that logs the notification and returns success. Replace with real APNs/FCM implementation. See DEPLOYMENT.md section 10.

**Error responses:**

| Status | Condition |
|--------|-----------|
| `400` | Invalid request body |
| `400` | `device_token` or `title` is empty |
| `401` | Missing or invalid token |
| `500` | Push delivery failed (APNs or FCM error) |

---

### Pairing

The pairing flow uses the relay, not the cloud API. The relay exposes HTTP endpoints for pairing code generation and status polling. These are served at `https://relay.relix.sh/pair/...`, not at `api.relix.sh`.

---

#### `POST /pair/code`

Generate a new pairing code. Called by the mobile app when the user taps "Add Machine."

**Relay endpoint:** `POST https://relay.relix.sh/pair/code`

**Request:**

```json
{
  "user_id": "usr_a1b2c3d4",
  "mobile_public_key": "base64url-encoded-x25519-public-key"
}
```

| Field | Description |
|-------|-------------|
| `user_id` | The authenticated user's ID (from the JWT) |
| `mobile_public_key` | The mobile app's X25519 public key, generated fresh for this pairing session |

**Response — `200 OK`:**

```json
{
  "code": "847293",
  "expires_at": 1741691400
}
```

| Field | Description |
|-------|-------------|
| `code` | 6-digit numeric pairing code shown in the mobile app |
| `expires_at` | Unix timestamp. Code expires after 5 minutes. |

**Rate limiting:** Maximum 3 active codes per user. Per-IP rate limit: 1 request per 3 seconds.

---

#### `GET /pair/status/:code`

Poll for pairing completion. The mobile app polls this endpoint after displaying the code, waiting for the agent to call `relixctl pair <code>`.

**Relay endpoint:** `GET https://relay.relix.sh/pair/status/:code`

**Response — `200 OK` (pending):**

```json
{
  "status": "pending",
  "sas": null
}
```

**Response — `200 OK` (completed):**

```json
{
  "status": "completed",
  "machine_id": "mch_a1b2c3d4e5f6",
  "agent_public_key": "base64url-encoded-x25519-public-key",
  "sas": ["🐶", "🌊", "🔑", "🎸"]
}
```

| Field | Description |
|-------|-------------|
| `status` | `"pending"` — agent has not entered the code yet. `"completed"` — pairing succeeded. `"expired"` — code expired before use. `"failed"` — too many wrong attempts. |
| `machine_id` | The machine ID assigned to the newly paired machine |
| `agent_public_key` | The agent's X25519 public key — used by the mobile app to set up E2E encryption |
| `sas` | Short Authentication String — 4 emojis derived from the ECDH shared secret. Both the mobile app and the agent display this. Users can visually verify they are pairing with the correct device. |

**Poll interval:** 2 seconds is appropriate. Stop polling after `status` is `"completed"`, `"expired"`, or `"failed"`, or after the `expires_at` time has passed.

---

## Relay WebSocket Protocol

The relay is a WebSocket server at `wss://relay.relix.sh`. Clients authenticate immediately on connect by sending an `auth` message.

### Connection and Authentication

```javascript
const ws = new WebSocket('wss://relay.relix.sh');
ws.onopen = () => {
  ws.send(JSON.stringify({
    v: 1,
    type: 'auth',
    payload: '<base64-encoded-jwt-token>'
  }));
};
```

The relay validates the JWT. If invalid, the connection is closed immediately with code `4001`.

### Envelope Format

All messages are JSON:

```json
{
  "v": 1,
  "type": "session_event",
  "machine_id": "mch_a1b2c3d4",
  "session_id": "s_e5f6a1b2",
  "timestamp": 1741689600,
  "payload": "<base64-encoded-nacl-box-ciphertext>"
}
```

The `payload` field is always E2E encrypted. The relay routes based on `machine_id` without reading `payload`.

### Message Types

| Type | Direction | Description |
|------|-----------|-------------|
| `auth` | client → relay | JWT authentication. Must be first message after connect. |
| `session_list` | agent → relay → mobile | Agent enumerates active sessions. `payload` contains `[]Session`. |
| `session_event` | agent → relay → mobile | Claude Code event (assistant message, tool use, permission request, etc.). `payload` contains `Event`. |
| `user_input` | mobile → relay → agent | User message from phone. `payload` contains `UserInput` with `kind="message"`. |
| `approval_response` | mobile → relay → agent | Approval decision. `payload` contains `UserInput` with `kind="approval_response"`. |
| `machine_status` | agent → relay | Online/offline/active status update. Not encrypted (metadata only). |
| `ping` | bidirectional | Keepalive. Send every 30 seconds. Relay closes connection after 60 seconds of silence. |
| `pong` | bidirectional | Response to ping. |

### Encrypted Payload (After Decryption)

```json
{
  "kind": "assistant_message",
  "data": { "content": "I'll edit the file now." },
  "seq": 42
}
```

`seq` is monotonically increasing per session, used for replay detection and ordering buffered messages.

---

## Health Check

#### `GET /health`

Public endpoint, no authentication required.

**Response — `200 OK`:**

```json
{
  "status": "ok"
}
```

Use this endpoint for uptime monitoring, load balancer health checks, and deployment verification.
