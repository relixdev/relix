# Plan 1: Protocol & Shared Types

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the shared Go module that defines wire protocol types, encryption primitives, and message serialization used by agent, relay, and mobile (via JSON).

**Architecture:** A standalone Go module (`github.com/relixdev/protocol`) containing zero-dependency types, NaCl encryption wrappers, and JSON serialization. All other Relix Go services import this module. The mobile app uses the JSON schema definitions directly.

**Tech Stack:** Go 1.23+, golang.org/x/crypto/nacl/box, encoding/json

**Spec:** `docs/specs/2026-03-11-relix-design.md`

---

## File Structure

```
protocol/
├── go.mod
├── go.sum
├── envelope.go         # Wire protocol envelope type + serialization (Unix timestamp)
├── envelope_test.go
├── messages.go         # All message payload types + Session/Event/UserInput types
├── messages_test.go
├── adapter.go          # CopilotAdapter interface (for agent plugin system)
├── adapter_test.go
├── crypto.go           # X25519 key generation, NaCl box encrypt/decrypt
├── crypto_test.go
├── version.go          # Protocol version constants
└── .gitignore
```

---

## Chunk 1: Module Setup & Envelope Types

### Task 1: Initialize Go Module

**Files:**
- Create: `protocol/go.mod`

- [ ] **Step 1: Create go.mod**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote
mkdir -p protocol
cd protocol
go mod init github.com/relixdev/protocol
```

- [ ] **Step 2: Add crypto dependency**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote/protocol
go get golang.org/x/crypto
```

- [ ] **Step 3: Create .gitignore**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote/protocol
echo "coverage.out" > .gitignore
```

- [ ] **Step 4: Commit**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote
git add protocol/go.mod protocol/go.sum protocol/.gitignore
git commit -m "chore: init protocol Go module with crypto dependency"
```

---

### Task 2: Protocol Version Constants

**Files:**
- Create: `protocol/version.go`

- [ ] **Step 1: Write version.go**

```go
// protocol/version.go
package protocol

const (
	// ProtocolVersion is the current wire protocol version.
	ProtocolVersion = 1

	// MinSupportedVersion is the oldest version this code can handle.
	MinSupportedVersion = 1
)
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote/protocol
go build ./...
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote
git add protocol/version.go
git commit -m "feat(protocol): add protocol version constants"
```

---

### Task 3: Envelope Type — Tests First

**Files:**
- Create: `protocol/envelope.go`
- Create: `protocol/envelope_test.go`

- [ ] **Step 1: Write failing tests for Envelope**

```go
// protocol/envelope_test.go
package protocol

import (
	"encoding/json"
	"testing"
)

func TestEnvelopeMarshalJSON(t *testing.T) {
	env := Envelope{
		V:         ProtocolVersion,
		Type:      MsgSessionEvent,
		MachineID: "m_abc123",
		SessionID: "s_def456",
		Timestamp: 1741689600,
		Payload:   []byte("test payload"),
	}

	data, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Verify wire format: timestamp must be a number, payload must be base64 string
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("raw unmarshal: %v", err)
	}
	ts, ok := raw["timestamp"].(float64)
	if !ok {
		t.Fatalf("timestamp is not a number: %T", raw["timestamp"])
	}
	if int64(ts) != 1741689600 {
		t.Errorf("timestamp: got %v, want 1741689600", ts)
	}

	var decoded Envelope
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.V != ProtocolVersion {
		t.Errorf("version: got %d, want %d", decoded.V, ProtocolVersion)
	}
	if decoded.Type != MsgSessionEvent {
		t.Errorf("type: got %q, want %q", decoded.Type, MsgSessionEvent)
	}
	if decoded.MachineID != "m_abc123" {
		t.Errorf("machine_id: got %q, want %q", decoded.MachineID, "m_abc123")
	}
	if decoded.SessionID != "s_def456" {
		t.Errorf("session_id: got %q, want %q", decoded.SessionID, "s_def456")
	}
	if string(decoded.Payload) != "test payload" {
		t.Errorf("payload: got %q, want %q", decoded.Payload, "test payload")
	}
}

func TestEnvelopeRequiresVersion(t *testing.T) {
	raw := `{"type":"session_event","machine_id":"m_1","session_id":"s_1","timestamp":0,"payload":""}`
	var env Envelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if env.V != 0 {
		t.Errorf("expected zero version for missing field, got %d", env.V)
	}
}

func TestMessageTypeConstants(t *testing.T) {
	types := []MessageType{
		MsgAuth,
		MsgSessionList,
		MsgSessionEvent,
		MsgUserInput,
		MsgApprovalResponse,
		MsgPing,
		MsgPong,
		MsgMachineStatus,
	}
	seen := make(map[MessageType]bool)
	for _, mt := range types {
		if mt == "" {
			t.Error("empty message type")
		}
		if seen[mt] {
			t.Errorf("duplicate message type: %q", mt)
		}
		seen[mt] = true
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote/protocol
go test ./... -v
```

