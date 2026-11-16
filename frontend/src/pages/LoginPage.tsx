import { login } from '../api/auth';
import { AuthForm } from '../components/AuthForm';

/** Login page (US-19 scenario 2). */
export function LoginPage() {
  return (
    <AuthForm
      title="Log in"
      submitLabel="Log in"
      passwordAutoComplete="current-password"
      submit={login}
    />
  );
}
