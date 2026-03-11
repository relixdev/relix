package push_test

import (
	"context"
	"testing"

	"github.com/relixdev/relix/cloud/internal/push"
)

func testSend(t *testing.T, svc push.Service) {
	t.Helper()
	n := push.Notification{
		DeviceToken: "tok123",
		Title:       "Approval needed",
		Body:        "Edit src/main.go",
		Data:        map[string]string{"session_id": "s_abc"},
	}
	if err := svc.Send(context.Background(), n); err != nil {
		t.Errorf("Send returned error: %v", err)
	}
}

func TestAPNsStub(t *testing.T) {
	testSend(t, push.NewAPNs())
}

func TestFCMStub(t *testing.T) {
	testSend(t, push.NewFCM())
}
