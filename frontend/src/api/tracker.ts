/** One day of the meal tracker (US-08), from GET /api/daily-summary. */
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

/** Loads the meals of `date` (YYYY-MM-DD); without it, of today in the user's time zone. */
export const getDailySummary = (date?: string) =>
  apiFetch<DailySummary>(date ? `/daily-summary?date=${date}` : '/daily-summary');
