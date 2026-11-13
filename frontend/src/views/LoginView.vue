<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { faArrowRight } from '@fortawesome/free-solid-svg-icons/faArrowRight'
import { faRightFromBracket } from '@fortawesome/free-solid-svg-icons/faRightFromBracket'
import { faRightToBracket } from '@fortawesome/free-solid-svg-icons/faRightToBracket'
import { faSpinner } from '@fortawesome/free-solid-svg-icons/faSpinner'
import { faUserPlus } from '@fortawesome/free-solid-svg-icons/faUserPlus'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import { problemFromError, type ApiProblem } from '../api/client'
import { t, tCode } from '../locales'
import {
  isRestoringSession,
  registerAccount,
  restoreSession,
  signIn,
  signOut,
  signedInUser,
} from '../session'

// 登录与注册共用一张表单：切换时只换按钮文案与“再输一次密码”一栏。
type AuthMode = 'signIn' | 'register'

const route = useRoute()
const router = useRouter()

const mode = ref<AuthMode>('signIn')
const displayName = ref('')
const password = ref('')
const passwordConfirmation = ref('')
const securityQuestion = ref('')
const securityAnswer = ref('')
const isSubmitting = ref(false)
const isSigningOut = ref(false)
const failure = ref<ApiProblem | null>(null)

// 密码规则由后端判定；这里只把返回的规则逐条翻出来，不在前端复制一份阈值。
const passwordRules = computed(() => {
  const rules = failure.value?.params?.rules
  if (!Array.isArray(rules)) {
    return []
  }
  return rules.filter((rule): rule is string => typeof rule === 'string')
})

const canSubmit = computed(() => {
  if (isSubmitting.value || displayName.value.trim() === '' || password.value === '') {
    return false
  }
  return mode.value !== 'register' || passwordConfirmation.value !== ''
})

const submitIcon = computed(() =>
  mode.value === 'register' ? faUserPlus : faRightToBracket,
)

// 改过输入就把上一次的失败提示收走，别让旧错误留在页面上。
watch([mode, displayName, password, passwordConfirmation, securityQuestion, securityAnswer], () => {
  failure.value = null
})

onMounted(() => {
  void restoreSession()
})

function switchMode(nextMode: AuthMode): void {
  mode.value = nextMode
  password.value = ''
  passwordConfirmation.value = ''
  securityQuestion.value = ''
  securityAnswer.value = ''
  failure.value = null
}

async function submitCredentials(): Promise<void> {
  failure.value = null
  const name = displayName.value.trim()
  if (name === '' || password.value === '') {
    failure.value = { code: 'auth_fields_required' }
    return
  }
  if (mode.value === 'register' && password.value !== passwordConfirmation.value) {
    failure.value = { code: 'auth_password_mismatch' }
    return
  }
  if (
    mode.value === 'register' &&
    (securityQuestion.value.trim() === '') !== (securityAnswer.value.trim() === '')
  ) {
    failure.value = { code: 'auth_security_incomplete' }
    return
  }

  isSubmitting.value = true
  try {
    if (mode.value === 'register') {
      await registerAccount(
        name,
        password.value,
        securityQuestion.value,
        securityAnswer.value,
      )
    } else {
      await signIn(name, password.value)
    }
    await router.push(targetPathAfterSuccess())
  } catch (error) {
    failure.value = problemFromError(error)
  } finally {
    isSubmitting.value = false
  }
}

async function leaveAccount(): Promise<void> {
  isSigningOut.value = true
  failure.value = null
  try {
    await signOut()
  } catch (error) {
    failure.value = problemFromError(error)
  } finally {
    isSigningOut.value = false
  }
}

