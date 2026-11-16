import { useCallback, useEffect, useRef, useState, type ReactNode } from 'react';
import { getSession, type AuthResponse } from '../api/auth';
import { AuthContext } from './context';

export function AuthProvider({ children }: { children: ReactNode }) {
  const [session, updateSession] = useState<AuthResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const revision = useRef(0);
  const setSession = useCallback((value: AuthResponse | null) => {
    revision.current++;
    updateSession(value);
    setLoading(false);
  }, []);

  useEffect(() => {
    let active = true;
    const startedAt = revision.current;
    getSession().then(
      (value) => {
        if (active && revision.current === startedAt) setSession(value);
      },
      () => {
        if (active && revision.current === startedAt) setSession(null);
      },
    );
    return () => { active = false; };
  }, [setSession]);

  return <AuthContext.Provider value={{ session, loading, setSession }}>{children}</AuthContext.Provider>;
}
