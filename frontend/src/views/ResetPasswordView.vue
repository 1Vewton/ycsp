<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import { faArrowLeft } from '@fortawesome/free-solid-svg-icons/faArrowLeft'
import { faArrowRight } from '@fortawesome/free-solid-svg-icons/faArrowRight'
import { faSpinner } from '@fortawesome/free-solid-svg-icons/faSpinner'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import { problemFromError, type ApiProblem } from '../api/client'
import { t, tCode } from '../locales'
import { lookupSecurityQuestion, resetPassword } from '../session'

const router = useRouter()

const step = ref<'lookup' | 'reset'>('lookup')
const displayName = ref('')
const securityQuestion = ref('')
const securityAnswer = ref('')
const newPassword = ref('')
const passwordConfirmation = ref('')
const isSubmitting = ref(false)
const failure = ref<ApiProblem | null>(null)

const passwordRules = computed(() => {
  const rules = failure.value?.params?.rules
  if (!Array.isArray(rules)) {
    return []
  }
  return rules.filter((rule): rule is string => typeof rule === 'string')
})

async function lookupQuestion(): Promise<void> {
  failure.value = null
  if (displayName.value.trim() === '') {
    failure.value = { code: 'auth_nickname_required' }
    return
  }

  isSubmitting.value = true
  try {
    securityQuestion.value = await lookupSecurityQuestion(displayName.value.trim())
    step.value = 'reset'
  } catch (error) {
    failure.value = problemFromError(error)
  } finally {
    isSubmitting.value = false
  }
}

async function submitReset(): Promise<void> {
  failure.value = null
  if (securityAnswer.value.trim() === '' || newPassword.value === '') {
    failure.value = { code: 'auth_reset_fields_required' }
    return
  }
  if (newPassword.value !== passwordConfirmation.value) {
    failure.value = { code: 'auth_password_mismatch' }
    return
  }

  isSubmitting.value = true
  try {
    await resetPassword(displayName.value.trim(), securityAnswer.value, newPassword.value)
    await router.push('/history')
  } catch (error) {
    failure.value = problemFromError(error)
  } finally {
    isSubmitting.value = false
  }
}
</script>


<template>
  <section class="page">
    <h1 class="page-title">{{ t('reset_heading') }}</h1>
    <p class="page-intro">{{ t('reset_intro') }}</p>

    <form v-if="step === 'lookup'" class="reset-card" @submit.prevent="lookupQuestion">
      <div class="reset-field">
        <label class="reset-label" for="reset-display-name">
          {{ t('reset_display_name_label') }}
        </label>
        <input
          id="reset-display-name"
          v-model="displayName"
          class="reset-input"
          type="text"
          autocomplete="username"
          :placeholder="t('login_display_name_placeholder')"
        />
      </div>

      <button type="submit" class="primary-button" :disabled="isSubmitting">
        <FontAwesomeIcon
          :icon="isSubmitting ? faSpinner : faArrowRight"
          :spin="isSubmitting"
          aria-hidden="true"
        />
        <span>{{ isSubmitting ? t('reset_lookuping') : t('reset_lookup_button') }}</span>
      </button>

      <p v-if="failure" class="status-error" role="alert">
        {{ tCode(failure.code, failure.params) }}
      </p>
    </form>

    <form v-else class="reset-card" @submit.prevent="submitReset">
      <p class="reset-question">{{ t('reset_question_intro') }}</p>
      <p class="reset-question-text">{{ securityQuestion }}</p>

      <div class="reset-field">
        <label class="reset-label" for="reset-answer">
          {{ t('reset_answer_label') }}
        </label>
        <input
          id="reset-answer"
          v-model="securityAnswer"
          class="reset-input"
          type="text"
          autocomplete="off"
          :placeholder="t('reset_answer_placeholder')"
        />
      </div>

      <div class="reset-field">
        <label class="reset-label" for="reset-new-password">
          {{ t('reset_new_password_label') }}
        </label>
        <input
          id="reset-new-password"
          v-model="newPassword"
          class="reset-input"
          type="password"
          autocomplete="new-password"
          aria-describedby="reset-password-hint"
        />
        <p id="reset-password-hint" class="reset-hint">
          {{ t('login_password_policy_hint') }}
        </p>
      </div>

      <div class="reset-field">
        <label class="reset-label" for="reset-password-confirmation">
          {{ t('reset_new_password_confirm_label') }}
        </label>
        <input
          id="reset-password-confirmation"
          v-model="passwordConfirmation"
          class="reset-input"
          type="password"
          autocomplete="new-password"
        />
      </div>

      <button type="submit" class="primary-button" :disabled="isSubmitting">
        <FontAwesomeIcon
          :icon="isSubmitting ? faSpinner : faArrowRight"
          :spin="isSubmitting"
          aria-hidden="true"
        />
        <span>{{ isSubmitting ? t('reset_submitting') : t('reset_submit_button') }}</span>
      </button>

      <div v-if="failure" class="status-error" role="alert">
        <p class="failure-text">{{ tCode(failure.code, failure.params) }}</p>
        <ul v-if="passwordRules.length > 0" class="rule-list">
          <li v-for="rule in passwordRules" :key="rule">
            {{ tCode(rule, failure.params) }}
          </li>
        </ul>
      </div>
    </form>

    <p class="reset-links">
      <RouterLink class="link-button" :to="{ name: 'login' }">
        <FontAwesomeIcon :icon="faArrowLeft" aria-hidden="true" />
        <span>{{ t('reset_back_to_login') }}</span>
      </RouterLink>
    </p>
  </section>
</template>


<style scoped>
.reset-card {
  margin-top: 1rem;
  padding: 1.5rem;
  background-color: var(--color-surface);
  border-radius: var(--radius-small);
}

.reset-field {
  margin-bottom: 1rem;
}

.reset-label {
  display: block;
  margin-bottom: 0.35rem;
  font-weight: 600;
}

.reset-input {
  width: 100%;
  padding: 0.65rem 0.8rem;
  font-family: inherit;
  font-size: 1rem;
  color: var(--color-text);
  background-color: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-small);
}

.reset-input:focus {
  border-color: var(--color-primary);
  outline: 2px solid var(--color-primary);
  outline-offset: 1px;
}

.reset-hint {
  margin: 0.4rem 0 0;
  font-size: 0.9rem;
  color: var(--color-text-muted);
}

.reset-question {
  margin: 0;
  color: var(--color-text-muted);
}

.reset-question-text {
  margin: 0.35rem 0 1rem;
  font-size: 1.1rem;
  font-weight: 600;
}

.failure-text {
  margin: 0;
}

.rule-list {
  margin: 0.5rem 0 0;
  padding-left: 1.2rem;
}

.reset-links {
  margin-top: 1.5rem;
}
</style>
