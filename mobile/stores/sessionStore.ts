import { create } from 'zustand';
import type { Payload } from '../lib/protocol';

interface SessionState {
  currentSessionId: string | null;
  events: Payload[];
  setSession: (sessionId: string) => void;
  addEvent: (event: Payload) => void;
  clearSession: () => void;
}

export const useSessionStore = create<SessionState>((set) => ({
  currentSessionId: null,
  events: [],

  setSession: (sessionId: string) => {
    set({ currentSessionId: sessionId, events: [] });
  },

  addEvent: (event: Payload) => {
    set((state) => ({ events: [...state.events, event] }));
  },

  clearSession: () => {
    set({ currentSessionId: null, events: [] });
  },
}));
