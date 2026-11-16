import { render, screen, waitFor } from '@testing-library/react';
import { RouterProvider, createMemoryRouter } from 'react-router-dom';
import { vi } from 'vitest';
import { providerFutureFlags, routerFutureFlags, routes } from './routes';

function renderAt(path: string) {
  const router = createMemoryRouter(routes, {
    initialEntries: [path],
    future: routerFutureFlags,
  });
  return render(<RouterProvider router={router} future={providerFutureFlags} />);
}

function mockFetch(status: number, body: unknown) {
  const fetchMock = vi.fn().mockResolvedValue(
    new Response(JSON.stringify(body), {
      status,
      headers: { 'Content-Type': 'application/json' },
    }),
  );
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

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
  ])('routes %s to its page', async (path, heading) => {
    if (path === '/profile') mockFetch(404, { error: { code: 'profile_not_found' } });
    renderAt(path);
    expect(screen.getByRole('heading', { level: 1, name: heading })).toBeInTheDocument();
    if (path === '/profile') await waitFor(() => expect(screen.getByRole('button', { name: 'Save profile' })).toBeInTheDocument());
  });

  it('renders a not-found page for unknown routes', () => {
    renderAt('/nope');
    expect(screen.getByRole('heading', { level: 1, name: /page not found/i })).toBeInTheDocument();
  });
});
