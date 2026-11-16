import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getHealth, type HealthResponse } from '../api/client';

type ApiState =
  | { kind: 'loading' }
  | { kind: 'ok'; health: HealthResponse }
  | { kind: 'error'; message: string };

export function HomePage() {
  const [api, setApi] = useState<ApiState>({ kind: 'loading' });

  useEffect(() => {
    let cancelled = false;
    getHealth()
      .then((health) => !cancelled && setApi({ kind: 'ok', health }))
      .catch((err: unknown) => {
        if (!cancelled) {
          setApi({ kind: 'error', message: err instanceof Error ? err.message : String(err) });
        }
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <section>
      <h1>Welcome to DietNow</h1>
      <p className="lead">
        Track your meals and weight, hit your calorie and macro targets, and get science-backed
        recipe ideas from an AI nutritionist.
      </p>

      <div className="card-grid">
        <Link to="/tracker" className="card">
          <h2>Daily Tracker</h2>
          <p>Log breakfast, lunch, dinner and snacks.</p>
        </Link>
        <Link to="/chat" className="card">
          <h2>AI Nutritionist</h2>
          <p>Ask questions and generate recipes.</p>
        </Link>
        <Link to="/profile" className="card">
          <h2>Profile</h2>
          <p>Set your biometrics and goal.</p>
        </Link>
      </div>

      <p className="api-status" role="status">
        API:{' '}
        {api.kind === 'loading' && <span>checking…</span>}
        {api.kind === 'ok' && (
          <span className="status-ok">
            online ({api.health.service} {api.health.version})
          </span>
        )}
        {api.kind === 'error' && <span className="status-error">unreachable ({api.message})</span>}
      </p>
    </section>
  );
}
