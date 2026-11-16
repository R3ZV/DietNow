import { useState } from 'react';
import { useAuth } from '../auth/context';
import { logout } from '../api/auth';
import { errorMessage } from '../api/client';
import { NavLink, Outlet, useNavigate } from 'react-router-dom';

const navItems = [
  { to: '/', label: 'Home', end: true },
  { to: '/tracker', label: 'Tracker', end: false },
  { to: '/chat', label: 'AI Chat', end: false },
  { to: '/profile', label: 'Profile', end: false },
];

/** Shared page chrome: header with navigation, main content area, footer. */
export function Layout() {
  const { session, loading, setSession } = useAuth();
  const navigate = useNavigate();
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleLogout() {
    setPending(true);
    setError(null);
    try {
      await logout();
      setSession(null);
      navigate('/login');
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="app-shell">
      <header className="app-header">
        <NavLink to="/" className="brand" end>
          DietNow
        </NavLink>
        <nav aria-label="Main">
          <ul className="nav-list">
            {navItems.map((item) => (
              <li key={item.to}>
                <NavLink to={item.to} end={item.end}>
                  {item.label}
                </NavLink>
              </li>
            ))}
            {!loading && (session ? (
              <li><button type="button" onClick={handleLogout} disabled={pending}>Log out</button></li>
            ) : (
              <>
                <li><NavLink to="/login">Log in</NavLink></li>
                <li><NavLink to="/register">Create account</NavLink></li>
              </>
            ))}
          </ul>
        </nav>
      </header>
      <main className="app-main">
        {error && <p role="alert" className="status-error">{error}</p>}
        <Outlet />
      </main>
      <footer className="app-footer">DietNow · nutrition, calorie &amp; weight tracking</footer>
    </div>
  );
}
