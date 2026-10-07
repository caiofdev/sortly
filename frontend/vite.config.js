import { writeFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// O Vite esvazia dist/ a cada build; o backend Go precisa que a pasta exista
// (//go:embed all:frontend/dist), por isso o .gitkeep versionado é recriado.
function keepDistFolder() {
  return {
    name: 'sortly:keep-dist-folder',
    closeBundle() {
      writeFileSync(resolve(import.meta.dirname, 'dist/.gitkeep'), '');
    }
  };
}

export default defineConfig({
  base: './',
  plugins: [react(), keepDistFolder()],
  server: {
    port: 5173,
    strictPort: true
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./vitest.setup.js'],
    restoreMocks: true,
    coverage: {
      include: ['src/**/*.{js,jsx}'],
      exclude: ['src/main.jsx']
    }
  }
});
