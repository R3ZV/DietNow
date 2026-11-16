import { useEffect, useState } from 'react';
import { getDailySummary, type DailySummary, type MealType } from '../api/tracker';

const mealLabels: Record<MealType, string> = {
  breakfast: 'Breakfast',
  lunch: 'Lunch',
  dinner: 'Dinner',
  snack: 'Snacks',
};

/** The daily tracker (US-08): the day's four meals, each with its foods and calorie subtotal. */
export function TrackerPage() {
  const [summary, setSummary] = useState<DailySummary | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    getDailySummary()
      .then((day) => !cancelled && setSummary(day))
      .catch((err: unknown) => {
        if (!cancelled) setError(err instanceof Error ? err.message : String(err));
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <section>
      <h1>Daily Tracker</h1>
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
