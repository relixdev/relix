package relay

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/relixdev/protocol"
)

// SendMachineStatus sends a MsgMachineStatus envelope with the given status.
// Valid status values are "online", "offline", and "active".
func SendMachineStatus(ctx context.Context, client *RelayClient, machineID, status string) error {
	var s protocol.MachineStatus
	switch status {
	case "online":
		s = protocol.StatusOnline
	case "offline":
		s = protocol.StatusOffline
	case "active":
		s = protocol.StatusActive
	default:
		return fmt.Errorf("unknown machine status: %q", status)
	}

	payload, err := json.Marshal(protocol.MachineStatusMessage{
		MachineID: machineID,
		Status:    s,
	})
	if err != nil {
		return fmt.Errorf("marshal machine status: %w", err)
	}

	env := protocol.NewEnvelope(protocol.MsgMachineStatus, machineID)
	env.Payload = payload

	return client.Send(ctx, env)
}
