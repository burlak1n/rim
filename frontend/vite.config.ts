import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte()],
  server: {
    port: 80,
    host: '0.0.0.0', // Позволяет доступ с любого IP
    strictPort: true,
    proxy: {
      '/api': {
        target: 'http://localhost:3000',
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path // Убираем перезапись пути, чтобы /api/v1/contacts попадал на backend как есть
      }
    }
  }
})
