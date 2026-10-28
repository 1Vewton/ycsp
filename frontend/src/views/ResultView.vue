<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import { faArrowLeft } from '@fortawesome/free-solid-svg-icons/faArrowLeft'
import { faBan } from '@fortawesome/free-solid-svg-icons/faBan'
import { faCircleCheck } from '@fortawesome/free-solid-svg-icons/faCircleCheck'
import { faCircleXmark } from '@fortawesome/free-solid-svg-icons/faCircleXmark'
import { faHourglassHalf } from '@fortawesome/free-solid-svg-icons/faHourglassHalf'
import { faRotateRight } from '@fortawesome/free-solid-svg-icons/faRotateRight'
import { faSpinner } from '@fortawesome/free-solid-svg-icons/faSpinner'
import { faXmark } from '@fortawesome/free-solid-svg-icons/faXmark'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import StructureViewer from '../components/StructureViewer.vue'
import {
  cancelTask,
  fetchTask,
  fetchTaskResult,
  problemFromError,
  type StructureResult,
  type TaskProblem,
  type TaskSnapshot,
  type TaskStatus,
} from '../api/tasks'
import { t } from '../locales'

// 计算通常一百多毫秒；一秒刷新一次足够，也不会给后端添压力。
const POLL_INTERVAL_MS = 1000

// 三种正常状态的文字，直接复用语言包里已有的任务状态文案。
const STEP_TAGS = ['task_status_pending', 'task_status_running', 'task_status_succeeded'] as const

const STATUS_TAGS: Record<TaskStatus, string> = {
  pending: 'task_status_pending',
  running: 'task_status_running',
  succeeded: 'task_status_succeeded',
  failed: 'task_status_failed',
  cancelled: 'task_status_cancelled',
}

// 每种状态配一个图标，学生扫一眼就知道进行到哪一步了。
const STATUS_ICONS: Record<TaskStatus, IconDefinition> = {
  pending: faHourglassHalf,
  running: faSpinner,
  succeeded: faCircleCheck,
  failed: faCircleXmark,
  cancelled: faBan,
}

const route = useRoute()

const rawTaskId = computed(() => String(route.params.taskId ?? ''))
const taskId = computed(() => Number(rawTaskId.value))
const hasValidTaskId = computed(() => Number.isInteger(taskId.value) && taskId.value > 0)

const task = ref<TaskSnapshot | null>(null)
const loadFailure = ref<TaskProblem | null>(null)
const cancelFailure = ref<TaskProblem | null>(null)
const isCancelling = ref(false)
const structure = ref<StructureResult | null>(null)
const structureFailure = ref<TaskProblem | null>(null)
const isLoadingStructure = ref(false)

let pollTimer: ReturnType<typeof setTimeout> | null = null
// 离开页面后不再写状态、不再排下一次轮询（请求本身无法取消，丢掉结果即可）。
let isStopped = false

const statusTag = computed(() => (task.value === null ? '' : STATUS_TAGS[task.value.status]))
const statusIcon = computed(() => (task.value === null ? null : STATUS_ICONS[task.value.status]))
const isWaiting = computed(
  () => task.value?.status === 'pending' || task.value?.status === 'running',
)
const showSteps = computed(
  () => task.value !== null && task.value.status !== 'failed' && task.value.status !== 'cancelled',
)

const stageIndex = computed(() => {
  switch (task.value?.status) {
    case 'pending':
      return 0
    case 'running':
      return 1
    default:
      // succeeded 走完三步；还没拿到任务时不会显示步骤。
      return 2
  }
})

const durationSeconds = computed(() =>
  task.value?.duration_ms == null ? null : (task.value.duration_ms / 1000).toFixed(2),
)

function stepClass(index: number): string {
  if (task.value?.status === 'succeeded' || index < stageIndex.value) {
    return 'task-step-done'
  }
  if (index === stageIndex.value) {
    return 'task-step-active'
  }
  return ''
}

