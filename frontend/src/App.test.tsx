import { screen, within, waitFor } from '@testing-library/react';
import { mockFetch, renderAt } from './test/helpers';

describe('app routing and layout', () => {
  it('renders the home page inside the shared layout and shows API health', async () => {
    const fetchMock = mockFetch(200, {
      status: 'ok',
      service: 'dietnow-api',
      version: 'dev',
      time: '2026-01-01T00:00:00Z',
    });

    renderAt('/');

    expect(screen.getByRole('navigation', { name: 'Main' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 1, name: /welcome to dietnow/i })).toBeInTheDocument();
    expect(await screen.findByText(/online \(dietnow-api dev\)/i)).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith('/api/health', expect.anything());
  });

  it('shows the API error message when the health check fails', async () => {
    mockFetch(503, { error: { code: 'unavailable', message: 'service down' } });

    renderAt('/');

    expect(await screen.findByText(/unreachable \(service down\)/i)).toBeInTheDocument();
  });

  it.each([
    ['/tracker', /daily tracker/i],
    ['/chat', /ai nutritionist/i],
    ['/profile', /profile/i],
    ['/register', /create account/i],
    ['/login', /log in/i],
  ])('routes %s to its page', async (path, heading) => {
    mockFetch(401, {});
    if (path === '/profile') mockFetch(404, { error: { code: 'profile_not_found' } });
    renderAt(path);
    expect(screen.getByRole('heading', { level: 1, name: heading })).toBeInTheDocument();
    await screen.findByRole('link', { name: 'Log in' });
    if (path === '/profile') await waitFor(() => expect(screen.getByRole('button', { name: 'Save profile' })).toBeInTheDocument());
  });

  it('links to the login and registration pages from the header (US-19)', async () => {
    mockFetch(401, {});
    renderAt('/tracker');
    const nav = within(screen.getByRole('navigation', { name: 'Main' }));

    expect(await nav.findByRole('link', { name: 'Log in' })).toHaveAttribute('href', '/login');
    expect(nav.getByRole('link', { name: 'Create account' })).toHaveAttribute('href', '/register');
  });

  it('renders a not-found page for unknown routes', async () => {
    mockFetch(401, {});
    renderAt('/nope');
    expect(screen.getByRole('heading', { level: 1, name: /page not found/i })).toBeInTheDocument();
    await screen.findByRole('link', { name: 'Log in' });
  });
});
