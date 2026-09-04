/// <reference types="vitest/config" />
/// <reference types="vite/client" />

import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv } from 'vite'

export default ({ mode }: { mode: string }) => {
  process.env = { ...process.env, ...loadEnv(mode, process.cwd()) }

  return defineConfig({
    base: process.env.VITE_APP_BASE_URL || '',
    plugins: [react()],
    resolve: {
      tsconfigPaths: true,
    },
    server: {
      open: 'index.dev.html',
      port: 3000,
    },
    build: {
      emptyOutDir: true,
      outDir: './static',
    },
    test: {
      environment: 'jsdom',
      globals: true,
      setupFiles: './test/setup.ts',
    },
  })
}