async function loadTask(): Promise<void> {
  try {
    const snapshot = await fetchTask(taskId.value)
    if (isStopped) {
      return
    }
    task.value = snapshot
    loadFailure.value = null

    if (snapshot.status === 'pending' || snapshot.status === 'running') {
      scheduleRefresh()
      return
    }
    stopPolling()
    if (snapshot.status === 'succeeded') {
      void loadStructure()
    }
  } catch (error) {
    if (isStopped) {
      return
    }
    stopPolling()
    loadFailure.value = problemFromError(error)
  }
}

// 算完后单独取结构结果；失败时保留重试按钮，不把整页判成失败。
async function loadStructure(): Promise<void> {
  if (task.value === null || task.value.status !== 'succeeded') {
    return
  }
  isLoadingStructure.value = true
  structureFailure.value = null
  try {
    const result = await fetchTaskResult(task.value.id)
    if (!isStopped) {
      structure.value = result
    }
  } catch (error) {
    if (!isStopped) {
      structureFailure.value = problemFromError(error)
    }
  } finally {
    isLoadingStructure.value = false
  }
}

function scheduleRefresh(): void {
  stopPolling()
  if (isStopped) {
    return
  }
  pollTimer = setTimeout(() => {
    void loadTask()
  }, POLL_INTERVAL_MS)
}

function stopPolling(): void {
  if (pollTimer !== null) {
    clearTimeout(pollTimer)
    pollTimer = null
  }
}

function startWatching(): void {
  stopPolling()
  task.value = null
  loadFailure.value = null
  cancelFailure.value = null
  structure.value = null
  structureFailure.value = null
  isLoadingStructure.value = false

  if (!hasValidTaskId.value) {
    loadFailure.value = { code: 'task_not_found', params: { task_id: rawTaskId.value } }
    return
  }
  void loadTask()
}

async function cancelCurrentTask(): Promise<void> {
  if (task.value === null || task.value.status !== 'pending') {
    return
  }
  isCancelling.value = true
  cancelFailure.value = null
  try {
    const snapshot = await cancelTask(task.value.id)
    if (isStopped) {
      return
    }
    stopPolling()
    task.value = snapshot
  } catch (error) {
    if (isStopped) {
      return
    }
    cancelFailure.value = problemFromError(error)
    // 取消与开始计算同时发生时，重新拉一次状态，页面才不会停在旧状态。
    void loadTask()
  } finally {
    isCancelling.value = false
  }
}

function retryLoading(): void {
  if (!hasValidTaskId.value) {
    return
  }
  loadFailure.value = null
  void loadTask()
}

onMounted(startWatching)
onBeforeUnmount(() => {
  isStopped = true
  stopPolling()
})
watch(taskId, startWatching)
</script>