// 登录后回到进入登录页之前的位置；只接受站内路径，避免被 query 带到外部站点。
// redirect 指回登录页本身时改为去预测页，免得登录后停在原地。
function targetPathAfterSuccess(): string {
  const requested = route.query.redirect
  if (typeof requested !== 'string') {
    return '/predict'
  }
  const path = requested.split(/[?#]/, 1)[0]
  const isInternalPath =
    path.startsWith('/') && !path.startsWith('//') && path !== '/login'
  return isInternalPath ? requested : '/predict'
}
</script>

<template>
  <section class="page">
    <h1 class="page-title">{{ t('login_heading') }}</h1>
    <p class="page-intro">{{ t('login_intro') }}</p>

    <p v-if="isRestoringSession" class="status-hint" role="status">
      {{ t('auth_checking_session') }}
    </p>

    <section v-else-if="signedInUser" class="auth-card">
      <h2 class="auth-title">{{ t('auth_signed_in_title') }}</h2>
      <p class="auth-text">
        {{ t('auth_signed_in_as', { name: signedInUser.display_name }) }}
      </p>
      <div class="auth-actions">
        <RouterLink class="primary-button" to="/predict">
          <FontAwesomeIcon :icon="faArrowRight" aria-hidden="true" />
          <span>{{ t('auth_continue_button') }}</span>
        </RouterLink>
        <RouterLink class="secondary-button" to="/account">
          <span>{{ t('account_heading') }}</span>
        </RouterLink>
        <button
          type="button"
          class="secondary-button"
          :disabled="isSigningOut"
          @click="leaveAccount"
        >
          <FontAwesomeIcon :icon="faRightFromBracket" aria-hidden="true" />
          <span>{{ isSigningOut ? t('auth_sign_out_busy') : t('auth_sign_out_button') }}</span>
        </button>
      </div>
      <p v-if="failure" class="status-error" role="alert">
        {{ tCode(failure.code, failure.params) }}
      </p>
    </section>

    <template v-else>
      <div class="mode-switch" role="group" :aria-label="t('login_mode_label')">
        <button
          type="button"
          class="mode-button"
          :class="{ 'mode-button-active': mode === 'signIn' }"
          :aria-pressed="mode === 'signIn'"
          @click="switchMode('signIn')"
        >
          {{ t('login_mode_sign_in') }}
        </button>
        <button
          type="button"
          class="mode-button"
          :class="{ 'mode-button-active': mode === 'register' }"
          :aria-pressed="mode === 'register'"
          @click="switchMode('register')"
        >
          {{ t('login_mode_register') }}
        </button>
      </div>

      <form class="auth-card" @submit.prevent="submitCredentials">
        <div class="auth-field">
          <label class="auth-label" for="login-display-name">
            {{ t('login_display_name_label') }}
          </label>
          <input
            id="login-display-name"
            v-model="displayName"
            class="auth-input"
            type="text"
            autocomplete="username"
            aria-describedby="login-display-name-hint"
            :placeholder="t('login_display_name_placeholder')"
          />
          <p id="login-display-name-hint" class="auth-hint">
            {{ t('login_display_name_hint') }}
          </p>
        </div>

        <div class="auth-field">
          <label class="auth-label" for="login-password">
            {{ t('login_password_label') }}
          </label>
          <input
            id="login-password"
            v-model="password"
            class="auth-input"
            type="password"
            :autocomplete="mode === 'register' ? 'new-password' : 'current-password'"
            :aria-describedby="mode === 'register' ? 'login-password-hint' : undefined"
            :placeholder="t('login_password_placeholder')"
          />
          <p v-if="mode === 'register'" id="login-password-hint" class="auth-hint">
            {{ t('login_password_policy_hint') }}
          </p>
          <p v-else class="auth-hint">
            <RouterLink class="link-button" to="/reset-password">
              {{ t('login_forgot_password_link') }}
            </RouterLink>
          </p>
        </div>

        <div v-if="mode === 'register'" class="auth-field">
          <label class="auth-label" for="login-password-confirmation">
            {{ t('login_password_confirm_label') }}
          </label>
          <input
            id="login-password-confirmation"
            v-model="passwordConfirmation"
            class="auth-input"
            type="password"
            autocomplete="new-password"
            :placeholder="t('login_password_confirm_placeholder')"
          />
        </div>

        <div v-if="mode === 'register'" class="auth-field">
          <label class="auth-label" for="login-security-question">
            {{ t('login_security_question_label') }}
          </label>
          <input
            id="login-security-question"
            v-model="securityQuestion"
            class="auth-input"
            type="text"
            :placeholder="t('login_security_question_placeholder')"
          />
        </div>

        <div v-if="mode === 'register'" class="auth-field">
          <label class="auth-label" for="login-security-answer">
            {{ t('login_security_answer_label') }}
          </label>
          <input
            id="login-security-answer"
            v-model="securityAnswer"
            class="auth-input"
            type="text"
            autocomplete="off"
            :placeholder="t('login_security_answer_placeholder')"
          />
          <p class="auth-hint">{{ t('login_security_hint') }}</p>
        </div>

        <button type="submit" class="primary-button auth-submit" :disabled="!canSubmit">
          <FontAwesomeIcon
            :icon="isSubmitting ? faSpinner : submitIcon"
            :spin="isSubmitting"
            aria-hidden="true"
          />
          <span>
            {{ isSubmitting ? t('login_submitting') : t(mode === 'register' ? 'login_submit_register' : 'login_submit_sign_in') }}
          </span>
        </button>

        <div v-if="failure" class="status-error" role="alert">
          <p class="failure-text">{{ tCode(failure.code, failure.params) }}</p>
          <ul v-if="passwordRules.length > 0" class="rule-list">
            <li v-for="rule in passwordRules" :key="rule">
              {{ tCode(rule, failure.params) }}
            </li>
          </ul>
        </div>

        <button
          type="button"
          class="switch-link"
          @click="switchMode(mode === 'signIn' ? 'register' : 'signIn')"
        >
          {{ mode === 'signIn' ? t('login_switch_to_register') : t('login_switch_to_sign_in') }}
        </button>
      </form>

      <p class="anonymous-hint">
        {{ t('login_anonymous_hint') }}
        <RouterLink class="link-button" to="/predict">
          {{ t('login_anonymous_link') }}
        </RouterLink>
      </p>
    </template>
  </section>
</template>

<style scoped>
.mode-switch {
  display: inline-flex;
  gap: 0.35rem;
  margin-bottom: 1rem;
  padding: 0.25rem;
  background-color: var(--color-surface);
  border: 1px solid var(--color-divider);
  border-radius: var(--radius-pill);
}

.mode-button {
  padding: 0.4rem 1rem;
  font-family: inherit;
  font-size: 0.95rem;
  color: var(--color-primary);
  background-color: transparent;
  border: none;
  border-radius: var(--radius-pill);
  cursor: pointer;
}

.mode-button:hover {
  background-color: var(--color-primary-soft);
}

.mode-button-active {
  color: #ffffff;
  background-color: var(--color-primary);
}

.mode-button-active:hover {
  background-color: var(--color-primary-dark);
}

.auth-card {
  margin-top: 1rem;
  padding: 1.5rem;
  background-color: var(--color-surface);
  border-radius: var(--radius-small);
}

.auth-title {
  margin: 0 0 0.5rem;
  font-size: 1.2rem;
}

.auth-text {
  margin: 0 0 1rem;
  color: var(--color-text-muted);
}

.auth-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
}

.auth-field {
  margin-bottom: 1rem;
}

.auth-label {
  display: block;
  margin-bottom: 0.35rem;
  font-weight: 600;
}

.auth-input {
  width: 100%;
  padding: 0.65rem 0.8rem;
  font-family: inherit;
  font-size: 1rem;
  color: var(--color-text);
  background-color: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-small);
}

.auth-input:focus {
  border-color: var(--color-primary);
  outline: 2px solid var(--color-primary);
  outline-offset: 1px;
}

.auth-hint {
  margin: 0.4rem 0 0;
  font-size: 0.9rem;
  color: var(--color-text-muted);
}

.auth-submit {
  margin-top: 0.25rem;
}

.failure-text {
  margin: 0;
}

.rule-list {
  margin: 0.5rem 0 0;
  padding-left: 1.2rem;
}

.switch-link {
  display: block;
  margin-top: 1rem;
  padding: 0;
  font-family: inherit;
  font-size: 0.95rem;
  color: var(--color-primary);
  text-decoration: underline;
  background: none;
  border: none;
  cursor: pointer;
}

.anonymous-hint {
  margin-top: 1.5rem;
  color: var(--color-text-muted);
}
</style>

