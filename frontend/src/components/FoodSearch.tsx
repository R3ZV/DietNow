import { useState, type FormEvent } from 'react';
import { errorMessage } from '../api/client';
import { searchFoods, type Food } from '../api/foods';

/**
 * Searches the food database by name and lists each match with its nutrition per 100 g (US-09).
 * Clicking a match hands that food to `onSelect` (US-10).
 */
export function FoodSearch({ onSelect }: { onSelect: (food: Food) => void }) {
  const [query, setQuery] = useState('');
  // null until the first search, so "no matches" is not shown before the user searched.
  const [foods, setFoods] = useState<Food[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [pending, setPending] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError(null);
    setPending(true);
    try {
      setFoods((await searchFoods(query)).foods);
    } catch (err) {
      setError(errorMessage(err));
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
      {foods?.length === 0 && <p className="muted">No foods found.</p>}
      <ul>
        {foods?.map((food) => {
          const n = food.nutrition;
          return (
            <li key={food.id}>
              <button type="button" onClick={() => onSelect(food)}>
                {`${food.name}: ${n.caloriesKcal} kcal, ${n.proteinG} g protein, ${n.carbsG} g carbs, ${n.fatG} g fat per 100 g`}
              </button>
            </li>
          );
        })}
      </ul>
    </div>
  );
}
