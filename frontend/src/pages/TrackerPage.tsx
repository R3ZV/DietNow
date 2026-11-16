import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ApiError, errorMessage } from '../api/client';
import type { Food } from '../api/foods';
import {
  addMealEntry,
  deleteMealEntry,
  getDailySummary,
  type DailySummary,
  type MealType,
  type NewMealEntry,
} from '../api/tracker';
import { FoodSearch } from '../components/FoodSearch';
import { PortionForm } from '../components/PortionForm';

const mealLabels: Record<MealType, string> = {
  breakfast: 'Breakfast',
  lunch: 'Lunch',
  dinner: 'Dinner',
  snack: 'Snacks',
};

/**
 * The daily tracker: one day's four meals, each with its foods and calorie subtotal (US-08).
 * Add Food opens a food search under that meal (US-09); picking a food asks for the grams
 * and logs the portion (US-10).
 */
export function TrackerPage() {
  const navigate = useNavigate();
  // Unset until the user picks a day; the backend then shows today in the user's time zone.
  const [date, setDate] = useState<string>();
  const [summary, setSummary] = useState<DailySummary | null>(null);
  const [error, setError] = useState<string | null>(null);
  // Bumped to reload the day after a change.
  const [reloads, setReloads] = useState(0);
  // The meal whose Add Food panel is open (one at a time), and the food picked in it.
  const [addingTo, setAddingTo] = useState<MealType | null>(null);
  const [picked, setPicked] = useState<Food | null>(null);

  useEffect(() => {
    // `cancelled` drops a slow answer for a day the user has already left.
    let cancelled = false;
    getDailySummary(date)
      .then((day) => {
        if (cancelled) return;
        setSummary(day);
        setError(null);
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        // replace: Back from the login page must not land here and redirect again.
        if (err instanceof ApiError && err.status === 401) navigate('/login', { replace: true });
        else setError(err instanceof Error ? err.message : String(err));
      });
    return () => {
      cancelled = true;
    };
  }, [date, reloads, navigate]);

  function toggleAddFood(meal: MealType) {
    setAddingTo(addingTo === meal ? null : meal);
    setPicked(null);
  }

  async function logPortion(entry: NewMealEntry) {
    await addMealEntry(entry);
    setAddingTo(null);
    setPicked(null);
    reloadDay(entry.date);
  }

  async function removeEntry(id: number, day: string) {
    try {
      await deleteMealEntry(id);
      reloadDay(day);
    } catch (err) {
      setError(errorMessage(err));
    }
  }

  function reloadDay(day: string) {
    // Pin the changed day, so the reload cannot jump to a new "today" after midnight.
    setDate(day);
    setReloads((n) => n + 1);
  }

  return (
    <section>
      <h1>Daily Tracker</h1>
      {error && (
        <p role="alert" className="status-error">
          {error}
        </p>
      )}
      {summary && (
        <>
          <label>
            Date{' '}
            <input type="date" value={date || summary.date} onChange={(e) => setDate(e.target.value)} />
          </label>
          <p>Total: {Math.round(summary.consumed.caloriesKcal)} kcal</p>
          <div className="card-grid">
            {summary.meals.map((meal) => (
              <section key={meal.mealType} aria-label={mealLabels[meal.mealType]} className="card">
                <h2>{mealLabels[meal.mealType]}</h2>
                <p>{Math.round(meal.subtotal.caloriesKcal)} kcal</p>
                <ul>
                  {meal.entries.map((entry) => (
                    <li key={entry.id}>
                      {`${entry.foodName}, ${entry.portionGrams} g, ${Math.round(entry.nutrition.caloriesKcal)} kcal`}{' '}
                      <button
                        type="button"
                        aria-label={`Delete ${entry.foodName}`}
                        onClick={() => removeEntry(entry.id, summary.date)}
                      >
                        Delete
                      </button>
                    </li>
                  ))}
                </ul>
                <button
                  type="button"
                  aria-expanded={addingTo === meal.mealType}
                  onClick={() => toggleAddFood(meal.mealType)}
                >
                  Add Food
                </button>
                {addingTo === meal.mealType &&
                  (picked ? (
                    <PortionForm
                      food={picked}
                      onSubmit={(grams) =>
                        logPortion({ date: summary.date, mealType: meal.mealType, foodId: picked.id, grams })
                      }
                    />
                  ) : (
                    <FoodSearch onSelect={setPicked} />
                  ))}
              </section>
            ))}
          </div>
        </>
      )}
    </section>
  );
}
