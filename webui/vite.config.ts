/// <reference types="vitest" />
/// <reference types="vite/client" />

import react from '@vitejs/plugin-react'
import { defineConfig, loadEnv } from 'vite'
import viteTsconfigPaths from 'vite-tsconfig-paths'

export default ({ mode }: { mode: string }) => {
  process.env = { ...process.env, ...loadEnv(mode, process.cwd()) }

  return defineConfig({
    base: process.env.VITE_APP_BASE_URL || '',
    plugins: [react(), viteTsconfigPaths()],
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
