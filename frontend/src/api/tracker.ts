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
  /** Always the four meals, in this order: breakfast, lunch, dinner, snack. */
  meals: { mealType: MealType; entries: MealEntry[]; subtotal: { caloriesKcal: number } }[];
}

/** Loads today's meals; the backend decides "today" in the user's time zone. */
export const getDailySummary = () => apiFetch<DailySummary>('/daily-summary');
