import react from '@vitejs/plugin-react';
import { defineConfig, loadEnv } from 'vite';

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  // Reads frontend/.env, .env.local, .env.[mode] (see .env.example).
  const env = loadEnv(mode, '.', 'VITE_');
  const apiTarget = env.VITE_API_PROXY_TARGET || 'http://localhost:8080';

  return {
    plugins: [react()],
    server: {
      port: 5173,
      // Forward /api/* to the Go server so the browser only talks to one origin.
      // Keep the Host header (no changeOrigin): the API blocks a POST whose
      // Origin differs from the Host it was sent to.
      proxy: {
        '/api': { target: apiTarget },
      },
    },
    test: {
      environment: 'jsdom',
      globals: true,
      setupFiles: ['./src/test/setup.ts'],
    },
  };
});
