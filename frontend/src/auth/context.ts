import { createContext, useContext } from 'react';
import type { AuthResponse } from '../api/auth';

export interface AuthState {
  session: AuthResponse | null;
  loading: boolean;
  setSession: (session: AuthResponse | null) => void;
}

export const AuthContext = createContext<AuthState | null>(null);

export function useAuth() {
  const auth = useContext(AuthContext);
  if (!auth) throw new Error('useAuth requires AuthProvider');
  return auth;
}
