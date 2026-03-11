import { create } from 'zustand';
import * as api from '../lib/api';
import type { Session } from '../lib/protocol';

export interface Machine {
  id: string;
  name: string;
  status: 'online' | 'offline' | 'active';
  sessions: Session[];
  public_key: string;
  last_seen?: string;
}

export interface Approval {
  id: string;
  machine_id: string;
  session_id: string;
  tool: string;
  description: string;
  timestamp: number;
}

interface MachineState {
  machines: Machine[];
  pendingApprovals: Approval[];
  isLoading: boolean;
  error: string | null;
  fetchMachines: (token: string) => Promise<void>;
  updateMachineStatus: (machineId: string, status: Machine['status']) => void;
  addApproval: (approval: Approval) => void;
  resolveApproval: (approvalId: string, allowed: boolean) => void;
  setSessions: (machineId: string, sessions: Session[]) => void;
}

export const useMachineStore = create<MachineState>((set, get) => ({
  machines: [],
  pendingApprovals: [],
  isLoading: false,
  error: null,

  fetchMachines: async (token: string) => {
    set({ isLoading: true, error: null });
    try {
      const apiMachines = await api.listMachines(token);
      const machines: Machine[] = apiMachines.map((m) => ({
        id: m.id,
        name: m.name,
        public_key: m.public_key,
        status: (m.status as Machine['status']) ?? 'offline',
        last_seen: m.last_seen,
        sessions: get().machines.find((existing) => existing.id === m.id)?.sessions ?? [],
      }));
      set({ machines, error: null });
    } catch (e: any) {
      set({ error: e.message ?? 'Failed to fetch machines' });
      throw e;
    } finally {
      set({ isLoading: false });
    }
  },

  updateMachineStatus: (machineId: string, status: Machine['status']) => {
    set((state) => ({
      machines: state.machines.map((m) =>
        m.id === machineId ? { ...m, status } : m,
      ),
    }));
  },

  setSessions: (machineId: string, sessions: Session[]) => {
    set((state) => ({
      machines: state.machines.map((m) =>
        m.id === machineId ? { ...m, sessions } : m,
      ),
    }));
  },

  addApproval: (approval: Approval) => {
    set((state) => ({
      pendingApprovals: [...state.pendingApprovals, approval],
    }));
  },

  resolveApproval: (approvalId: string, _allowed: boolean) => {
    set((state) => ({
      pendingApprovals: state.pendingApprovals.filter((a) => a.id !== approvalId),
    }));
  },
}));
