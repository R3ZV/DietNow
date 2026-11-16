/** The meal tracker: one day from GET /api/daily-summary (US-08) and logging portions (US-10). */
import { apiFetch } from './client';

export type MealType = 'breakfast' | 'lunch' | 'dinner' | 'snack';

export interface MealEntry {
  id: number;
  foodName: string;
  portionGrams: number;
  nutrition: { caloriesKcal: number };
}

export interface DailySummary {
  /** The day shown, YYYY-MM-DD. */
  date: string;
  /** The whole day, all four meals together. */
  consumed: { caloriesKcal: number };
  /** Always the four meals, in this order: breakfast, lunch, dinner, snack. */
  meals: { mealType: MealType; entries: MealEntry[]; subtotal: { caloriesKcal: number } }[];
}

export interface NewMealEntry {
  /** YYYY-MM-DD */
  date: string;
  mealType: MealType;
  foodId: number;
  grams: number;
}

/** Logs a portion; the backend scales the food's per-100 g nutrition to `grams` (US-10). */
export const addMealEntry = (entry: NewMealEntry) =>
  apiFetch('/meals', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(entry),
  });

/** Removes a logged entry (US-10). */
export const deleteMealEntry = (id: number) => apiFetch(`/meals/${id}`, { method: 'DELETE' });

/** Loads the meals of `date` (YYYY-MM-DD); without it, of today in the user's time zone. */
export const getDailySummary = (date?: string) =>
  apiFetch<DailySummary>(date ? `/daily-summary?date=${date}` : '/daily-summary');
