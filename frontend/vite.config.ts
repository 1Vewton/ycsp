import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import { defineConfig, loadEnv } from 'vite'

import { getEnv, getEnvNumber } from './env.ts'

// .env 未提供对应项时使用的默认值。
const DEFAULT_FRONTEND_PORT = 5173
const DEFAULT_BACKEND_BASE_URL = 'http://127.0.0.1:8000'

// 3Dmol 的按需 chunk 约 530 kB，超过 Vite 默认的 500 kB 提示线；给它留出余量。
const CHUNK_SIZE_WARNING_LIMIT_KB = 600

// 开发端口与后端地址来自 frontend/.env，未提供时用上面的默认值。
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [vue()],
    build: {
      // 3Dmol 单独成块、只在结果页按需下载，不影响首页与预测页的首屏。
      chunkSizeWarningLimit: CHUNK_SIZE_WARNING_LIMIT_KB,
    },
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
