import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// OpenDataBuddy 是 Go 单进程部署：Vite 构建产物输出到 dist，由 Go 直接 serve。
// 开发时把 /api 代理到本地 Go 后端 8000。
export default defineConfig({
  plugins: [vue()],
  base: '/',
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    chunkSizeWarningLimit: 1500,
  },
  server: {
    host: '127.0.0.1',
    port: 3082,
    strictPort: true,
    proxy: {
      '/api': { target: 'http://127.0.0.1:8000', changeOrigin: true },
    },
  },
})
