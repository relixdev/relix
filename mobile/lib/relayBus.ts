import type { Payload, MessageType } from './protocol';

/**
 * Simple event bus for session screens to send messages via the relay client
 * held in the root layout. Avoids prop-drilling or context for the relay ref.
 */

type SendPayload = {
  machineId: string;
  sessionId: string;
  type: MessageType;
  payload: Payload;
};

type Listener = (msg: SendPayload) => void;

class MessageBus {
  private listeners: Set<Listener> = new Set();

  on(listener: Listener): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  emit(_event: 'send', msg: SendPayload): void {
    for (const listener of this.listeners) {
      try {
        listener(msg);
      } catch {
        // swallow listener errors
      }
    }
  }
}

export const RelayMessageBus = new MessageBus();
