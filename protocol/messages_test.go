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
