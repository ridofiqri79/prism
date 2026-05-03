import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'

const usePolling = process.env.VITE_USE_POLLING === 'true'

export default defineConfig({
  plugins: [tailwindcss(), vue()],
  build: {
    rolldownOptions: {
      output: {
        codeSplitting: {
          groups: [
            {
              name: 'vue-vendor',
              test: /node_modules[\\/](vue|@vue|pinia|vue-router)[\\/]/,
              maxSize: 420 * 1024,
              priority: 30,
            },
            {
              name: 'prime-vendor',
              test: /node_modules[\\/](primevue|@primeuix|primeicons)[\\/]/,
              maxSize: 420 * 1024,
              priority: 20,
            },
            {
              name: 'chart-vendor',
              test: /node_modules[\\/](echarts|zrender|vue-echarts)[\\/]/,
              maxSize: 420 * 1024,
              priority: 10,
            },
          ],
        },
      },
    },
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    host: '0.0.0.0',
    port: 5173,
    strictPort: true,
    hmr: {
      clientPort: 5173,
    },
    watch: usePolling
      ? {
          usePolling: true,
          interval: 250,
        }
      : undefined,
  },
})
