import { fireEvent, screen } from '@testing-library/react';
import { mockFetch, renderPage } from '../test/helpers';
import { RegisterPage } from './RegisterPage';

function submitRegistration(email: string, password: string) {
  fireEvent.change(screen.getByLabelText('Email'), { target: { value: email } });
  fireEvent.change(screen.getByLabelText('Password'), { target: { value: password } });
  fireEvent.click(screen.getByRole('button', { name: 'Create account' }));
}

describe('register page (US-19 scenario 1)', () => {
  it('rejects a password shorter than 8 characters without calling the server', () => {
    const fetchMock = mockFetch(201, {});
    renderPage('/register', <RegisterPage />);

    submitRegistration('ana@example.com', 'short');

    expect(screen.getByRole('alert')).toHaveTextContent('Password must be at least 8 characters.');
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it('creates the account and opens the daily tracker', async () => {
    const fetchMock = mockFetch(201, {
      user: { id: 1, email: 'ana@example.com', displayName: '' },
      expiresAt: '2026-11-23T18:00:00Z',
    });
    renderPage('/register', <RegisterPage />);

    submitRegistration('ana@example.com', 'secret123');

    expect(await screen.findByTestId('current-path')).toHaveTextContent('/tracker');
    expect(fetchMock).toHaveBeenCalledWith('/api/auth/register', expect.anything());
  });

  it('shows the server message when the account cannot be created', async () => {
    mockFetch(409, {
      error: { code: 'email_taken', message: 'an account with this email already exists' },
    });
    renderPage('/register', <RegisterPage />);

    submitRegistration('ana@example.com', 'secret123');

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'an account with this email already exists',
    );
  });
});
