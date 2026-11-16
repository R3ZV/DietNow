import { NavLink, Outlet } from 'react-router-dom';

const navItems = [
  { to: '/', label: 'Home', end: true },
  { to: '/tracker', label: 'Tracker', end: false },
  { to: '/chat', label: 'AI Chat', end: false },
  { to: '/profile', label: 'Profile', end: false },
];

/** Shared page chrome: header with navigation, main content area, footer. */
export function Layout() {
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
          </ul>
        </nav>
      </header>
      <main className="app-main">
        <Outlet />
      </main>
      <footer className="app-footer">DietNow · nutrition, calorie &amp; weight tracking</footer>
    </div>
  );
}
