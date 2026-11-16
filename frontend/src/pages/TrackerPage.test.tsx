import { screen, within } from '@testing-library/react';
import { mockFetch, renderPage } from '../test/helpers';
import { TrackerPage } from './TrackerPage';

/** A /api/daily-summary response: one 80 g oatmeal entry at breakfast, the other meals empty. */
function summary(date: string) {
  return {
    date,
    meals: [
      {
        mealType: 'breakfast',
        entries: [{ id: 1, foodName: 'Oatmeal', portionGrams: 80, nutrition: { caloriesKcal: 300 } }],
        subtotal: { caloriesKcal: 300 },
      },
      { mealType: 'lunch', entries: [], subtotal: { caloriesKcal: 0 } },
      { mealType: 'dinner', entries: [], subtotal: { caloriesKcal: 0 } },
      { mealType: 'snack', entries: [], subtotal: { caloriesKcal: 0 } },
    ],
  };
}

describe('tracker page (US-08)', () => {
  it("shows today's four meals, each with its calorie subtotal and an Add Food button", async () => {
    const fetchMock = mockFetch(200, summary('2026-11-16'));
    renderPage('/tracker', <TrackerPage />);

    const breakfast = await screen.findByRole('region', { name: 'Breakfast' });
    expect(within(breakfast).getByText('300 kcal')).toBeInTheDocument();
    expect(within(breakfast).getByText('Oatmeal, 80 g, 300 kcal')).toBeInTheDocument();

    for (const meal of ['Breakfast', 'Lunch', 'Dinner', 'Snacks']) {
      const section = screen.getByRole('region', { name: meal });
      expect(within(section).getByRole('button', { name: 'Add Food' })).toBeInTheDocument();
    }
    for (const meal of ['Lunch', 'Dinner', 'Snacks']) {
      expect(within(screen.getByRole('region', { name: meal })).getByText('0 kcal')).toBeInTheDocument();
    }
    // No date sent: the backend picks "today" in the user's own time zone.
    expect(fetchMock).toHaveBeenCalledWith('/api/daily-summary', expect.anything());
  });
});
