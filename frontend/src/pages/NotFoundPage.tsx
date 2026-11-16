import { Link } from 'react-router-dom';

export function NotFoundPage() {
  return (
    <section>
      <h1>Page not found</h1>
      <p className="lead">The page you are looking for does not exist.</p>
      <Link to="/">Back to home</Link>
    </section>
  );
}
