import { fireEvent, render, screen } from '@testing-library/react';
import { vi } from 'vitest';
import { mockFetch } from '../test/helpers';
import { FoodSearch } from './FoodSearch';

const chicken = {
  id: 3,
  name: 'Chicken Breast, cooked',
  nutrition: { caloriesKcal: 165, proteinG: 31, carbsG: 0, fatG: 3.6 },
};
const chickenText = 'Chicken Breast, cooked: 165 kcal, 31 g protein, 0 g carbs, 3.6 g fat per 100 g';

function search(text: string) {
  fireEvent.change(screen.getByLabelText('Search food'), { target: { value: text } });
  fireEvent.click(screen.getByRole('button', { name: 'Search' }));
}

describe('food search (US-09)', () => {
  it('lists matching foods with calories and macros per 100 g', async () => {
    const fetchMock = mockFetch(200, { foods: [chicken] });
    render(<FoodSearch onSelect={vi.fn()} />);

    search('Chicken breast');

    expect(await screen.findByText(chickenText)).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith('/api/foods?q=Chicken%20breast', expect.anything());
  });

  it('says so when nothing matches', async () => {
    mockFetch(200, { foods: [] });
    render(<FoodSearch onSelect={vi.fn()} />);

    search('xyz');

    expect(await screen.findByText('No foods found.')).toBeInTheDocument();
  });

  it('shows nothing before the first search', () => {
    render(<FoodSearch onSelect={vi.fn()} />);

    expect(screen.queryByText('No foods found.')).not.toBeInTheDocument();
  });
});

describe('picking a food (US-10)', () => {
  it('hands the clicked food to onSelect', async () => {
    mockFetch(200, { foods: [chicken] });
    const onSelect = vi.fn();
    render(<FoodSearch onSelect={onSelect} />);

    search('Chicken breast');
    fireEvent.click(await screen.findByRole('button', { name: chickenText }));

    expect(onSelect).toHaveBeenCalledWith(chicken);
  });
});
