// Helpers shared by the frontend tests.
import { render } from '@testing-library/react';
import { RouterProvider, createMemoryRouter } from 'react-router-dom';
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

/** Replaces the global fetch so every request answers `status` with `body` as JSON. */
export function mockFetch(status: number, body: unknown) {
  const fetchMock = vi.fn().mockResolvedValue(
    new Response(JSON.stringify(body), {
      status,
      headers: { 'Content-Type': 'application/json' },
    }),
  );
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}
