// Helpers shared by the frontend tests.
import { render } from '@testing-library/react';
import type { ReactElement } from 'react';
import {
  MemoryRouter,
  Route,
  RouterProvider,
  Routes,
  createMemoryRouter,
  useLocation,
} from 'react-router-dom';
import { vi } from 'vitest';
import { providerFutureFlags, routerFutureFlags, routes } from '../routes';

/** Renders the whole app at `path`, with the real routes and layout. */
export function renderAt(path: string) {
  const router = createMemoryRouter(routes, {
    initialEntries: [path],
    future: routerFutureFlags,
  });
  return render(<RouterProvider router={router} future={providerFutureFlags} />);
}

/**
 * Renders a single page at `path`. Navigating anywhere else shows the new path
 * in a `current-path` element, so a test can assert where the page sent the user.
 *
 * Uses a plain MemoryRouter on purpose: the data router behind renderAt builds
 * a fetch Request on every navigation, and in jsdom that Request rejects the
 * AbortSignal, so navigation cannot be tested through renderAt.
 */
export function renderPage(path: string, page: ReactElement) {
  return render(
    <MemoryRouter initialEntries={[path]} future={{ ...providerFutureFlags, v7_relativeSplatPath: true }}>
      <Routes>
        <Route path={path} element={page} />
        <Route path="*" element={<CurrentPath />} />
      </Routes>
    </MemoryRouter>,
  );
}

function CurrentPath() {
  return <p data-testid="current-path">{useLocation().pathname}</p>;
}

/** Replaces the global fetch so every request answers `status` with `body` as JSON. */
export function mockFetch(status: number, body: unknown) {
  // A fresh Response per call: a response body can be read only once.
  const fetchMock = vi.fn(
    async () =>
      new Response(JSON.stringify(body), {
        status,
        headers: { 'Content-Type': 'application/json' },
      }),
  );
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}
