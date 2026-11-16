import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ApiError } from '../api/client';
import { getDailySummary, type DailySummary, type MealType } from '../api/tracker';

const mealLabels: Record<MealType, string> = {
  breakfast: 'Breakfast',
  lunch: 'Lunch',
  dinner: 'Dinner',
  snack: 'Snacks',
};

/** The daily tracker (US-08): one day's four meals, each with its foods and calorie subtotal. */
export function TrackerPage() {
  const navigate = useNavigate();
  // Unset until the user picks a day; the backend then shows today in the user's time zone.
  const [date, setDate] = useState<string>();
  const [summary, setSummary] = useState<DailySummary | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // `cancelled` drops a slow answer for a day the user has already left.
    let cancelled = false;
    getDailySummary(date)
      .then((day) => !cancelled && setSummary(day))
      .catch((err: unknown) => {
        if (cancelled) return;
        // replace: Back from the login page must not land here and redirect again.
        if (err instanceof ApiError && err.status === 401) navigate('/login', { replace: true });
        else setError(err instanceof Error ? err.message : String(err));
      });
    return () => {
      cancelled = true;
    };
  }, [date, navigate]);

  return (
    <section>
      <h1>Daily Tracker</h1>
      {summary && (
        <label>
          Date{' '}
          <input type="date" value={date || summary.date} onChange={(e) => setDate(e.target.value)} />
        </label>
      )}
      {error && (
        <p role="alert" className="status-error">
          {error}
        </p>
      )}
      <div className="card-grid">
        {summary?.meals.map((meal) => (
          <section key={meal.mealType} aria-label={mealLabels[meal.mealType]} className="card">
            <h2>{mealLabels[meal.mealType]}</h2>
            <p>{Math.round(meal.subtotal.caloriesKcal)} kcal</p>
            <ul>
              {meal.entries.map((entry) => (
                <li key={entry.id}>
                  {`${entry.foodName}, ${entry.portionGrams} g, ${Math.round(entry.nutrition.caloriesKcal)} kcal`}
                </li>
              ))}
            </ul>
            {/* shortcut: disabled until US-09 adds food search, enable it there */}
            <button type="button" disabled>
              Add Food
            </button>
          </section>
        ))}
      </div>
    </section>
  );
}
