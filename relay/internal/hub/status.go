package hub

import (
	"context"
	"encoding/json"

	"github.com/relixdev/protocol"
)

// broadcastMachineStatus sends a MsgMachineStatus envelope to all mobile
// connections for userID. Errors are silently dropped (best-effort).
// Must NOT be called while h.mu is held.
func (h *Hub) broadcastMachineStatus(userID, machineID string, status protocol.MachineStatus) {
	payload, err := json.Marshal(protocol.MachineStatusMessage{
		MachineID: machineID,
		Status:    status,
	})
	if err != nil {
		return
	}

	env := protocol.NewEnvelope(protocol.MsgMachineStatus, machineID)
	env.Payload = payload

	mobiles := h.GetMobiles(userID)
	ctx := context.Background()
	for _, m := range mobiles {
		_ = m.WriteEnvelope(ctx, env)
	}
}
