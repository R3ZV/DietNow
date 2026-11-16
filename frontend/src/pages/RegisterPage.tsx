import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { register } from '../api/auth';
import { ApiError } from '../api/client';

/** Same minimum as the backend (backend/internal/auth/password.go), checked early for instant feedback. */
const MIN_PASSWORD_LENGTH = 8;

/** Account creation form (US-19 scenario 1). On success the user lands on the daily tracker. */
export function RegisterPage() {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (password.length < MIN_PASSWORD_LENGTH) {
      setError(`Password must be at least ${MIN_PASSWORD_LENGTH} characters.`);
      return;
    }
    setError(null);
    setPending(true);
    try {
      await register({ email, password });
      navigate('/tracker');
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Could not reach the server. Try again.');
      setPending(false);
    }
  }

  return (
    <section>
      <h1>Create account</h1>
      <form className="form" onSubmit={handleSubmit}>
        <label>
          Email
          <input
            type="email"
            autoComplete="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
        </label>
        <label>
          Password
          <input
            type="password"
            autoComplete="new-password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </label>
        {error && (
          <p role="alert" className="status-error">
            {error}
          </p>
        )}
        <button type="submit" disabled={pending}>
          Create account
        </button>
      </form>
    </section>
  );
}
