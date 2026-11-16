import { useEffect, useState, type FormEvent } from 'react';
import { ApiError, apiFetch } from '../api/client';
import './ProfilePage.css';

type Intake = {
  gender: string; birthDate: string; heightCm: string; currentWeightKg: string; activityLevel: string;
};
type Profile = Omit<Intake, 'heightCm' | 'currentWeightKg'> & {
  heightCm: number; currentWeightKg: number; timezone: string;
  goal: string; targetWeightKg: number; timeframeWeeks: number;
};
const empty: Intake = { gender: '', birthDate: '', heightCm: '', currentWeightKg: '', activityLevel: '' };
const activities = [['sedentary', 'Sedentary'], ['light', 'Lightly active'], ['moderate', 'Moderately active'], ['very', 'Very active'], ['extra', 'Extra active']];

export function ProfilePage() {
  const [form, setForm] = useState<Intake>(empty);
  const [profile, setProfile] = useState<Profile>();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [message, setMessage] = useState('');
  const [saved, setSaved] = useState(false);
  const [loadFailed, setLoadFailed] = useState(false);
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    const controller = new AbortController();
    setLoading(true); setLoadFailed(false); setMessage('');
    apiFetch<{ profile: Profile }>('/profile', { signal: controller.signal })
      .then(({ profile: p }) => {
        setProfile(p);
        setForm({ gender: p.gender, birthDate: p.birthDate, heightCm: String(p.heightCm), currentWeightKg: String(p.currentWeightKg), activityLevel: p.activityLevel });
      })
      .catch((error: unknown) => {
        if (controller.signal.aborted) return;
        if (error instanceof ApiError && error.status === 404 && error.code === 'profile_not_found') return;
        setLoadFailed(true);
        setMessage(error instanceof ApiError && error.status === 401 ? 'Please sign in before completing your profile.' : 'Could not load your profile. Please try again.');
      })
      .finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, [attempt]);

  function update(key: keyof Intake, value: string) {
    setForm(current => ({ ...current, [key]: value }));
    setErrors(current => ({ ...current, [key]: '' }));
    setSaved(false); setMessage('');
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    const invalid: Record<string, string> = {};
    if (!['male', 'female'].includes(form.gender)) invalid.gender = 'Select a gender.';
    const birth = new Date(`${form.birthDate}T00:00:00`);
    const today = new Date();
    let age = today.getFullYear() - birth.getFullYear();
    if (today.getMonth() < birth.getMonth() || (today.getMonth() === birth.getMonth() && today.getDate() < birth.getDate())) age--;
    if (!form.birthDate || Number.isNaN(birth.getTime()) || age < 13 || age > 100) invalid.birthDate = 'Enter a birth date for an age between 13 and 100.';
    if (!form.heightCm || Number(form.heightCm) < 100 || Number(form.heightCm) > 250) invalid.heightCm = 'Enter a height between 100 and 250 cm.';
    if (!form.currentWeightKg || Number(form.currentWeightKg) < 30 || Number(form.currentWeightKg) > 300) invalid.currentWeightKg = 'Enter a weight between 30 and 300 kg.';
    if (!activities.some(([value]) => value === form.activityLevel)) invalid.activityLevel = 'Select an activity level.';
    setErrors(invalid); setSaved(false); setMessage('');
    if (Object.keys(invalid).length) return;
    setSaving(true);
    try {
      const result = await apiFetch<{ profile: Profile }>('/profile', {
        method: 'PUT', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...form, heightCm: Number(form.heightCm), currentWeightKg: Number(form.currentWeightKg),
          timezone: profile?.timezone || Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC',
          goal: profile?.goal || 'maintain', targetWeightKg: profile?.targetWeightKg || 0, timeframeWeeks: profile?.timeframeWeeks || 0 }),
      });
      setProfile(result.profile); setSaved(true);
    } catch (error) {
      if (error instanceof ApiError) {
        setErrors(error.fields);
        setMessage(error.status === 401 ? 'Your session has expired. Please sign in again.' : Object.keys(error.fields).length ? 'Please correct the highlighted fields.' : error.message);
      } else setMessage('Could not save your profile. Your entries are still here; please try again.');
    } finally { setSaving(false); }
  }

  const field = (key: keyof Intake, label: string, type: string, extra = {}) => (
    <div className="intake-field"><label htmlFor={key}>{label}</label>
      <input id={key} type={type} value={form[key]} onChange={e => update(key, e.target.value)} aria-invalid={!!errors[key]} aria-describedby={errors[key] ? `${key}-error` : undefined} required {...extra} />
      {errors[key] && <span className="intake-error" id={`${key}-error`}>{errors[key]}</span>}
    </div>
  );
  const select = (key: 'gender' | 'activityLevel', label: string, options: string[][]) => (
    <div className="intake-field"><label htmlFor={key}>{label}</label>
      <select id={key} value={form[key]} onChange={e => update(key, e.target.value)} aria-invalid={!!errors[key]} aria-describedby={errors[key] ? `${key}-error` : undefined} required>
        <option value="">Choose an option</option>{options.map(([value, text]) => <option key={value} value={value}>{text}</option>)}
      </select>
      {errors[key] && <span className="intake-error" id={`${key}-error`}>{errors[key]}</span>}
    </div>
  );
  return <section className="intake">
    <h1>Your profile</h1>
    <p className="lead">Tell us about yourself to calculate your daily calorie and macro targets.</p>
    {loading ? <p role="status">Loading your profile…</p> : <>
      {message && <p role="alert" className="intake-error">{message}</p>}
      {loadFailed ? <button onClick={() => setAttempt(n => n + 1)}>Try again</button> :
        <form onSubmit={submit} noValidate>
          <fieldset disabled={saving}>
            <legend>Biometric details</legend>
            <div className="intake-grid">
              {select('gender', 'Gender', [['female', 'Female'], ['male', 'Male']])}
              {field('birthDate', 'Date of birth', 'date')}
              {field('heightCm', 'Height (cm)', 'number', { min: 100, max: 250, step: 'any' })}
              {field('currentWeightKg', 'Current weight (kg)', 'number', { min: 30, max: 300, step: 'any' })}
              {select('activityLevel', 'Activity level', activities)}
            </div>
            <p className="muted">Your age is calculated from your date of birth. New profiles start with a maintenance calorie target.</p>
            <button type="submit">{saving ? 'Saving…' : 'Save profile'}</button>
          </fieldset>
          {saved && <p role="status">Profile saved. Your calorie and macro targets have been updated.</p>}
        </form>}
    </>}
  </section>;
}
