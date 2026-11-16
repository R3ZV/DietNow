import { fireEvent, screen, within } from '@testing-library/react';
import { mockFetch, renderPage } from '../test/helpers';
import { TrackerPage } from './TrackerPage';

/** A /api/daily-summary response: one 80 g, 300 kcal entry at breakfast, the other meals empty. */
function summary(date: string, foodName = 'Oatmeal') {
  return {
    date,
    consumed: { caloriesKcal: 300 },
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

  it('clears an old error once another day loads', async () => {
    mockFetch(200, summary('2026-11-16'));
    renderPage('/tracker', <TrackerPage />);
    const dateInput = await screen.findByLabelText('Date');

    mockFetch(500, { error: { code: 'internal', message: 'something went wrong' } });
    fireEvent.change(dateInput, { target: { value: '2026-11-15' } });
    expect(await screen.findByRole('alert')).toHaveTextContent('something went wrong');

    mockFetch(200, summary('2026-11-14', 'Eggs'));
    fireEvent.change(dateInput, { target: { value: '2026-11-14' } });
    expect(await screen.findByText('Eggs, 80 g, 300 kcal')).toBeInTheDocument();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('sends a logged-out visitor to the login page', async () => {
    mockFetch(401, { error: { code: 'unauthenticated', message: 'authentication required' } });
    renderPage('/tracker', <TrackerPage />);

    expect(await screen.findByTestId('current-path')).toHaveTextContent('/login');
  });
});

describe('adding food from the tracker (US-09)', () => {
  it('opens the food search under the meal whose Add Food was tapped', async () => {
    mockFetch(200, summary('2026-11-16'));
    renderPage('/tracker', <TrackerPage />);
    const lunch = await screen.findByRole('region', { name: 'Lunch' });

    fireEvent.click(within(lunch).getByRole('button', { name: 'Add Food' }));
    mockFetch(200, {
      foods: [
        {
          id: 3,
          name: 'Chicken Breast, cooked',
          nutrition: { caloriesKcal: 165, proteinG: 31, carbsG: 0, fatG: 3.6 },
        },
      ],
    });
    fireEvent.change(within(lunch).getByLabelText('Search food'), { target: { value: 'Chicken breast' } });
    fireEvent.click(within(lunch).getByRole('button', { name: 'Search' }));

    expect(
      await within(lunch).findByText(
        'Chicken Breast, cooked: 165 kcal, 31 g protein, 0 g carbs, 3.6 g fat per 100 g',
      ),
    ).toBeInTheDocument();
    // Only the tapped meal gets a search box.
    expect(screen.getAllByLabelText('Search food')).toHaveLength(1);
  });
});

describe('logging portions (US-10)', () => {
  it("shows the day's calorie total", async () => {
    mockFetch(200, summary('2026-11-16'));
    renderPage('/tracker', <TrackerPage />);

    expect(await screen.findByText('Total: 300 kcal')).toBeInTheDocument();
  });

  it('logs 80 g of oatmeal at breakfast and updates the subtotal and the total (scenario 1)', async () => {
    mockFetch(200, emptyDay('2026-11-16'));
    renderPage('/tracker', <TrackerPage />);
    const breakfast = await screen.findByRole('region', { name: 'Breakfast' });

    fireEvent.click(within(breakfast).getByRole('button', { name: 'Add Food' }));
    mockFetch(200, {
      foods: [{ id: 1, name: 'Oatmeal', nutrition: { caloriesKcal: 375, proteinG: 13, carbsG: 68, fatG: 7 } }],
    });
    fireEvent.change(within(breakfast).getByLabelText('Search food'), { target: { value: 'Oatmeal' } });
    fireEvent.click(within(breakfast).getByRole('button', { name: 'Search' }));
    fireEvent.click(await within(breakfast).findByRole('button', { name: /^Oatmeal:/ }));

    // From here on the backend answers both the POST and the reload with the updated day.
    const fetchMock = mockFetch(200, summary('2026-11-16'));
    fireEvent.change(within(breakfast).getByLabelText('Grams of Oatmeal'), { target: { value: '80' } });
    fireEvent.click(within(breakfast).getByRole('button', { name: 'Add' }));

    expect(await within(breakfast).findByText('Oatmeal, 80 g, 300 kcal')).toBeInTheDocument();
    expect(within(breakfast).getByText('300 kcal')).toBeInTheDocument();
    expect(screen.getByText('Total: 300 kcal')).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/meals',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ date: '2026-11-16', mealType: 'breakfast', foodId: 1, grams: 80 }),
      }),
    );
    expect(screen.queryByLabelText('Search food')).not.toBeInTheDocument();
  });
});

/** A /api/daily-summary response with nothing logged. */
function emptyDay(date: string) {
  return {
    date,
    consumed: { caloriesKcal: 0 },
    meals: ['breakfast', 'lunch', 'dinner', 'snack'].map((mealType) => ({
      mealType,
      entries: [],
      subtotal: { caloriesKcal: 0 },
    })),
  };
}
