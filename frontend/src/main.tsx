import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { RouterProvider, createBrowserRouter } from 'react-router-dom';
import { providerFutureFlags, routerFutureFlags, routes } from './routes';
import './index.css';

const router = createBrowserRouter(routes, { future: routerFutureFlags });

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <RouterProvider router={router} future={providerFutureFlags} />
  </StrictMode>,
);
