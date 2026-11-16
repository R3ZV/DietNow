import { fireEvent, screen, within } from '@testing-library/react';
import { mockFetch, renderPage } from '../test/helpers';
import { TrackerPage } from './TrackerPage';

/** A /api/daily-summary response: one 80 g, 300 kcal entry at breakfast, the other meals empty. */
function summary(date: string, foodName = 'Oatmeal') {
  return {
    date,
    meals: [
      {
        mealType: 'breakfast',
        entries: [{ id: 1, foodName, portionGrams: 80, nutrition: { caloriesKcal: 300 } }],
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

  it('shows the meals of the date picked in the date selector', async () => {
    mockFetch(200, summary('2026-11-16'));
    renderPage('/tracker', <TrackerPage />);
    const dateInput = await screen.findByLabelText('Date');
    expect(dateInput).toHaveValue('2026-11-16');

    const fetchMock = mockFetch(200, summary('2026-11-15', 'Eggs'));
    fireEvent.change(dateInput, { target: { value: '2026-11-15' } });

    expect(await screen.findByText('Eggs, 80 g, 300 kcal')).toBeInTheDocument();
    expect(screen.getByLabelText('Date')).toHaveValue('2026-11-15');
    expect(fetchMock).toHaveBeenCalledWith('/api/daily-summary?date=2026-11-15', expect.anything());
  });

  it('sends a logged-out visitor to the login page', async () => {
    mockFetch(401, { error: { code: 'unauthenticated', message: 'authentication required' } });
    renderPage('/tracker', <TrackerPage />);

    expect(await screen.findByTestId('current-path')).toHaveTextContent('/login');
  });
});
