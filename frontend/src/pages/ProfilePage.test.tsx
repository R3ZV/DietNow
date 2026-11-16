import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { vi } from 'vitest';
import { ProfilePage } from './ProfilePage';

const profile = { gender: 'male', birthDate: '2000-01-01', heightCm: 180, currentWeightKg: 72.5, activityLevel: 'moderate', timezone: 'Europe/Bucharest', goal: 'maintain', targetWeightKg: 72.5, timeframeWeeks: 0 };
const response = (status: number, body: unknown) => new Response(JSON.stringify(body), { status });
function mockNew() {
  const fetch = vi.fn().mockResolvedValueOnce(response(404, { error: { code: 'profile_not_found' } }));
  vi.stubGlobal('fetch', fetch);
  render(<ProfilePage />);
  return fetch;
}
async function fill() {
  await screen.findByLabelText('Gender');
  for (const [label, value] of [['Gender', 'male'], ['Date of birth', '2000-01-01'], ['Height (cm)', '180'], ['Current weight (kg)', '72.5'], ['Activity level', 'moderate']]) {
    fireEvent.change(screen.getByLabelText(label), { target: { value } });
  }
}
it('saves validated intake with numeric units and maintenance targets', async () => {
  const fetch = mockNew();
  fetch.mockResolvedValueOnce(response(200, { profile }));
  await fill();
  fireEvent.click(screen.getByRole('button', { name: 'Save profile' }));
  expect(await screen.findByText(/Profile saved/)).toBeInTheDocument();
  expect(fetch.mock.calls[1][0]).toBe('/api/profile');
  expect(JSON.parse(fetch.mock.calls[1][1].body)).toMatchObject({ heightCm: 180, currentWeightKg: 72.5, goal: 'maintain', birthDate: '2000-01-01' });
});
it.each(['-10', '600'])('rejects invalid weight %s without saving', async weight => {
  const fetch = mockNew(); await fill();
  fireEvent.change(screen.getByLabelText('Current weight (kg)'), { target: { value: weight } });
  fireEvent.click(screen.getByRole('button', { name: 'Save profile' }));
  expect(screen.getByText(/weight between 30 and 300/)).toBeInTheDocument();
  expect(fetch).toHaveBeenCalledTimes(1);
});
it('rejects missing required fields', async () => {
  const fetch = mockNew();
  fireEvent.click(await screen.findByRole('button', { name: 'Save profile' }));
  expect(screen.getByLabelText('Gender')).toHaveAttribute('aria-invalid', 'true');
  expect(screen.getByLabelText('Date of birth')).toHaveAttribute('aria-invalid', 'true');
  expect(fetch).toHaveBeenCalledTimes(1);
});
it('loads a persisted profile and preserves its goal when saving', async () => {
  const existing = { ...profile, goal: 'bulk', targetWeightKg: 80, timeframeWeeks: 20 };
  const fetch = vi.fn().mockImplementation(() => Promise.resolve(response(200, { profile: existing })));
  vi.stubGlobal('fetch', fetch); render(<ProfilePage />);
  expect(await screen.findByLabelText('Height (cm)')).toHaveValue(180);
  fireEvent.click(screen.getByRole('button', { name: 'Save profile' }));
  await screen.findByText(/Profile saved/);
  expect(JSON.parse(fetch.mock.calls[1][1].body)).toMatchObject({ goal: 'bulk', targetWeightKg: 80, timeframeWeeks: 20 });
});
it('shows server field errors and retains input for correction', async () => {
  const fetch = mockNew();
  fetch.mockResolvedValueOnce(response(400, { error: { code: 'validation_failed', message: 'Invalid fields', fields: { heightCm: 'Height rejected by server' } } }));
  await fill(); fireEvent.click(screen.getByRole('button', { name: 'Save profile' }));
  expect(await screen.findByText('Height rejected by server')).toBeInTheDocument();
  expect(screen.getByLabelText('Height (cm)')).toHaveValue(180);
});
it('requires a session and allows retry after signing in', async () => {
  const fetch = vi.fn().mockResolvedValueOnce(response(401, { error: { code: 'unauthenticated' } })).mockResolvedValueOnce(response(200, { profile }));
  vi.stubGlobal('fetch', fetch); render(<ProfilePage />);
  expect(await screen.findByRole('alert')).toHaveTextContent('Please sign in');
  expect(screen.queryByRole('button', { name: 'Save profile' })).not.toBeInTheDocument();
  fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
  await waitFor(() => expect(screen.getByLabelText('Height (cm)')).toHaveValue(180));
});
