/**
 * Minimal typed client for the DietNow API.
 *
 * All endpoints live under /api (same origin; the Vite dev server proxies it to
 * the Go server). Errors follow the envelope documented in docs/ARCHITECTURE.md:
 *   { "error": { "code": "not_found", "message": "..." } }
 */

export const API_BASE = '/api';

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: { Accept: 'application/json', ...init?.headers },
  });

  const body: unknown = await res.json().catch(() => null);

  if (!res.ok) {
    const err = (body as { error?: { code?: string; message?: string } } | null)?.error;
    throw new ApiError(res.status, err?.code ?? 'http_error', err?.message ?? res.statusText);
  }
  return body as T;
}

export interface HealthResponse {
  status: 'ok';
  service: string;
  version: string;
  time: string;
}

export const getHealth = () => apiFetch<HealthResponse>('/health');
