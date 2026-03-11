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
