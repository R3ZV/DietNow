/**
 * Account registration and login (US-19).
 *
 * The backend answers both calls by setting an HttpOnly session cookie, so the
 * frontend never handles a token: the browser sends the cookie on later /api
 * requests automatically. See backend/AUTH.md.
 */
import { apiFetch } from './client';

export interface Credentials {
  email: string;
  password: string;
}

export interface AuthResponse {
  user: { id: number; email: string; displayName: string };
  /** When the session expires (RFC 3339). */
  expiresAt: string;
}

/** Creates an account and starts a session. Rejects with ApiError, e.g. when the email is taken. */
export const register = (credentials: Credentials) => postCredentials('/auth/register', credentials);

/** Starts a session for an existing account. Rejects with ApiError on wrong email or password. */
export const login = (credentials: Credentials) => postCredentials('/auth/login', credentials);

function postCredentials(path: string, credentials: Credentials) {
  return apiFetch<AuthResponse>(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(credentials),
  });
}

/** Restores the current session from the browser's HttpOnly cookie. */
export const getSession = () => apiFetch<AuthResponse>('/auth/me');

export const logout = () => apiFetch<void>('/auth/logout', { method: 'POST' });
