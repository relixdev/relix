package hub

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/relixdev/protocol"
	"github.com/relixdev/relix/relay/internal/metrics"
)

// RouteEnvelope forwards env to the appropriate connection(s) based on role.
// If role is "agent", env is broadcast to all mobile connections for userID.
//   - If no mobiles are connected, the envelope is buffered for delivery when
//     a mobile connects.
//
// If role is "mobile", env is forwarded to the agent for (userID, env.MachineID).
// Returns an error if no target connections are found (agent-role returns nil
// when buffered successfully) or a write fails.
func (h *Hub) RouteEnvelope(userID, role string, env protocol.Envelope) error {
	start := time.Now()
	ctx := context.Background()

	switch role {
	case "agent":
		mobiles := h.GetMobiles(userID)
		if len(mobiles) == 0 {
			// No mobile connected — buffer the message.
			h.mu.Lock()
			buf := h.bufferForUser(userID)
			buf.Add(env)
			bufLen := float64(buf.Len())
			h.mu.Unlock()
			metrics.BufferSize.WithLabelValues(userID).Set(bufLen)
			return nil
		}
		var errs []error
		for _, m := range mobiles {
			if err := m.WriteEnvelope(ctx, env); err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return errors.Join(errs...)
		}
		metrics.MessagesRoutedTotal.WithLabelValues("agent_to_mobile").Inc()
		metrics.MessageLatency.Observe(time.Since(start).Seconds())
		return nil

	case "mobile":
		agent := h.GetAgent(userID, env.MachineID)
		if agent == nil {
			return fmt.Errorf("no agent for user %q machine %q", userID, env.MachineID)
		}
		if err := agent.WriteEnvelope(ctx, env); err != nil {
			return err
		}
		metrics.MessagesRoutedTotal.WithLabelValues("mobile_to_agent").Inc()
		metrics.MessageLatency.Observe(time.Since(start).Seconds())
		return nil

	default:
		return fmt.Errorf("unknown role %q", role)
	}
}
