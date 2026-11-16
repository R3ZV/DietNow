import { useState, type FormEvent } from 'react';
import { ApiError } from '../api/client';
import { searchFoods, type Food } from '../api/foods';

/** Searches the food database by name and lists each match with its nutrition per 100 g (US-09). */
export function FoodSearch() {
  const [query, setQuery] = useState('');
  const [foods, setFoods] = useState<Food[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setPending(true);
    try {
      setFoods((await searchFoods(query)).foods);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Could not reach the server. Try again.');
    }
    setPending(false);
  }

  return (
    <div>
      <form className="form" onSubmit={handleSubmit}>
        <label>
          Search food
          <input
            type="search"
            required
            minLength={2}
            maxLength={100}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </label>
        <button type="submit" disabled={pending}>
          Search
        </button>
      </form>
      {error && (
        <p role="alert" className="status-error">
          {error}
        </p>
      )}
      <ul>
        {foods.map(({ id, name, nutrition: n }) => (
          <li key={id}>
            {`${name}: ${n.caloriesKcal} kcal, ${n.proteinG} g protein, ${n.carbsG} g carbs, ${n.fatG} g fat per 100 g`}
          </li>
        ))}
      </ul>
    </div>
  );
}
