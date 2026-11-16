import { useState, type FormEvent } from 'react';
import { errorMessage } from '../api/client';
import type { Food } from '../api/foods';

interface PortionFormProps {
  food: Food;
  /** Saves the portion; on failure the form shows why and stays open. */
  onSubmit: (grams: number) => Promise<unknown>;
}

/** Asks how many grams of the picked food were eaten (US-10). */
export function PortionForm({ food, onSubmit }: PortionFormProps) {
  const [grams, setGrams] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setPending(true);
    try {
      await onSubmit(Number(grams));
    } catch (err) {
      setError(errorMessage(err));
      setPending(false);
    }
  }

  return (
    <form className="form" onSubmit={handleSubmit}>
      <label>
        Grams of {food.name}
        {/* The backend accepts more than 0 and at most 2000 g. */}
        <input
          type="number"
          required
          min={1}
          max={2000}
          step="any"
          value={grams}
          onChange={(e) => setGrams(e.target.value)}
        />
      </label>
      {error && (
        <p role="alert" className="status-error">
          {error}
        </p>
      )}
      <button type="submit" disabled={pending}>
        Add
      </button>
    </form>
  );
}
