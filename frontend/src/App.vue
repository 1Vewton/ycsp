<script setup lang="ts">
import { ref } from 'vue'

interface BackendPingResponse {
  service: string
  status: string
  message: string
}

const greetingFromCompute = ref('')
const errorMessage = ref('')
const isLoading = ref(false)

// 点击按钮走“前端 → 数据管理 → 数据处理”，把后端返回的中文提示显示出来。
async function pingComputeLayer(): Promise<void> {
  isLoading.value = true
  greetingFromCompute.value = ''
  errorMessage.value = ''

  try {
    const response = await fetch('/api/ping-compute')
    if (!response.ok) {
      throw new Error(`后端返回了状态码 ${response.status}`)
    }
    const payload = (await response.json()) as BackendPingResponse
    greetingFromCompute.value = payload.message
  } catch (error) {
    errorMessage.value = `暂时联系不上后端：${(error as Error).message}。请确认数据管理层与数据处理层都已启动。`
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <main class="page">
    <h1 class="page-title">ycsp 晶体结构预测教学平台</h1>
    <p class="page-intro">
      这是阶段 0 的骨架页面，用来检查「前端 → 数据管理 → 数据处理」三层是否连通。
    </p>

    <button class="ping-button" :disabled="isLoading" @click="pingComputeLayer">
      {{ isLoading ? '正在联系……' : '和数据处理层打个招呼' }}
    </button>

    <p v-if="isLoading" class="hint">正在等待后端应答，请稍候……</p>
    <p v-if="greetingFromCompute" class="success" role="status">
      {{ greetingFromCompute }}
    </p>
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
