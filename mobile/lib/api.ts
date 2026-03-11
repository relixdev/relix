export interface Machine {
  id: string;
  name: string;
  public_key: string;
  status: string;
  last_seen?: string;
}

let API_BASE = 'https://api.relix.sh';

export function setApiBase(url: string): void {
  API_BASE = url;
}

async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
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
  return res.json() as Promise<T>;
}

function authHeaders(token: string): Record<string, string> {
  return { Authorization: `Bearer ${token}` };
}

export async function login(
  provider: 'github' | 'email',
  credentials: any,
): Promise<{ token: string }> {
  return apiFetch('/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({ provider, ...credentials }),
  });
}

export async function refreshToken(token: string): Promise<{ token: string }> {
  return apiFetch('/v1/auth/refresh', {
    method: 'POST',
    headers: authHeaders(token),
  });
}

export async function listMachines(token: string): Promise<Machine[]> {
  return apiFetch('/v1/machines', {
    headers: authHeaders(token),
  });
}

export async function registerMachine(
  token: string,
  name: string,
  publicKey: string,
): Promise<Machine> {
  return apiFetch('/v1/machines', {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ name, public_key: publicKey }),
  });
}

export async function deleteMachine(
  token: string,
  machineId: string,
): Promise<void> {
  await apiFetch(`/v1/machines/${machineId}`, {
    method: 'DELETE',
    headers: authHeaders(token),
  });
}

export async function registerPushToken(
  token: string,
  deviceToken: string,
  platform: 'ios' | 'android',
): Promise<void> {
  await apiFetch('/v1/push-tokens', {
    method: 'POST',
    headers: authHeaders(token),
    body: JSON.stringify({ device_token: deviceToken, platform }),
  });
}

export async function getPlan(
  token: string,
): Promise<{ tier: string; limits: any }> {
  return apiFetch('/v1/plan', {
    headers: authHeaders(token),
  });
}

export async function generatePairingCode(
  token: string,
): Promise<{ code: string; expires_at: number }> {
  return apiFetch('/v1/pairing/code', {
    method: 'POST',
    headers: authHeaders(token),
  });
}

export async function checkPairingStatus(
  token: string,
  code: string,
): Promise<{ status: 'pending' | 'completed'; machine_id?: string; peer_public_key?: string }> {
  return apiFetch(`/v1/pairing/code/${code}/status`, {
    headers: authHeaders(token),
  });
}

export async function renameMachine(
  token: string,
  machineId: string,
  name: string,
): Promise<Machine> {
  return apiFetch(`/v1/machines/${machineId}`, {
    method: 'PATCH',
    headers: authHeaders(token),
    body: JSON.stringify({ name }),
  });
}
