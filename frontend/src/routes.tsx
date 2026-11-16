import type { RouteObject } from 'react-router-dom';
import { Layout } from './components/Layout';
import { ProfilePage } from './pages/ProfilePage';
import { HomePage } from './pages/HomePage';
import { LoginPage } from './pages/LoginPage';
import { NotFoundPage } from './pages/NotFoundPage';
import { PlaceholderPage } from './pages/PlaceholderPage';
import { RegisterPage } from './pages/RegisterPage';

/** Opt into React Router v7 behaviour now so the later upgrade is a no-op. */
export const routerFutureFlags = {
  v7_relativeSplatPath: true,
  v7_fetcherPersist: true,
  v7_normalizeFormMethod: true,
  v7_partialHydration: true,
  v7_skipActionErrorRevalidation: true,
} as const;

export const providerFutureFlags = { v7_startTransition: true } as const;

/**
 * Application routes. Every page renders inside the shared <Layout>.
 * Feature tasks replace the placeholder pages with real implementations.
 */
export const routes: RouteObject[] = [
  {
    path: '/',
    element: <Layout />,
    children: [
      { index: true, element: <HomePage /> },
      { path: 'register', element: <RegisterPage /> },
      { path: 'login', element: <LoginPage /> },
      {
        path: 'tracker',
        element: (
          <PlaceholderPage
            title="Daily Tracker"
            description="Breakfast, lunch, dinner and snacks with live calorie and macro totals."
          />
        ),
      },
      {
        path: 'chat',
        element: (
          <PlaceholderPage
            title="AI Nutritionist"
            description="Chat about food, get recipes from what's in your fridge, and save favorites."
          />
        ),
      },
      {
        path: 'profile',
        element: (
          <ProfilePage />
        ),
      },
      { path: '*', element: <NotFoundPage /> },
    ],
  },
];
