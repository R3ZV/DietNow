/** Food lookup (US-09), from GET /api/foods. */
import { apiFetch } from './client';

export interface Food {
  id: number;
  name: string;
  /** Per 100 g. */
  nutrition: { caloriesKcal: number; proteinG: number; carbsG: number; fatG: number };
}

/** Finds foods whose name contains `query` (2-100 characters, checked by the backend). */
export const searchFoods = (query: string) =>
  apiFetch<{ foods: Food[] }>(`/foods?q=${encodeURIComponent(query)}`);
