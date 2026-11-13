<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'

import { faArrowRight } from '@fortawesome/free-solid-svg-icons/faArrowRight'
import { faListUl } from '@fortawesome/free-solid-svg-icons/faListUl'
import { faRotate } from '@fortawesome/free-solid-svg-icons/faRotate'
import { faSpinner } from '@fortawesome/free-solid-svg-icons/faSpinner'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import { problemFromError, type ApiProblem } from '../api/client'
import { fetchTaskHistory, type TaskHistoryItem } from '../api/tasks'
import { currentLocale, t, tCode } from '../locales'
import { isRestoringSession, restoreSession, signedInUser } from '../session'

const PAGE_SIZE = 20

const items = ref<TaskHistoryItem[]>([])
const total = ref(0)
const offset = ref(0)
const isLoading = ref(false)
const hasLoaded = ref(false)
const failure = ref<ApiProblem | null>(null)

const hasMore = computed(() => items.value.length < total.value)

onMounted(() => {
  void restoreSession()
})

// 登录状态恢复完成、或用户在页面内登录后，自动加载第一页。
watch(
  signedInUser,
  (user) => {
    if (user) {
      void loadPage(0)
      return
    }
    items.value = []
    total.value = 0
    offset.value = 0
    hasLoaded.value = false
    failure.value = null
  },
  { immediate: true },
)

async function loadPage(nextOffset: number): Promise<void> {
  if (isLoading.value) {
    return
  }
  isLoading.value = true
  failure.value = null
  try {
    const page = await fetchTaskHistory(PAGE_SIZE, nextOffset)
    items.value = nextOffset === 0 ? page.tasks : [...items.value, ...page.tasks]
    total.value = page.total
    offset.value = page.offset + page.tasks.length
    hasLoaded.value = true
  } catch (error) {
    failure.value = problemFromError(error)
  } finally {
    isLoading.value = false
  }
}

function loadMore(): void {
  if (!isLoading.value && hasMore.value) {
    void loadPage(offset.value)
  }
}

function retry(): void {
  void loadPage(offset.value)
}

function statusText(item: TaskHistoryItem): string {
  return t(`task_status_${item.status}`)
}

function resultText(item: TaskHistoryItem): string {
  if (item.result) {
    return t(`prototype_${item.result.prototype}`)
  }
  if (item.error) {
    return tCode(item.error.code, item.error.params)
  }
  return ''
}

function durationText(item: TaskHistoryItem): string {
  if (item.duration_ms === null) {
    return ''
  }
  const seconds = (item.duration_ms / 1000).toFixed(2)
  return t('result_duration_seconds', { seconds })
}

