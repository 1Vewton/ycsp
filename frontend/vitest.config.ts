import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vitest/config'

// 前端自动化测试：纯函数、接口封装与组件。组件需要 DOM，所以统一跑在 jsdom 里。
// 单独成文件而不是写进 vite.config.ts，测试依赖不会混进生产构建。
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  test: {
    environment: 'jsdom',
    include: ['tests/**/*.test.ts'],
    // 语言包与 fetch 都是模块级状态：每个用例前恢复 spy、清掉 stubGlobal，避免相互污染。
    restoreMocks: true,
    unstubGlobals: true,
  },
})