<template>
  <section class="page">
    <h1 class="page-title">{{ t('result_heading') }}</h1>
    <p class="page-intro">{{ t('result_intro') }}</p>

    <div v-if="task" class="task-header">
      <p class="task-formula">{{ task.formula }}</p>
      <span class="task-status" :class="`task-status-${task.status}`" aria-live="polite">
        <FontAwesomeIcon
          v-if="statusIcon"
          :icon="statusIcon"
          :spin="task.status === 'running'"
          aria-hidden="true"
        />
        <span>{{ t(statusTag) }}</span>
      </span>
    </div>
    <p class="task-label">{{ t('result_task_label', { task_id: task?.id ?? rawTaskId }) }}</p>
    <p v-if="task === null && loadFailure === null" class="task-hint">
      {{ t('result_loading') }}
    </p>

    <section v-if="task" class="task-card">
      <ol v-if="showSteps" class="task-steps">
        <li
          v-for="(tag, index) in STEP_TAGS"
          :key="tag"
          class="task-step"
          :class="stepClass(index)"
          :aria-current="index === stageIndex ? 'step' : undefined"
        >
          {{ t(tag) }}
        </li>
      </ol>

      <p v-if="isWaiting" class="task-hint">{{ t('result_polling_hint') }}</p>
      <p v-if="task.status === 'failed' && task.error" class="status-error" role="alert">
        {{ t(task.error.code, task.error.params) }}
      </p>
      <p v-if="task.status === 'failed'" class="task-hint">{{ t('result_failed_hint') }}</p>
      <p v-if="task.status === 'cancelled'" class="task-hint">{{ t('result_cancelled_hint') }}</p>
      <p v-if="durationSeconds !== null" class="task-hint">
        {{ t('result_duration_seconds', { seconds: durationSeconds }) }}
      </p>

      <button
        v-if="task.status === 'pending'"
        type="button"
        class="secondary-button cancel-button"
        :disabled="isCancelling"
        @click="cancelCurrentTask"
      >
        <FontAwesomeIcon :icon="faXmark" aria-hidden="true" />
        <span>{{ isCancelling ? t('result_cancelling') : t('result_cancel_button') }}</span>
      </button>
      <p v-if="cancelFailure" class="status-error" role="alert">
        {{ t(cancelFailure.code, cancelFailure.params) }}
      </p>
    </section>

    <template v-if="loadFailure">
      <p class="status-error" role="alert">{{ t(loadFailure.code, loadFailure.params) }}</p>
      <button
        v-if="hasValidTaskId"
        type="button"
        class="secondary-button retry-button"
        @click="retryLoading"
      >
        <FontAwesomeIcon :icon="faRotateRight" aria-hidden="true" />
        <span>{{ t('result_retry_button') }}</span>
      </button>
    </template>

    <section v-if="task?.status === 'succeeded'" class="structure-section">
      <h2 class="section-title">{{ t('structure_title') }}</h2>
      <p v-if="isLoadingStructure" class="task-hint">{{ t('structure_loading') }}</p>
      <template v-else-if="structureFailure">
        <p class="status-error" role="alert">
          {{ t(structureFailure.code, structureFailure.params) }}
        </p>
        <button type="button" class="secondary-button structure-retry" @click="loadStructure">
          <FontAwesomeIcon :icon="faRotateRight" aria-hidden="true" />
          <span>{{ t('result_retry_button') }}</span>
        </button>
      </template>
      <StructureViewer v-else-if="structure" :result="structure" />
    </section>

    <div class="result-actions">
      <RouterLink class="primary-button" to="/predict">
        <FontAwesomeIcon :icon="faRotateRight" aria-hidden="true" />
        <span>{{ t('predict_another') }}</span>
      </RouterLink>
      <RouterLink class="link-button" to="/">
        <FontAwesomeIcon :icon="faArrowLeft" aria-hidden="true" />
        <span>{{ t('back_home') }}</span>
      </RouterLink>
    </div>
  </section>
</template>

<style scoped>
.task-header {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 1rem;
  align-items: baseline;
}

.task-formula {
  margin: 0;
  font-size: 1.35rem;
  font-weight: 600;
}

.task-status {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.1rem 0.6rem;
  font-size: 0.85rem;
  color: var(--color-primary);
  background-color: var(--color-primary-soft);
  border-radius: var(--radius-pill);
}

.task-status-succeeded {
  color: var(--color-success-text);
  background-color: var(--color-success-background);
}

.task-status-failed {
  color: var(--color-error-text);
  background-color: var(--color-error-background);
}

.task-status-cancelled {
  color: var(--color-text-muted);
  background-color: var(--color-background);
}

.task-label {
  margin: 0.25rem 0 0;
  font-size: 0.9rem;
  color: var(--color-text-muted);
}

.task-card {
  margin-top: 1rem;
  padding: 1.25rem;
  background-color: var(--color-surface);
  border-radius: var(--radius-small);
}

.task-steps {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.task-step {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.3rem 0.8rem;
  font-size: 0.9rem;
  color: var(--color-text-muted);
  background-color: var(--color-background);
  border-radius: var(--radius-pill);
}

.task-step-active {
  color: #ffffff;
  background-color: var(--color-primary);
}

.task-step-done {
  color: var(--color-success-text);
  background-color: var(--color-success-background);
}

.task-hint {
  margin: 0.75rem 0 0;
  font-size: 0.925rem;
  color: var(--color-text-muted);
}

.cancel-button,
.retry-button {
  margin-top: 1rem;
}

.structure-section {
  margin-top: 2rem;
}

.structure-retry {
  margin-top: 1rem;
}

.result-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  align-items: center;
  margin-top: 1.5rem;
}
</style>