function formatDate(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  const locale = currentLocale.value === 'zh' ? 'zh-CN' : 'en'
  return new Intl.DateTimeFormat(locale, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
}
</script>


<template>
  <section class="page">
    <h1 class="page-title">{{ t('history_heading') }}</h1>
    <p class="page-intro">{{ t('history_intro') }}</p>

    <p v-if="isRestoringSession" class="status-hint" role="status">
      {{ t('auth_checking_session') }}
    </p>

    <section v-else-if="!signedInUser" class="history-card">
      <p class="history-text">{{ t('history_sign_in_hint') }}</p>
      <RouterLink
        class="primary-button"
        :to="{ name: 'login', query: { redirect: '/history' } }"
      >
        {{ t('history_sign_in_link') }}
      </RouterLink>
    </section>

    <template v-else>
      <p v-if="failure" class="status-error" role="alert">
        {{ tCode(failure.code, failure.params) }}
      </p>
      <button
        v-if="failure"
        type="button"
        class="secondary-button history-retry"
        :disabled="isLoading"
        @click="retry"
      >
        <FontAwesomeIcon :icon="isLoading ? faSpinner : faRotate" :spin="isLoading" aria-hidden="true" />
        <span>{{ t('result_retry_button') }}</span>
      </button>

      <p v-if="!hasLoaded && isLoading" class="status-hint" role="status">
        <FontAwesomeIcon :icon="faSpinner" spin aria-hidden="true" />
        <span>{{ t('history_loading_more') }}</span>
      </p>

      <section v-if="hasLoaded && items.length === 0" class="history-card">
        <p class="history-text">{{ t('history_empty') }}</p>
        <RouterLink class="primary-button" to="/predict">
          <FontAwesomeIcon :icon="faArrowRight" aria-hidden="true" />
          <span>{{ t('history_empty_link') }}</span>
        </RouterLink>
      </section>

      <section v-if="items.length > 0" class="history-list" :aria-label="t('history_heading')">
        <article v-for="item in items" :key="item.id" class="history-card">
          <div class="history-heading">
            <h2 class="history-formula">{{ item.formula }}</h2>
            <span class="history-status" :class="`history-status-${item.status}`">
              {{ statusText(item) }}
            </span>
          </div>

          <p class="history-meta">
            <span>{{ formatDate(item.created_at) }}</span>
            <span v-if="durationText(item)">· {{ durationText(item) }}</span>
          </p>

          <p v-if="resultText(item)" class="history-result">
            <FontAwesomeIcon :icon="faListUl" aria-hidden="true" />
            <span>{{ resultText(item) }}</span>
          </p>

          <RouterLink
            class="secondary-button history-open"
            :to="{ name: 'result', params: { taskId: String(item.id) } }"
          >
            <FontAwesomeIcon :icon="faArrowRight" aria-hidden="true" />
            <span>{{ t('history_open_button') }}</span>
          </RouterLink>
        </article>
      </section>

      <div v-if="hasLoaded && items.length > 0" class="history-footer">
        <p class="history-count">{{ t('history_showing', { shown: items.length, total }) }}</p>
        <button
          v-if="hasMore"
          type="button"
          class="secondary-button"
          :disabled="isLoading"
          @click="loadMore"
        >
          <FontAwesomeIcon :icon="isLoading ? faSpinner : faArrowRight" :spin="isLoading" aria-hidden="true" />
          <span>{{ isLoading ? t('history_loading_more') : t('history_load_more') }}</span>
        </button>
      </div>
    </template>
  </section>
</template>


<style scoped>
.history-card {
  margin-top: 1rem;
  padding: 1.25rem 1.5rem;
  background-color: var(--color-surface);
  border-radius: var(--radius-small);
}

.history-text {
  margin: 0 0 1rem;
  line-height: 1.7;
  color: var(--color-text-muted);
}

.history-list {
  margin-top: 0.5rem;
}

.history-heading {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 0.75rem;
  align-items: center;
  justify-content: space-between;
}

.history-formula {
  margin: 0;
  font-size: 1.25rem;
}

.history-status {
  padding: 0.15rem 0.55rem;
  font-size: 0.85rem;
  font-weight: 600;
  border-radius: var(--radius-pill);
  background-color: var(--color-primary-soft);
  color: var(--color-primary);
}

.history-status-succeeded {
  background-color: var(--color-success-background);
  color: var(--color-success-text);
}

.history-status-failed {
  background-color: var(--color-error-background);
  color: var(--color-error-text);
}

.history-status-cancelled {
  background-color: var(--color-divider);
  color: var(--color-text-muted);
}

.history-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin: 0.5rem 0 0;
  font-size: 0.9rem;
  color: var(--color-text-muted);
}

.history-result {
  display: flex;
  gap: 0.4rem;
  align-items: center;
  margin: 0.75rem 0 0;
  line-height: 1.6;
}

.history-open {
  margin-top: 1rem;
}

.history-retry {
  margin-top: 0.75rem;
}

.history-footer {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem 1rem;
  align-items: center;
  justify-content: space-between;
  margin-top: 1.25rem;
}

.history-count {
  margin: 0;
  font-size: 0.9rem;
  color: var(--color-text-muted);
}
</style>
