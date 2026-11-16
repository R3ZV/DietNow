import { Link } from 'react-router-dom';
import type { DailySummary } from '../api/tracker';
import './DailyDashboard.css';

const macros = [
  ['proteinG', 'Protein'], ['carbsG', 'Carbohydrates'], ['fatG', 'Fat'],
] as const;
const grams = (value: number) => Number(value.toFixed(1));

/** All totals come from the same response as the meals, including after add/delete. */
export function DailyDashboard({ summary }: { summary: DailySummary }) {
  const { consumed, targets, remaining } = summary;
  return <section className="daily-dashboard" aria-labelledby="daily-dashboard-title">
    <h2 id="daily-dashboard-title">Daily nutrition</h2>
    <div className="calorie-summary" aria-live="polite" aria-atomic="true">
      <p>Total: {Math.round(consumed.caloriesKcal)} kcal</p>
      {targets && remaining && <>
        <p>Daily target: {Math.round(targets.caloriesKcal)} kcal</p>
        <p className="calorie-budget">
          {Math.round(Math.abs(remaining.caloriesKcal))} kcal {remaining.caloriesKcal < 0 ? 'over target' : 'Remaining'}
        </p>
      </>}
    </div>
    {!targets || !remaining ? <p><Link to="/profile">Complete your profile</Link> to see your daily targets and remaining nutrition.</p> :
      <div className="macro-grid">{macros.map(([key, label]) => {
        const target = targets[key];
        const amount = consumed[key];
        return <div className="macro-card" key={key}>
          <h3>{label}</h3>
          <p>{grams(amount)} / {grams(target)} g</p>
          {target > 0 ? <progress aria-label={`${label} progress`} max={target} value={Math.min(target, Math.max(0, amount))}
            aria-valuetext={`${grams(amount)} of ${grams(target)} grams consumed`} /> : <p className="muted">No gram target set</p>}
          <p className={remaining[key] < 0 ? 'status-error' : 'muted'}>
            {grams(Math.abs(remaining[key]))} g {remaining[key] < 0 ? 'over target' : 'remaining'}
          </p>
        </div>;
      })}</div>}
  </section>;
}
