export interface Machine {
  id: string;
  name: string;
  public_key: string;
  status: string;
  last_seen?: string;
  user_id?: string;
  created_at?: string;
}

export interface User {
  id: string;
  email: string;
  github_id?: string;
  tier: string;
  created_at?: string;
}

export interface PlanInfo {
  tier: string;
  display_name: string;
  monthly_price_cents: number;
  machine_limit: number;
  session_limit: number;
}

const API_BASE = 'https://relix-api.shadowscale.dev';
const RELAY_BASE = 'https://relix-relay.shadowscale.dev';

export function getApiBase(): string {
  return API_BASE;
}

export function getRelayBase(): string {
  return RELAY_BASE;
}

export function getRelayWsUrl(): string {
  return 'wss://relix-relay.shadowscale.dev/v1/ws';
}

export function getGitHubAuthUrl(): string {
  return `${API_BASE}/auth/github`;
}

// ─── HTTP helpers ──────────────────────────────────────────────────────

async function apiFetch<T>(
  base: string,
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const res = await fetch(`${base}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers ?? {}),
    },
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`API ${res.status}: ${body}`);
  }
  // 204 No Content
  if (res.status === 204) return undefined as unknown as T;
  return res.json() as Promise<T>;
}

function authHeaders(token: string): Record<string, string> {
  return { Authorization: `Bearer ${token}` };
}

// ─── Auth ──────────────────────────────────────────────────────────────

export async function loginWithGithub(
  code: string,
): Promise<{ token: string; user: User }> {
  return apiFetch(API_BASE, '/auth/github', {
    method: 'POST',
    body: JSON.stringify({ code }),
  });
}

export async function registerWithEmail(
  email: string,
  password: string,
): Promise<{ token: string; user: User }> {
  return apiFetch(API_BASE, '/auth/email/register', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });
}

export async function loginWithEmail(
  email: string,
  password: string,
): Promise<{ token: string; user: User }> {
  return apiFetch(API_BASE, '/auth/email/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });
}

/**
 * Unified login used by authStore. Routes to the correct endpoint based on provider and mode.
 */
export async function login(
  provider: 'github' | 'email',
  credentials: Record<string, any>,
): Promise<{ token: string; user: User }> {
  if (provider === 'github') {
    // GitHub OAuth: the token was already obtained from the redirect URL
    if (credentials.token) {
      const plan = await getPlan(credentials.token);
      return {
        token: credentials.token,
        user: { id: '', email: '', tier: plan.tier },
      };
    }
    return loginWithGithub(credentials.code);
  }
  // Email auth
  if (credentials.mode === 'signup') {
    return registerWithEmail(credentials.email, credentials.password);
  }
  return loginWithEmail(credentials.email, credentials.password);
}

export async function refreshToken(
  token: string,
): Promise<{ token: string; user: User }> {
  return apiFetch(API_BASE, '/auth/refresh', {
    method: 'POST',
    headers: authHeaders(token),
  });
}

// ─── Machines ──────────────────────────────────────────────────────────

export async function listMachines(token: string): Promise<Machine[]> {
  const res = await apiFetch<Machine[] | { machines: Machine[] }>(
    API_BASE,
    '/machines',
    { headers: authHeaders(token) },
  );
  if (Array.isArray(res)) return res;
  return res.machines ?? [];
}

export async function registerMachine(
  token: string,
  name: string,
  publicKey: string,
): Promise<Machine> {
  return apiFetch(API_BASE, '/machines', {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ name, public_key: publicKey }),
  });
}

export async function deleteMachine(
  token: string,
  machineId: string,
): Promise<void> {
  await apiFetch(API_BASE, `/machines/${machineId}`, {
    method: 'DELETE',
    headers: authHeaders(token),
  });
}

export async function renameMachine(
  token: string,
  machineId: string,
  name: string,
): Promise<Machine> {
  return apiFetch(API_BASE, `/machines/${machineId}`, {
    method: 'PATCH',
    headers: authHeaders(token),
    body: JSON.stringify({ name }),
  });
}

// ─── Billing ───────────────────────────────────────────────────────────

export async function getPlan(
  token: string,
): Promise<PlanInfo> {
  return apiFetch(API_BASE, '/billing/plan', {
    headers: authHeaders(token),
  });
}

export async function createCheckoutSession(
  token: string,
  tier: string,
): Promise<{ session_id: string; checkout_url: string }> {
  return apiFetch(API_BASE, '/billing/checkout', {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ tier }),
  });
}

export async function createBillingPortalSession(
  token: string,
  returnUrl: string,
): Promise<{ portal_url: string }> {
  return apiFetch(API_BASE, '/billing/portal', {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ return_url: returnUrl }),
  });
}

// ─── Push ──────────────────────────────────────────────────────────────

export async function registerPushToken(
  token: string,
  deviceToken: string,
  platform: 'ios' | 'android',
): Promise<void> {
  await apiFetch(API_BASE, '/push/register', {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ device_token: deviceToken, platform }),
  });
}

// ─── Pairing (via Relay) ──────────────────────────────────────────────

export async function generatePairingCode(
  token: string,
  userId?: string,
  mobilePublicKey?: string,
): Promise<{ code: string; expires_at: number }> {
  return apiFetch(RELAY_BASE, '/pair/code', {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({
      user_id: userId,
      mobile_public_key: mobilePublicKey,
    }),
  });
}

export async function checkPairingStatus(
  token: string,
  code: string,
): Promise<{
  status: 'pending' | 'completed' | 'expired' | 'failed';
  machine_id?: string;
  agent_public_key?: string;
  peer_public_key?: string;
  sas?: string[];
}> {
  return apiFetch(RELAY_BASE, `/pair/status/${code}`, {
    headers: authHeaders(token),
  });
}
