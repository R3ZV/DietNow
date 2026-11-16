import { fireEvent, screen } from '@testing-library/react';
import { mockFetch, renderPage } from '../test/helpers';
import { LoginPage } from './LoginPage';

function submitLogin(email: string, password: string) {
  fireEvent.change(screen.getByLabelText('Email'), { target: { value: email } });
  fireEvent.change(screen.getByLabelText('Password'), { target: { value: password } });
  fireEvent.click(screen.getByRole('button', { name: 'Log in' }));
}

describe('login page (US-19 scenario 2)', () => {
  it('logs the user in and opens the daily tracker', async () => {
    const fetchMock = mockFetch(200, {
      user: { id: 1, email: 'ana@example.com', displayName: '' },
      expiresAt: '2026-11-23T18:00:00Z',
    });
    renderPage('/login', <LoginPage />);

    submitLogin('ana@example.com', 'secret123');

    expect(await screen.findByTestId('current-path')).toHaveTextContent('/tracker');
    expect(fetchMock).toHaveBeenCalledWith('/api/auth/login', expect.anything());
  });

  it('shows the server message when the credentials are wrong', async () => {
    mockFetch(401, { error: { code: 'invalid_credentials', message: 'invalid email or password' } });
    renderPage('/login', <LoginPage />);

    submitLogin('ana@example.com', 'wrong-pass');

    expect(await screen.findByRole('alert')).toHaveTextContent('invalid email or password');
  });
});
