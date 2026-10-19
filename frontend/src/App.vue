<script setup lang="ts">
import { ref } from 'vue'

import { translate } from './messages'

// 后端应答统一为 { service, status, code, params? }（见 backend/README.md）。
interface BackendResponse {
  service: string
  status: string
  code: string
  params?: Record<string, unknown>
}

const greeting = ref('')
const errorMessage = ref('')
const isLoading = ref(false)

// 点击按钮走“前端 → 数据管理 → 数据处理”，再把后端返回的 code 翻成当前语言显示。
async function pingComputeLayer(): Promise<void> {
  isLoading.value = true
  greeting.value = ''
  errorMessage.value = ''

  try {
    const response = await fetch('/api/ping-compute')
    const payload = (await response.json()) as BackendResponse
    const text = translate(payload.code, payload.params)
    if (response.ok && payload.status === 'ok') {
      greeting.value = text
    } else {
      errorMessage.value = text
    }
  } catch (error) {
    // 技术细节只打印到控制台，界面上只显示用户看得懂的一句话。
    console.error(error)
    errorMessage.value = translate('request_failed')
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <main class="page">
    <h1 class="page-title">{{ translate('page_title') }}</h1>
    <p class="page-intro">{{ translate('page_intro') }}</p>

    <button class="ping-button" :disabled="isLoading" @click="pingComputeLayer">
      {{ isLoading ? translate('ping_button_busy') : translate('ping_button_idle') }}
    </button>

    <p v-if="isLoading" class="hint">{{ translate('ping_waiting') }}</p>
    <p v-if="greeting" class="success" role="status">{{ greeting }}</p>
    <p v-if="errorMessage" class="error" role="alert">{{ errorMessage }}</p>
  </main>
</template>

<style scoped>
.page {
  max-width: 42rem;
  margin: 4rem auto;
  padding: 0 1.5rem;
  font-family: 'Microsoft YaHei', 'PingFang SC', system-ui, sans-serif;
  color: #1f2933;
}

.page-title {
  margin-bottom: 0.5rem;
  font-size: 1.6rem;
  color: #0b5d8a;
}

.page-intro {
  margin-top: 0;
  line-height: 1.7;
  color: #52606d;
}

.ping-button {
  margin-top: 1.5rem;
  padding: 0.7rem 1.4rem;
  font-size: 1rem;
  color: #ffffff;
  background-color: #0b5d8a;
  border: none;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: background-color 0.2s;
}

.ping-button:hover:not(:disabled) {
  background-color: #094a6e;
}

.ping-button:disabled {
  background-color: #9aa5b1;
  cursor: progress;
}

.hint {
  margin-top: 1rem;
  color: #52606d;
}

.success {
  margin-top: 1rem;
  padding: 0.75rem 1rem;
  color: #014d40;
  background-color: #e3f9e5;
  border-radius: 0.5rem;
}

.error {
  margin-top: 1rem;
  padding: 0.75rem 1rem;
  color: #8b1111;
  background-color: #ffe3e3;
  border-radius: 0.5rem;
}
</style>
