import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import { defineConfig, loadEnv } from 'vite'

import { getEnv, getEnvNumber } from './env.ts'

// .env 未提供对应项时使用的默认值。
const DEFAULT_FRONTEND_PORT = 5173
const DEFAULT_BACKEND_BASE_URL = 'http://127.0.0.1:8000'

// 开发端口与后端地址来自 frontend/.env，未提供时用上面的默认值。
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      port: getEnvNumber(env, 'VITE_FRONTEND_PORT', DEFAULT_FRONTEND_PORT),
      proxy: {
        '/api': {
          target: getEnv(env, 'VITE_BACKEND_BASE_URL', DEFAULT_BACKEND_BASE_URL),
          changeOrigin: true,
        },
      },
    },
  }
})
