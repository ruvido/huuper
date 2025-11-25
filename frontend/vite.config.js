import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte()],
  build: {
    outDir: '../pb_public',
    emptyOutDir: true,
  },
  optimizeDeps: {
    exclude: ['@jsquash/jpeg', '@jsquash/resize']
  }
})
