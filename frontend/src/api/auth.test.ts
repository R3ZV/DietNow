import { mockFetch } from '../test/helpers';
import { ApiError } from './client';
import { login, register } from './auth';

const authResponse = {
  user: { id: 1, email: 'ana@example.com', displayName: '' },
  expiresAt: '2026-11-23T18:00:00Z',
};

describe('auth API', () => {
  it.each([
    ['register', register, '/api/auth/register'],
    ['login', login, '/api/auth/login'],
  ])('%s sends the credentials as JSON to %s', async (_name, call, url) => {
    const fetchMock = mockFetch(200, authResponse);

    const result = await call({ email: 'ana@example.com', password: 'secret123' });

    expect(result).toEqual(authResponse);
    expect(fetchMock).toHaveBeenCalledWith(
      url,
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ email: 'ana@example.com', password: 'secret123' }),
        headers: expect.objectContaining({ 'Content-Type': 'application/json' }),
      }),
    );
  });

  it('turns a server error into an ApiError carrying the server message', async () => {
    mockFetch(401, { error: { code: 'invalid_credentials', message: 'invalid email or password' } });

    const attempt = login({ email: 'ana@example.com', password: 'wrong-pass' });

    await expect(attempt).rejects.toBeInstanceOf(ApiError);
    await expect(attempt).rejects.toThrow('invalid email or password');
  });
});
