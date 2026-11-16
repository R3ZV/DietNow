import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import type { Credentials } from '../api/auth';
import { errorMessage } from '../api/client';

interface AuthFormProps {
  title: string;
  submitLabel: string;
  /** Tells password managers whether to suggest a new password or fill the saved one. */
  passwordAutoComplete: 'new-password' | 'current-password';
  /** Returns a message to show instead of submitting, or null when the input is acceptable. */
  validate?: (credentials: Credentials) => string | null;
  /** The API call to make; it starts the session on success. */
  submit: (credentials: Credentials) => Promise<unknown>;
}

/**
 * Email + password form shared by the login and registration pages (US-19).
 * On success it opens the daily tracker; on failure it shows why.
 */
export function AuthForm({ title, submitLabel, passwordAutoComplete, validate, submit }: AuthFormProps) {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const invalid = validate?.({ email, password }) ?? null;
    if (invalid) {
      setError(invalid);
      return;
    }
    setError(null);
    setPending(true);
    try {
      await submit({ email, password });
      navigate('/tracker');
    } catch (err) {
      setError(errorMessage(err));
      setPending(false);
    }
  }

  return (
    <section>
      <h1>{title}</h1>
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
            autoComplete={passwordAutoComplete}
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
          {submitLabel}
        </button>
      </form>
    </section>
  );
}
