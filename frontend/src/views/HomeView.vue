<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'

import type { IconDefinition } from '@fortawesome/fontawesome-svg-core'
import { faArrowRight } from '@fortawesome/free-solid-svg-icons/faArrowRight'
import { faBookOpen } from '@fortawesome/free-solid-svg-icons/faBookOpen'
import { faFlask } from '@fortawesome/free-solid-svg-icons/faFlask'
import { faKeyboard } from '@fortawesome/free-solid-svg-icons/faKeyboard'
import { faStethoscope } from '@fortawesome/free-solid-svg-icons/faStethoscope'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import { t } from '../locales'

// 后端应答统一为 { service, status, code, params? }（见 backend/README.md）。
interface BackendResponse {
  service: string
  status: string
  code: string
  params?: Record<string, unknown>
}

// “三步上手”的说明：只存文案 tag 与图标，具体语言由语言包决定。
interface GuideStep {
  titleTag: string
  descriptionTag: string
  icon: IconDefinition
}

const guideSteps: GuideStep[] = [
  {
    titleTag: 'home_step_input_title',
    descriptionTag: 'home_step_input_description',
    icon: faKeyboard,
  },
  {
    titleTag: 'home_step_predict_title',
    descriptionTag: 'home_step_predict_description',
    icon: faFlask,
  },
  {
    titleTag: 'home_step_read_title',
    descriptionTag: 'home_step_read_description',
    icon: faBookOpen,
  },
]

const supportedFamilies: string[] = [
  'home_supported_ionic',
  'home_supported_covalent',
  'home_supported_metallic',
]

const checkMessage = ref('')
const checkError = ref('')
const isChecking = ref(false)

// 平台没反应时，学生可以先自己确认后台是否在线，不必对着空白页猜。
async function checkBackgroundServices(): Promise<void> {
  isChecking.value = true
  checkMessage.value = ''
  checkError.value = ''

  try {
    const response = await fetch('/api/ping-compute')
    const payload = (await response.json()) as BackendResponse
    const text = t(payload.code, payload.params)
    if (response.ok && payload.status === 'ok') {
      checkMessage.value = text
    } else {
      checkError.value = text
    }
  } catch (error) {
    // 技术细节只打印到控制台，界面上只显示用户看得懂的一句话。
    console.error(error)
    checkError.value = t('request_failed')
  } finally {
    isChecking.value = false
  }
}
</script>

<template>
  <section class="page">
    <h1 class="page-title">{{ t('home_heading') }}</h1>
    <p class="page-intro">{{ t('home_intro') }}</p>

    <h2 class="section-title">{{ t('home_steps_title') }}</h2>
    <ol class="step-list">
      <li v-for="(step, index) in guideSteps" :key="step.titleTag" class="step-card">
        <FontAwesomeIcon class="step-icon" :icon="step.icon" aria-hidden="true" />
        <span class="step-number" aria-hidden="true">{{ index + 1 }}</span>
        <h3 class="step-title">{{ t(step.titleTag) }}</h3>
        <p class="step-description">{{ t(step.descriptionTag) }}</p>
      </li>
    </ol>

    <h2 class="section-title">{{ t('home_supported_title') }}</h2>
    <ul class="support-list">
      <li v-for="tag in supportedFamilies" :key="tag">{{ t(tag) }}</li>
    </ul>

    <RouterLink class="primary-button" to="/predict">
      <FontAwesomeIcon :icon="faArrowRight" aria-hidden="true" />
      <span>{{ t('home_start_button') }}</span>
    </RouterLink>

    <section class="service-check">
      <h2 class="section-title">{{ t('home_check_title') }}</h2>
      <p class="page-intro">{{ t('home_check_description') }}</p>
      <button
        type="button"
        class="secondary-button"
        :disabled="isChecking"
        @click="checkBackgroundServices"
      >
        <FontAwesomeIcon :icon="faStethoscope" aria-hidden="true" />
        <span>{{ isChecking ? t('ping_button_busy') : t('ping_button_idle') }}</span>
      </button>
      <p v-if="isChecking" class="status-hint">{{ t('ping_waiting') }}</p>
      <p v-if="checkMessage" class="status-success" role="status">{{ checkMessage }}</p>
      <p v-if="checkError" class="status-error" role="alert">{{ checkError }}</p>
    </section>
  </section>
</template>

<style scoped>
.step-list {
  display: grid;
  gap: 0.75rem;
  margin: 0 0 1.5rem;
  padding: 0;
  list-style: none;
}

@media (min-width: 40rem) {
  .step-list {
    grid-template-columns: repeat(3, 1fr);
  }
}

.step-card {
  position: relative;
  padding: 1rem 3rem 1rem 1.25rem;
  background-color: var(--color-surface);
  border-radius: var(--radius-small);
}

/* 装饰图标：很淡，不抢正文，也不参与朗读。 */
.step-icon {
  position: absolute;
  top: 1rem;
  right: 1rem;
  font-size: 1.5rem;
  color: var(--color-primary);
  opacity: 0.18;
}

.step-number {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.75rem;
  height: 1.75rem;
  margin-bottom: 0.5rem;
  font-size: 0.9rem;
  font-weight: 600;
  color: #ffffff;
  background-color: var(--color-primary);
  border-radius: var(--radius-pill);
}

.step-title {
  margin: 0 0 0.35rem;
  font-size: 1rem;
}

.step-description {
  margin: 0;
  font-size: 0.95rem;
  color: var(--color-text-muted);
}

.support-list {
  margin: 0 0 1.5rem;
  padding-left: 1.2rem;
  color: var(--color-text-muted);
}

.support-list li + li {
  margin-top: 0.35rem;
}

.service-check {
  margin-top: 2.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--color-divider);
}
</style>
