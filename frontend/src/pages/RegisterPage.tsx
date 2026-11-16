import { register, type Credentials } from '../api/auth';
import { AuthForm } from '../components/AuthForm';

/** Same minimum as the backend (backend/internal/auth/password.go), checked early for instant feedback. */
const MIN_PASSWORD_LENGTH = 8;

/** Account creation page (US-19 scenario 1). */
export function RegisterPage() {
  return (
    <AuthForm
      title="Create account"
      submitLabel="Create account"
      passwordAutoComplete="new-password"
      validate={checkPasswordLength}
      submit={register}
    />
  );
}

function checkPasswordLength({ password }: Credentials) {
  return password.length < MIN_PASSWORD_LENGTH
    ? `Password must be at least ${MIN_PASSWORD_LENGTH} characters.`
    : null;
}
