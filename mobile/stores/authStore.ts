import { create } from 'zustand';
import * as SecureStore from 'expo-secure-store';
import * as api from '../lib/api';
import type { User } from '../lib/api';

const TOKEN_KEY = 'relix_auth_token';
const USER_KEY = 'relix_user';

interface AuthState {
  token: string | null;
  user: User | null;
  isLoading: boolean;
  error: string | null;
  login: (provider: 'github' | 'email', credentials: Record<string, any>) => Promise<void>;
  logout: () => Promise<void>;
  loadToken: () => Promise<void>;
  refreshToken: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: null,
  user: null,
  isLoading: false,
  error: null,

  loadToken: async () => {
    set({ isLoading: true, error: null });
    try {
      const token = await SecureStore.getItemAsync(TOKEN_KEY);
      const userJson = await SecureStore.getItemAsync(USER_KEY);
      const user = userJson ? (JSON.parse(userJson) as User) : null;
      set({ token, user });
    } catch (e: any) {
      set({ error: e.message ?? 'Failed to load token' });
    } finally {
      set({ isLoading: false });
    }
  },

  login: async (provider, credentials) => {
    set({ isLoading: true, error: null });
    try {
      const result = await api.login(provider, credentials);
      const { token, user } = result;
      await SecureStore.setItemAsync(TOKEN_KEY, token);
      await SecureStore.setItemAsync(USER_KEY, JSON.stringify(user));
      set({ token, user, error: null });
    } catch (e: any) {
      set({ error: e.message ?? 'Login failed' });
      throw e;
    } finally {
      set({ isLoading: false });
    }
  },

  logout: async () => {
    set({ isLoading: true, error: null });
    try {
      await SecureStore.deleteItemAsync(TOKEN_KEY);
      await SecureStore.deleteItemAsync(USER_KEY);
      set({ token: null, user: null });
    } catch (e: any) {
      set({ error: e.message ?? 'Logout failed' });
    } finally {
      set({ isLoading: false });
    }
  },

  refreshToken: async () => {
    const { token } = get();
    if (!token) return;
    set({ isLoading: true, error: null });
    try {
      const result = await api.refreshToken(token);
      await SecureStore.setItemAsync(TOKEN_KEY, result.token);
      if (result.user) {
        await SecureStore.setItemAsync(USER_KEY, JSON.stringify(result.user));
        set({ token: result.token, user: result.user });
      } else {
        set({ token: result.token });
      }
    } catch (e: any) {
      set({ error: e.message ?? 'Token refresh failed' });
      throw e;
    } finally {
      set({ isLoading: false });
    }
  },
}));
