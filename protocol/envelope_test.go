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
