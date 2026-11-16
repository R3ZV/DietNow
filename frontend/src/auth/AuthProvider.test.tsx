import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { vi } from 'vitest';
import { AuthProvider } from './AuthProvider';
import { Layout } from '../components/Layout';
import { LoginPage } from '../pages/LoginPage';
import { RegisterPage } from '../pages/RegisterPage';

const session = { user: { id: 1, email: 'test@dietnow.app', displayName: '' }, expiresAt: '2099-01-01T00:00:00Z' };
const response = (body: unknown, status = 200) => new Response(JSON.stringify(body), { status });

function app(path = '/tracker') {
  return render(
    <MemoryRouter initialEntries={[path]} future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
      <AuthProvider>
        <Routes>
          <Route element={<Layout />}>
            <Route path="/tracker" element={<h1>Tracker</h1>} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
          </Route>
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  );
}

it('restores the cookie session after remount and hides guest links', async () => {
  const fetchMock = vi.fn(async () => response(session));
  vi.stubGlobal('fetch', fetchMock);
  const first = app();
  expect(screen.queryByRole('link', { name: 'Create account' })).not.toBeInTheDocument();
  expect(await screen.findByRole('button', { name: 'Log out' })).toBeInTheDocument();
  first.unmount();
  app();
  expect(await screen.findByRole('button', { name: 'Log out' })).toBeInTheDocument();
  expect(screen.queryByRole('link', { name: 'Log in' })).not.toBeInTheDocument();
  expect(fetchMock).toHaveBeenCalledTimes(2);
  expect(fetchMock).toHaveBeenCalledWith('/api/auth/me', expect.anything());
});

it.each(['/login', '/register'])('updates navigation after submitting %s despite a late session lookup', async (path) => {
  let resolveSession!: (value: Response) => void;
  vi.stubGlobal('fetch', vi.fn((url: string) => url === '/api/auth/me'
    ? new Promise<Response>((resolve) => { resolveSession = resolve; })
    : Promise.resolve(response(session))));
  app(path);
  fireEvent.change(screen.getByLabelText('Email'), { target: { value: 'test@dietnow.app' } });
  fireEvent.change(screen.getByLabelText('Password'), { target: { value: 'Password123!' } });
  fireEvent.click(screen.getByRole('button', { name: path === '/login' ? 'Log in' : 'Create account' }));
  expect(await screen.findByRole('heading', { name: 'Tracker' })).toBeInTheDocument();
  await act(async () => { resolveSession(response({}, 401)); });
  expect(screen.getByRole('button', { name: 'Log out' })).toBeInTheDocument();
  expect(screen.queryByRole('link', { name: 'Create account' })).not.toBeInTheDocument();
});

it.each(['/login', '/register'])('redirects a restored session away from %s', async (path) => {
  vi.stubGlobal('fetch', vi.fn(async () => response(session)));
  app(path);
  expect(await screen.findByRole('heading', { name: 'Tracker' })).toBeInTheDocument();
});

it('retains login on a failed logout and clears it after a successful logout', async () => {
  let failLogout = true;
  const fetchMock = vi.fn(async (url: string) => {
    if (url === '/api/auth/me') return response(session);
    return failLogout ? response({ error: { message: 'Try again' } }, 503) : new Response(null, { status: 204 });
  });
  vi.stubGlobal('fetch', fetchMock);
  app();
  fireEvent.click(await screen.findByRole('button', { name: 'Log out' }));
  expect(await screen.findByRole('alert')).toHaveTextContent('Try again');
  expect(screen.getByRole('button', { name: 'Log out' })).toBeEnabled();
  failLogout = false;
  fireEvent.click(screen.getByRole('button', { name: 'Log out' }));
  await waitFor(() => expect(screen.getByRole('link', { name: 'Create account' })).toBeInTheDocument());
  expect(screen.queryByRole('button', { name: 'Log out' })).not.toBeInTheDocument();
  expect(fetchMock).toHaveBeenCalledWith('/api/auth/logout', expect.objectContaining({ method: 'POST' }));
});
