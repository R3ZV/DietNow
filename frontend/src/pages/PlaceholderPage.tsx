interface PlaceholderPageProps {
  title: string;
  description: string;
}

/** Temporary page for routes whose feature has not been implemented yet. */
export function PlaceholderPage({ title, description }: PlaceholderPageProps) {
  return (
    <section>
      <h1>{title}</h1>
      <p className="lead">{description}</p>
      <p className="muted">Coming soon.</p>
    </section>
  );
}
