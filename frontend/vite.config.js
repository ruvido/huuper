import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import { readFileSync } from 'fs'
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))
const version = readFileSync(resolve(__dirname, '../VERSION'), 'utf-8').trim()
const buildDate = new Date().toISOString()
const fastBuild = process.env.FAST_BUILD === '1'

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte()],
  define: {
    __APP_VERSION__: JSON.stringify(version),
    __BUILD_DATE__: JSON.stringify(buildDate)
  },
  build: {
    outDir: '../pb_public',
    emptyOutDir: true,
    reportCompressedSize: false,
    minify: fastBuild ? false : 'esbuild',
    cssMinify: !fastBuild,
  }
})