Expected: FAIL — types not defined

- [ ] **Step 3: Write envelope.go to make tests pass**

```go
// protocol/envelope.go
package protocol

import "time"

// MessageType identifies the kind of wire message.
type MessageType string

const (
	MsgAuth             MessageType = "auth"
	MsgSessionList      MessageType = "session_list"
	MsgSessionEvent     MessageType = "session_event"
	MsgUserInput        MessageType = "user_input"
	MsgApprovalResponse MessageType = "approval_response"
	MsgPing             MessageType = "ping"
	MsgPong             MessageType = "pong"
	MsgMachineStatus    MessageType = "machine_status"
)

// Envelope is the outer wire message. The relay reads only the unencrypted
// fields for routing; Payload is an E2E-encrypted blob (base64-encoded in JSON).
// Timestamp is Unix seconds (integer) for cross-language compatibility.
type Envelope struct {
	V         int         `json:"v"`
	Type      MessageType `json:"type"`
	MachineID string      `json:"machine_id"`
	SessionID string      `json:"session_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
	Payload   []byte      `json:"payload,omitempty"`
}

// Now returns the Timestamp as a time.Time for convenience.
func (e Envelope) Now() time.Time {
	return time.Unix(e.Timestamp, 0)
}

// NewEnvelope creates an Envelope with the current time.
func NewEnvelope(msgType MessageType, machineID string) Envelope {
	return Envelope{
		V:         ProtocolVersion,
		Type:      msgType,
		MachineID: machineID,
		Timestamp: time.Now().Unix(),
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote/protocol
go test ./... -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote
git add protocol/envelope.go protocol/envelope_test.go
git commit -m "feat(protocol): add Envelope wire type and message type constants"
```

---

### Task 4: Message Payload Types — Tests First

**Files:**
- Create: `protocol/messages.go`
- Create: `protocol/messages_test.go`

- [ ] **Step 1: Write failing tests for payload types**

```go
// protocol/messages_test.go
package protocol

import (
	"encoding/json"
	"testing"
)

func TestPayloadRoundTrip(t *testing.T) {
	original := Payload{
		Kind: PayloadAssistantMessage,
		Seq:  42,
		Data: json.RawMessage(`{"text":"Hello from Claude"}`),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded Payload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Kind != PayloadAssistantMessage {
		t.Errorf("kind: got %q, want %q", decoded.Kind, PayloadAssistantMessage)
	}
	if decoded.Seq != 42 {
		t.Errorf("seq: got %d, want 42", decoded.Seq)
	}
}

func TestAuthMessageRoundTrip(t *testing.T) {
	msg := AuthMessage{Token: "jwt.token.here"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded AuthMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Token != "jwt.token.here" {
		t.Errorf("token: got %q", decoded.Token)
	}
}

func TestMachineStatusRoundTrip(t *testing.T) {
	msg := MachineStatusMessage{
		MachineID: "m_abc",
		Name:      "MacBook Pro",
		Status:    StatusOnline,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded MachineStatusMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Status != StatusOnline {
		t.Errorf("status: got %q, want %q", decoded.Status, StatusOnline)
	}
}

func TestPayloadKindConstants(t *testing.T) {
	kinds := []PayloadKind{
		PayloadAssistantMessage,
		PayloadToolUse,
		PayloadToolResult,
		PayloadPermissionRequest,
		PayloadUserMessage,
		PayloadApproval,
		PayloadError,
		PayloadSessionEnd,
	}
	seen := make(map[PayloadKind]bool)
	for _, k := range kinds {
		if k == "" {
			t.Error("empty payload kind")
		}
		if seen[k] {
			t.Errorf("duplicate kind: %q", k)
		}
		seen[k] = true
	}
}

func TestApprovalResponseDataRoundTrip(t *testing.T) {
	msg := ApprovalResponseData{ToolUseID: "tu_123", Approved: true}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded ApprovalResponseData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.ToolUseID != "tu_123" {
		t.Errorf("tool_use_id: got %q", decoded.ToolUseID)
	}
	if !decoded.Approved {
		t.Error("approved: got false, want true")
	}
}

func TestUserMessageDataRoundTrip(t *testing.T) {
	msg := UserMessageData{Text: "add auth middleware"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded UserMessageData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Text != "add auth middleware" {
		t.Errorf("text: got %q", decoded.Text)
	}
}

func TestSessionRoundTrip(t *testing.T) {
	s := Session{
		ID:        "s_abc123",
		Tool:      "claude-code",
		Project:   "/home/user/my-project",
		Status:    SessionActive,
		StartedAt: 1741689600,
	}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded Session
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.ID != "s_abc123" {
		t.Errorf("id: got %q", decoded.ID)
	}
	if decoded.Tool != "claude-code" {
		t.Errorf("tool: got %q", decoded.Tool)
	}
	if decoded.Status != SessionActive {
		t.Errorf("status: got %q", decoded.Status)
	}
}

func TestEventRoundTrip(t *testing.T) {
	e := Event{
		Kind: PayloadToolUse,
		Data: json.RawMessage(`{"tool":"Edit","file":"src/auth.ts"}`),
		Seq:  7,
	}
	data, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Kind != PayloadToolUse {
		t.Errorf("kind: got %q", decoded.Kind)
	}
	if decoded.Seq != 7 {
		t.Errorf("seq: got %d", decoded.Seq)
	}
}

func TestSessionStatusConstants(t *testing.T) {
	statuses := []SessionStatus{
		SessionActive,
		SessionWaitingApproval,
		SessionIdle,
		SessionEnded,
	}
	seen := make(map[SessionStatus]bool)
	for _, s := range statuses {
		if s == "" {
			t.Error("empty session status")
		}
		if seen[s] {
			t.Errorf("duplicate: %q", s)
		}
		seen[s] = true
	}
}

func TestToolUseDataRoundTrip(t *testing.T) {
	msg := ToolUseData{ToolUseID: "tu_1", Tool: "Edit", Input: "src/auth.ts"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded ToolUseData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Tool != "Edit" {
		t.Errorf("tool: got %q", decoded.Tool)
	}
}

func TestPermissionRequestDataRoundTrip(t *testing.T) {
	msg := PermissionRequestData{ToolUseID: "tu_1", Tool: "Bash", Description: "Run npm test"}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded PermissionRequestData
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded.Description != "Run npm test" {
		t.Errorf("description: got %q", decoded.Description)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote/protocol
go test ./... -v
```

Expected: FAIL — types not defined

- [ ] **Step 3: Write messages.go**

All types including Session, Event, UserInput live here so SessionListMessage compiles.

```go
// protocol/messages.go
package protocol

import "encoding/json"

// PayloadKind identifies the decrypted payload content type.
type PayloadKind string

const (
	PayloadAssistantMessage  PayloadKind = "assistant_message"
	PayloadToolUse           PayloadKind = "tool_use"
	PayloadToolResult        PayloadKind = "tool_result"
	PayloadPermissionRequest PayloadKind = "permission_request"
	PayloadUserMessage       PayloadKind = "user_message"
	PayloadApproval          PayloadKind = "approval"
	PayloadError             PayloadKind = "error"
	PayloadSessionEnd        PayloadKind = "session_end"
)

// Payload is the decrypted inner content of an Envelope.
type Payload struct {
	Kind PayloadKind     `json:"kind"`
	Seq  uint64          `json:"seq"`
	Data json.RawMessage `json:"data"`
}

// --- Session & Event types ---

// SessionStatus represents the state of a coding session.
type SessionStatus string

const (
	SessionActive          SessionStatus = "active"
	SessionWaitingApproval SessionStatus = "waiting_approval"
	SessionIdle            SessionStatus = "idle"
	SessionEnded           SessionStatus = "ended"
)

// Session represents a coding tool session on a machine.
type Session struct {
	ID        string        `json:"id"`
	Tool      string        `json:"tool"`
	Project   string        `json:"project"`
	Status    SessionStatus `json:"status"`
	StartedAt int64         `json:"started_at"`
}

// Event represents a single event from a coding tool session.
type Event struct {
	Kind PayloadKind     `json:"kind"`
	Data json.RawMessage `json:"data"`
	Seq  uint64          `json:"seq"`
}

// UserInputKind identifies the type of user input.
type UserInputKind string

const (
	InputMessage  UserInputKind = "message"
	InputApproval UserInputKind = "approval"
)

// UserInput represents input from the user (message or approval).
type UserInput struct {
	Kind UserInputKind   `json:"kind"`
	Data json.RawMessage `json:"data"`
}

// --- Wire message types ---

// MachineStatus represents the connection state of a machine.
type MachineStatus string

const (
	StatusOnline  MachineStatus = "online"
	StatusOffline MachineStatus = "offline"
	StatusActive  MachineStatus = "active"
)

// AuthMessage is sent on WebSocket connect to authenticate.
type AuthMessage struct {
	Token string `json:"token"`
}

// MachineStatusMessage is sent by the agent to update its status.
type MachineStatusMessage struct {
	MachineID string        `json:"machine_id"`
	Name      string        `json:"name"`
	Status    MachineStatus `json:"status"`
	OS        string        `json:"os,omitempty"`
	Hostname  string        `json:"hostname,omitempty"`
}

// SessionListMessage enumerates active sessions on a machine.
type SessionListMessage struct {
	MachineID string    `json:"machine_id"`
	Sessions  []Session `json:"sessions"`
}

// --- Payload data types ---

// ApprovalResponseData is the decrypted content of an approval response.
type ApprovalResponseData struct {
	ToolUseID string `json:"tool_use_id"`
	Approved  bool   `json:"approved"`
}

// UserMessageData is the decrypted content of a user message.
type UserMessageData struct {
	Text string `json:"text"`
}

// ToolUseData is the decrypted content of a tool use request.
type ToolUseData struct {
	ToolUseID string `json:"tool_use_id"`
	Tool      string `json:"tool"`
	Input     string `json:"input"`
}

// PermissionRequestData is the decrypted content of a permission request.
type PermissionRequestData struct {
	ToolUseID   string `json:"tool_use_id"`
	Tool        string `json:"tool"`
	Description string `json:"description"`
}

// ErrorData is the decrypted content of an error event.
type ErrorData struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// AssistantMessageData is the decrypted content of an assistant message.
type AssistantMessageData struct {
	Text string `json:"text"`
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote/protocol
go test ./... -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote
git add protocol/messages.go protocol/messages_test.go
git commit -m "feat(protocol): add payload types, auth, machine status, session list messages"
```

---

## Chunk 2: Adapter Interface & Crypto

### Task 5: CopilotAdapter Interface — Tests First

**Files:**
- Create: `protocol/adapter.go`
- Create: `protocol/adapter_test.go`

- [ ] **Step 1: Write failing test for adapter interface**

```go
// protocol/adapter_test.go
package protocol

import (
	"context"
	"encoding/json"
	"testing"
)

// mockAdapter verifies the interface is implementable
type mockAdapter struct {
	sessions []Session
}

func (m *mockAdapter) Discover(ctx context.Context) ([]Session, error) {
	return m.sessions, nil
}

func (m *mockAdapter) Attach(ctx context.Context, sessionID string) (<-chan Event, error) {
	ch := make(chan Event, 1)
	ch <- Event{Kind: PayloadAssistantMessage, Seq: 1, Data: json.RawMessage(`{"text":"hi"}`)}
	close(ch)
	return ch, nil
}

func (m *mockAdapter) Send(ctx context.Context, sessionID string, msg UserInput) error {
	return nil
}

func (m *mockAdapter) Detach(sessionID string) error {
	return nil
}

func TestMockImplementsCopilotAdapter(t *testing.T) {
	var _ CopilotAdapter = (*mockAdapter)(nil)

	m := &mockAdapter{sessions: []Session{
		{ID: "s_1", Tool: "claude-code", Project: "/tmp", Status: SessionActive},
	}}

	sessions, err := m.Discover(context.Background())
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].Tool != "claude-code" {
		t.Errorf("tool: got %q", sessions[0].Tool)
	}

	ch, err := m.Attach(context.Background(), "s_1")
	if err != nil {
		t.Fatalf("Attach: %v", err)
	}
	event := <-ch
	if event.Kind != PayloadAssistantMessage {
		t.Errorf("event kind: got %q", event.Kind)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote/protocol
go test ./... -v
```

Expected: FAIL — CopilotAdapter not defined

- [ ] **Step 3: Write adapter.go**

```go
// protocol/adapter.go
package protocol

import "context"

// CopilotAdapter is the interface that each coding tool integration must implement.
// The agent daemon uses this to bridge between a coding tool and the relay.
type CopilotAdapter interface {
	// Discover returns active sessions for this tool on the machine.
	Discover(ctx context.Context) ([]Session, error)

	// Attach connects to a specific session and returns a channel of events.
	// The channel is closed when the session ends or Detach is called.
	Attach(ctx context.Context, sessionID string) (<-chan Event, error)

	// Send delivers user input or an approval response to the session.
	Send(ctx context.Context, sessionID string, msg UserInput) error

	// Detach cleanly disconnects from a session.
	Detach(sessionID string) error
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote/protocol
go test ./... -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote
git add protocol/adapter.go protocol/adapter_test.go
git commit -m "feat(protocol): add CopilotAdapter interface"
```

---

### Task 6: Crypto — Key Generation & NaCl Box — Tests First

**Files:**
- Create: `protocol/crypto.go`
- Create: `protocol/crypto_test.go`

- [ ] **Step 1: Write failing tests**

```go
// protocol/crypto_test.go
package protocol

import (
	"bytes"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	if len(pub) != 32 {
		t.Errorf("public key length: got %d, want 32", len(pub))
	}
	if len(priv) != 32 {
		t.Errorf("private key length: got %d, want 32", len(priv))
	}

	// Keys should be different
	if bytes.Equal(pub[:], priv[:]) {
		t.Error("public and private keys are identical")
	}

	// Two calls should produce different keys
	pub2, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair (2): %v", err)
	}
	if bytes.Equal(pub[:], pub2[:]) {
		t.Error("two key pairs produced identical public keys")
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	// Alice and Bob generate key pairs
	alicePub, alicePriv, _ := GenerateKeyPair()
	bobPub, bobPriv, _ := GenerateKeyPair()

	plaintext := []byte("secret session data with code and approvals")

	// Alice encrypts for Bob
	ciphertext, err := Encrypt(plaintext, bobPub, alicePriv)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Ciphertext should be different from plaintext
	if bytes.Equal(ciphertext, plaintext) {
		t.Error("ciphertext equals plaintext")
	}

	// Ciphertext should be longer than plaintext (nonce + overhead)
	if len(ciphertext) <= len(plaintext) {
		t.Errorf("ciphertext not longer than plaintext: %d <= %d", len(ciphertext), len(plaintext))
	}

	// Bob decrypts
	decrypted, err := Decrypt(ciphertext, alicePub, bobPriv)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted: got %q, want %q", decrypted, plaintext)
	}
}

func TestDecryptWithWrongKeyFails(t *testing.T) {
	alicePub, alicePriv, _ := GenerateKeyPair()
	bobPub, _, _ := GenerateKeyPair()
	_, evePriv, _ := GenerateKeyPair()

	plaintext := []byte("secret data")
	ciphertext, _ := Encrypt(plaintext, bobPub, alicePriv)

	// Eve tries to decrypt with her private key (should fail)
	_, err := Decrypt(ciphertext, alicePub, evePriv)
	if err == nil {
		t.Error("expected decrypt to fail with wrong key, but it succeeded")
	}
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	alicePub, alicePriv, _ := GenerateKeyPair()
	bobPub, _, _ := GenerateKeyPair()
	_ = alicePub

	plaintext := []byte("same message")

	ct1, _ := Encrypt(plaintext, bobPub, alicePriv)
	ct2, _ := Encrypt(plaintext, bobPub, alicePriv)

	// Each encryption should use a random nonce, producing different ciphertext
	if bytes.Equal(ct1, ct2) {
		t.Error("two encryptions of the same plaintext produced identical ciphertext")
	}
}

func TestEncryptDecryptEmptyMessage(t *testing.T) {
	alicePub, alicePriv, _ := GenerateKeyPair()
	bobPub, bobPriv, _ := GenerateKeyPair()

	ciphertext, err := Encrypt([]byte{}, bobPub, alicePriv)
	if err != nil {
		t.Fatalf("Encrypt empty: %v", err)
	}

	decrypted, err := Decrypt(ciphertext, alicePub, bobPriv)
	if err != nil {
		t.Fatalf("Decrypt empty: %v", err)
	}

	if len(decrypted) != 0 {
		t.Errorf("expected empty, got %d bytes", len(decrypted))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote/protocol
go test ./... -v
```

Expected: FAIL — functions not defined

- [ ] **Step 3: Write crypto.go**

```go
// protocol/crypto.go
package protocol

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/box"
)

const nonceSize = 24

// PublicKey is a 32-byte X25519 public key.
type PublicKey [32]byte

// PrivateKey is a 32-byte X25519 private key.
type PrivateKey [32]byte

// GenerateKeyPair generates a new X25519 key pair for NaCl box.
func GenerateKeyPair() (PublicKey, PrivateKey, error) {
	pub, priv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return PublicKey{}, PrivateKey{}, err
	}
	return PublicKey(*pub), PrivateKey(*priv), nil
}

// Encrypt encrypts plaintext using NaCl box (X25519 + XSalsa20-Poly1305).
// Returns: nonce (24 bytes) || ciphertext.
func Encrypt(plaintext []byte, recipientPub PublicKey, senderPriv PrivateKey) ([]byte, error) {
	var nonce [nonceSize]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	pub := [32]byte(recipientPub)
	priv := [32]byte(senderPriv)

	sealed := box.Seal(nonce[:], plaintext, &nonce, &pub, &priv)
	return sealed, nil
}

// Decrypt decrypts a message produced by Encrypt.
// Expects: nonce (24 bytes) || ciphertext.
func Decrypt(message []byte, senderPub PublicKey, recipientPriv PrivateKey) ([]byte, error) {
	if len(message) < nonceSize+box.Overhead {
		return nil, errors.New("message too short")
	}

	var nonce [nonceSize]byte
	copy(nonce[:], message[:nonceSize])

	pub := [32]byte(senderPub)
	priv := [32]byte(recipientPriv)

	plaintext, ok := box.Open(nil, message[nonceSize:], &nonce, &pub, &priv)
	if !ok {
		return nil, errors.New("decryption failed: invalid key or corrupted message")
	}

	return plaintext, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote/protocol
go test ./... -v
```

Expected: PASS (all tests)

- [ ] **Step 5: Commit**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote
git add protocol/crypto.go protocol/crypto_test.go
git commit -m "feat(protocol): add X25519 NaCl box encryption primitives"
```

---

### Task 7: Convenience Helpers — Seal/Open Payload

**Files:**
- Modify: `protocol/crypto.go`
- Modify: `protocol/crypto_test.go`

- [ ] **Step 1: Write failing test for SealPayload/OpenPayload**

Add to `protocol/crypto_test.go`:

```go
func TestSealOpenPayload(t *testing.T) {
	alicePub, alicePriv, _ := GenerateKeyPair()
	bobPub, bobPriv, _ := GenerateKeyPair()

	payload := Payload{
		Kind: PayloadAssistantMessage,
		Seq:  1,
		Data: json.RawMessage(`{"text":"Hello"}`),
	}

	sealed, err := SealPayload(payload, bobPub, alicePriv)
	if err != nil {
		t.Fatalf("SealPayload: %v", err)
	}

	var opened Payload
	if err := OpenPayload(sealed, alicePub, bobPriv, &opened); err != nil {
		t.Fatalf("OpenPayload: %v", err)
	}

	if opened.Kind != PayloadAssistantMessage {
		t.Errorf("kind: got %q", opened.Kind)
	}
	if opened.Seq != 1 {
		t.Errorf("seq: got %d", opened.Seq)
	}
}
```

Add import `"encoding/json"` to the test file imports if not already present.

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote/protocol
go test -run TestSealOpenPayload -v
```

Expected: FAIL

- [ ] **Step 3: Add SealPayload and OpenPayload to crypto.go**

Append to `protocol/crypto.go`:

```go
// SealPayload JSON-marshals a Payload and encrypts it.
func SealPayload(p Payload, recipientPub PublicKey, senderPriv PrivateKey) ([]byte, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return Encrypt(data, recipientPub, senderPriv)
}

// OpenPayload decrypts and JSON-unmarshals a Payload.
func OpenPayload(ciphertext []byte, senderPub PublicKey, recipientPriv PrivateKey, out *Payload) error {
	plaintext, err := Decrypt(ciphertext, senderPub, recipientPriv)
	if err != nil {
		return err
	}
	return json.Unmarshal(plaintext, out)
}
```

Add `"encoding/json"` to the imports in crypto.go.

- [ ] **Step 4: Run all tests**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote/protocol
go test ./... -v
```

Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote
git add protocol/crypto.go protocol/crypto_test.go
git commit -m "feat(protocol): add SealPayload/OpenPayload convenience helpers"
```

---

### Task 8: Final — Run Full Test Suite & Tag

- [ ] **Step 1: Run full test suite with race detector**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote/protocol
go test -race -count=1 ./... -v
```

Expected: ALL PASS, no race conditions

- [ ] **Step 2: Check test coverage**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote/protocol
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

Expected: >80% coverage on all files

- [ ] **Step 3: Tag release**

```bash
cd /Users/zachforsyth/Side-Hustles/phone-apps/claude-remote
git tag protocol/v0.1.0
```
