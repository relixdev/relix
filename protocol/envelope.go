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
