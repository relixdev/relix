import { create } from 'zustand';
import type { Payload } from '../lib/protocol';

interface SessionState {
  currentSessionId: string | null;
  events: Payload[];
  /** All events keyed by session ID for multi-session support */
  sessionEvents: Record<string, Payload[]>;

  setSession: (sessionId: string) => void;
  addEvent: (event: Payload) => void;
  addSessionEvent: (sessionId: string, event: Payload) => void;
  getEventsForSession: (sessionId: string) => Payload[];
  clearSession: () => void;
}

export const useSessionStore = create<SessionState>((set, get) => ({
  currentSessionId: null,
  events: [],
  sessionEvents: {},

  setSession: (sessionId: string) => {
    const existing = get().sessionEvents[sessionId] ?? [];
    set({ currentSessionId: sessionId, events: existing });
  },

  addEvent: (event: Payload) => {
    set((state) => {
      const events = [...state.events, event];
      const sessionId = state.currentSessionId;
      if (!sessionId) return { events };
      const sessionEvents = { ...state.sessionEvents };
      sessionEvents[sessionId] = events;
      return { events, sessionEvents };
    });
  },

  addSessionEvent: (sessionId: string, event: Payload) => {
    set((state) => {
      const sessionEvents = { ...state.sessionEvents };
      const existing = sessionEvents[sessionId] ?? [];
      sessionEvents[sessionId] = [...existing, event];
      // If this is the current session, also update events
      if (state.currentSessionId === sessionId) {
        return { sessionEvents, events: sessionEvents[sessionId] };
      }
      return { sessionEvents };
    });
  },

  getEventsForSession: (sessionId: string) => {
    return get().sessionEvents[sessionId] ?? [];
  },

  clearSession: () => {
    set({ currentSessionId: null, events: [] });
  },
}));
