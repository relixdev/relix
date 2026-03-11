package relay

import (
	"context"
	"log"

	"github.com/relixdev/protocol"
)

// Bridge encrypts outbound events and decrypts inbound messages using NaCl box,
// sitting between the RelayClient and the session layer.
type Bridge struct {
	client    *RelayClient
	ownPub    protocol.PublicKey
	ownPriv   protocol.PrivateKey
	peerPub   protocol.PublicKey
	machineID string
}

// NewBridge creates a Bridge. ownPub/ownPriv are the agent's keys; peerPub is
// the mobile client's public key.
func NewBridge(client *RelayClient, ownPub protocol.PublicKey, ownPriv protocol.PrivateKey, peerPub protocol.PublicKey) *Bridge {
	return &Bridge{
		client:  client,
		ownPub:  ownPub,
		ownPriv: ownPriv,
		peerPub: peerPub,
	}
}

// SendEvent encrypts event as a Payload, wraps it in a MsgSessionEvent
// Envelope, and sends it via the client.
func (b *Bridge) SendEvent(ctx context.Context, event protocol.Event, sessionID string) error {
	p := protocol.Payload{
		Kind: event.Kind,
		Seq:  event.Seq,
		Data: event.Data,
	}

	ciphertext, err := protocol.SealPayload(p, b.peerPub, b.ownPriv)
	if err != nil {
		return err
	}

	env := protocol.NewEnvelope(protocol.MsgSessionEvent, b.client.MachineID())
	env.SessionID = sessionID
	env.Payload = ciphertext

	return b.client.Send(ctx, env)
}

// ReceiveLoop reads envelopes from the client's ReadLoop, filters for
// MsgUserInput and MsgApprovalResponse, decrypts their payloads, and delivers
// them to the returned channel. The channel is closed when the connection drops
// or ctx is cancelled.
func (b *Bridge) ReceiveLoop(ctx context.Context) <-chan protocol.Payload {
	ch := make(chan protocol.Payload, 64)
	rawCh := b.client.ReadLoop(ctx)

	go func() {
		defer close(ch)
		for {
			select {
			case env, ok := <-rawCh:
				if !ok {
					return
				}
				if env.Type != protocol.MsgUserInput && env.Type != protocol.MsgApprovalResponse {
					continue
				}
				var p protocol.Payload
				if err := protocol.OpenPayload(env.Payload, b.peerPub, b.ownPriv, &p); err != nil {
					log.Printf("bridge: decrypt failed (type=%s): %v", env.Type, err)
					continue
				}
				select {
				case ch <- p:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}
