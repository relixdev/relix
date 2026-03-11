package relay_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relixctl/internal/relay"
)

func TestSendMachineStatus(t *testing.T) {
	cases := []struct {
		status   string
		protoVal protocol.MachineStatus
	}{
		{"online", protocol.StatusOnline},
		{"offline", protocol.StatusOffline},
		{"active", protocol.StatusActive},
	}

	for _, tc := range cases {
		t.Run(tc.status, func(t *testing.T) {
			recv := make(chan protocol.Envelope, 4)
			srv := newTestServer(t, recv, nil)
			defer srv.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			c := relay.NewClient()
			if err := c.Connect(ctx, wsURL(srv), "tok", "m-status"); err != nil {
				t.Fatalf("Connect: %v", err)
			}
			defer c.Close()

			// Drain auth.
			select {
			case <-recv:
			case <-ctx.Done():
				t.Fatal("timeout waiting for auth")
			}

			if err := relay.SendMachineStatus(ctx, c, "m-status", tc.status); err != nil {
				t.Fatalf("SendMachineStatus: %v", err)
			}

			select {
			case env := <-recv:
				if env.Type != protocol.MsgMachineStatus {
					t.Fatalf("expected MsgMachineStatus, got %q", env.Type)
				}
				if env.MachineID != "m-status" {
					t.Fatalf("expected machine_id m-status, got %q", env.MachineID)
				}
				var msg protocol.MachineStatusMessage
				if err := json.Unmarshal(env.Payload, &msg); err != nil {
					t.Fatalf("unmarshal payload: %v", err)
				}
				if msg.Status != tc.protoVal {
					t.Fatalf("expected status %q, got %q", tc.protoVal, msg.Status)
				}
				if msg.MachineID != "m-status" {
					t.Fatalf("expected MachineID m-status, got %q", msg.MachineID)
				}
			case <-ctx.Done():
				t.Fatal("timeout waiting for machine status envelope")
			}
		})
	}
}
