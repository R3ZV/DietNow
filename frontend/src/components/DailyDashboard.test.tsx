import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { DailyDashboard } from './DailyDashboard';
import type { DailySummary } from '../api/tracker';

const day: DailySummary = {
  date: '2026-11-16', meals: [],
  consumed: { caloriesKcal: 1800, proteinG: 80, carbsG: 250, fatG: 45 },
  targets: { caloriesKcal: 3000, proteinG: 140, carbsG: 450, fatG: 70 },
  remaining: { caloriesKcal: 1200, proteinG: 60, carbsG: 200, fatG: 25 },
};
function show(summary: DailySummary) {
  return render(<MemoryRouter><DailyDashboard summary={summary} /></MemoryRouter>);
}
it('shows the acceptance example and all three macro progress bars', () => {
  show(day);
  expect(screen.getByText('1200 kcal Remaining')).toBeInTheDocument();
  expect(screen.getByText('Total: 1800 kcal')).toBeInTheDocument();
  expect(screen.getByText('Daily target: 3000 kcal')).toBeInTheDocument();
  for (const [name, value, max] of [['Protein', '80', '140'], ['Carbohydrates', '250', '450'], ['Fat', '45', '70']]) {
    const bar = screen.getByRole('progressbar', { name: `${name} progress` });
    expect(bar).toHaveAttribute('value', value);
    expect(bar).toHaveAttribute('max', max);
  }
});
it('labels over-target intake and caps the bar without hiding actual consumption', () => {
  show({ ...day, consumed: { ...day.consumed, caloriesKcal: 3100, proteinG: 150 }, remaining: { ...day.remaining!, caloriesKcal: -100, proteinG: -10 } });
  expect(screen.getByText('100 kcal over target')).toBeInTheDocument();
  expect(screen.getByText('150 / 140 g')).toBeInTheDocument();
  expect(screen.getByText('10 g over target')).toBeInTheDocument();
  expect(screen.getByRole('progressbar', { name: 'Protein progress' })).toHaveAttribute('value', '140');
});
it('keeps consumption visible and links to onboarding when targets are missing', () => {
  show({ ...day, targets: null, remaining: null });
  expect(screen.getByText('Total: 1800 kcal')).toBeInTheDocument();
  expect(screen.getByRole('link', { name: 'Complete your profile' })).toHaveAttribute('href', '/profile');
  expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
});
it('handles a zero macro target without an invalid progress range', () => {
  show({ ...day, targets: { ...day.targets!, carbsG: 0 }, remaining: { ...day.remaining!, carbsG: -250 } });
  expect(screen.getByText('No gram target set')).toBeInTheDocument();
  expect(screen.queryByRole('progressbar', { name: 'Carbohydrates progress' })).not.toBeInTheDocument();
  expect(screen.getByText('250 g over target')).toBeInTheDocument();
});
