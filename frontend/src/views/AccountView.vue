<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'

import { faCircleCheck } from '@fortawesome/free-solid-svg-icons/faCircleCheck'
import { faKey } from '@fortawesome/free-solid-svg-icons/faKey'
import { faShieldHalved } from '@fortawesome/free-solid-svg-icons/faShieldHalved'
import { faSpinner } from '@fortawesome/free-solid-svg-icons/faSpinner'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import { problemFromError, type ApiProblem } from '../api/client'
import { t, tCode } from '../locales'
import {
  changePassword,
  isRestoringSession,
  restoreSession,
  signedInUser,
  updateSecurityQuestion,
} from '../session'

const passwordCurrent = ref('')
const passwordNew = ref('')
const passwordConfirmation = ref('')
const passwordSubmitting = ref(false)
const passwordFailure = ref<ApiProblem | null>(null)
const passwordChanged = ref(false)

const securityCurrent = ref('')
const securityQuestion = ref('')
const securityAnswer = ref('')
const securitySubmitting = ref(false)
const securityFailure = ref<ApiProblem | null>(null)
const securityUpdated = ref(false)

const currentSecurityQuestion = computed(() => signedInUser.value?.security_question ?? '')

// 密码规则由后端判定；这里只把返回的规则逐条翻出来，不在前端复制一份阈值。
const passwordRules = computed(() => {
  const rules = passwordFailure.value?.params?.rules
  if (!Array.isArray(rules)) {
    return []
  }
  return rules.filter((rule): rule is string => typeof rule === 'string')
})

const canSubmitPassword = computed(
  () =>
    !passwordSubmitting.value &&
    passwordCurrent.value !== '' &&
    passwordNew.value !== '' &&
    passwordConfirmation.value !== '',
)

const canSubmitSecurity = computed(
  () =>
    !securitySubmitting.value &&
    securityCurrent.value !== '' &&
    securityQuestion.value.trim() !== '' &&
    securityAnswer.value.trim() !== '',
)

watch([passwordCurrent, passwordNew, passwordConfirmation], () => {
  passwordFailure.value = null
  passwordChanged.value = false
})

watch([securityCurrent, securityQuestion, securityAnswer], () => {
  securityFailure.value = null
  securityUpdated.value = false
})

onMounted(() => {
  void restoreSession()
})

async function submitPassword(): Promise<void> {
  passwordFailure.value = null
  passwordChanged.value = false
  if (passwordNew.value !== passwordConfirmation.value) {
    passwordFailure.value = { code: 'auth_password_mismatch' }
    return
  }

  passwordSubmitting.value = true
  try {
    await changePassword(passwordCurrent.value, passwordNew.value)
    passwordCurrent.value = ''
    passwordNew.value = ''
    passwordConfirmation.value = ''
    await nextTick()
    passwordChanged.value = true
  } catch (error) {
    passwordFailure.value = problemFromError(error)
  } finally {
    passwordSubmitting.value = false
  }
}

async function submitSecurityQuestion(): Promise<void> {
  securityFailure.value = null
  securityUpdated.value = false
  if (securityQuestion.value.trim() === '' || securityAnswer.value.trim() === '') {
    securityFailure.value = { code: 'auth_security_incomplete' }
    return
  }

  securitySubmitting.value = true
  try {
    await updateSecurityQuestion(
      securityCurrent.value,
      securityQuestion.value,
      securityAnswer.value,
    )
    securityCurrent.value = ''
    securityAnswer.value = ''
    await nextTick()
    securityUpdated.value = true
  } catch (error) {
    securityFailure.value = problemFromError(error)
  } finally {
    securitySubmitting.value = false
  }
}
</script>


<template>
  <section class="page">
    <h1 class="page-title">{{ t('account_heading') }}</h1>
    <p class="page-intro">{{ t('account_intro') }}</p>

    <p v-if="isRestoringSession" class="status-hint" role="status">
      {{ t('auth_checking_session') }}
    </p>

    <section v-else-if="!signedInUser" class="settings-card">
      <p class="settings-text">{{ t('account_signed_out_hint') }}</p>
      <RouterLink
        class="primary-button"
        :to="{ name: 'login', query: { redirect: '/account' } }"
      >
        {{ t('account_sign_in_link') }}
      </RouterLink>
    </section>

    <template v-else>
      <section class="settings-card">
        <h2 class="settings-title">{{ t('auth_signed_in_title') }}</h2>
        <p class="settings-text">
          {{ t('auth_signed_in_as', { name: signedInUser.display_name }) }}
        </p>
        <p class="settings-text">
          {{
            currentSecurityQuestion
              ? t('account_security_current_question', { question: currentSecurityQuestion })
              : t('account_security_not_set')
          }}
        </p>
      </section>

      <section class="settings-card">
        <h2 class="settings-title">
          <FontAwesomeIcon :icon="faKey" aria-hidden="true" />
          <span>{{ t('account_password_title') }}</span>
        </h2>
        <p class="settings-text">{{ t('account_password_intro') }}</p>

        <form @submit.prevent="submitPassword">
          <div class="settings-field">
            <label class="settings-label" for="account-current-password">
              {{ t('account_current_password_label') }}
            </label>
            <input
              id="account-current-password"
              v-model="passwordCurrent"
              class="settings-input"
              type="password"
              autocomplete="current-password"
            />
          </div>

          <div class="settings-field">
            <label class="settings-label" for="account-new-password">
              {{ t('account_new_password_label') }}
            </label>
            <input
              id="account-new-password"
              v-model="passwordNew"
              class="settings-input"
              type="password"
              autocomplete="new-password"
              aria-describedby="account-password-hint"
            />
            <p id="account-password-hint" class="settings-hint">
              {{ t('login_password_policy_hint') }}
            </p>
          </div>

          <div class="settings-field">
            <label class="settings-label" for="account-new-password-confirmation">
              {{ t('account_new_password_confirm_label') }}
            </label>
            <input
              id="account-new-password-confirmation"
              v-model="passwordConfirmation"
              class="settings-input"
              type="password"
              autocomplete="new-password"
            />
          </div>

          <button type="submit" class="primary-button" :disabled="!canSubmitPassword">
            <FontAwesomeIcon
              :icon="passwordSubmitting ? faSpinner : faKey"
              :spin="passwordSubmitting"
              aria-hidden="true"
            />
            <span>
              {{
                passwordSubmitting
                  ? t('account_password_submitting')
                  : t('account_change_password_button')
              }}
            </span>
          </button>

          <p v-if="passwordChanged" class="status-success" role="status">
            <FontAwesomeIcon :icon="faCircleCheck" aria-hidden="true" />
            <span>{{ t('account_password_changed') }}</span>
          </p>
          <div v-if="passwordFailure" class="status-error" role="alert">
            <p class="failure-text">{{ tCode(passwordFailure.code, passwordFailure.params) }}</p>
            <ul v-if="passwordRules.length > 0" class="rule-list">
              <li v-for="rule in passwordRules" :key="rule">
                {{ tCode(rule, passwordFailure.params) }}
              </li>
            </ul>
          </div>
        </form>
      </section>


      <section class="settings-card">
        <h2 class="settings-title">
          <FontAwesomeIcon :icon="faShieldHalved" aria-hidden="true" />
          <span>{{ t('account_security_title') }}</span>
        </h2>
        <p class="settings-text">{{ t('account_security_intro') }}</p>

        <form @submit.prevent="submitSecurityQuestion">
          <div class="settings-field">
            <label class="settings-label" for="account-security-current-password">
              {{ t('account_current_password_label') }}
            </label>
            <input
              id="account-security-current-password"
              v-model="securityCurrent"
              class="settings-input"
              type="password"
              autocomplete="current-password"
            />
          </div>

          <div class="settings-field">
            <label class="settings-label" for="account-security-question">
              {{ t('account_security_question_label') }}
            </label>
            <input
              id="account-security-question"
              v-model="securityQuestion"
              class="settings-input"
              type="text"
              :placeholder="t('account_security_question_placeholder')"
            />
          </div>

          <div class="settings-field">
            <label class="settings-label" for="account-security-answer">
              {{ t('account_security_answer_label') }}
            </label>
            <input
              id="account-security-answer"
              v-model="securityAnswer"
              class="settings-input"
              type="text"
              autocomplete="off"
              :placeholder="t('account_security_answer_placeholder')"
            />
            <p class="settings-hint">{{ t('login_security_hint') }}</p>
          </div>

          <button type="submit" class="primary-button" :disabled="!canSubmitSecurity">
            <FontAwesomeIcon
              :icon="securitySubmitting ? faSpinner : faShieldHalved"
              :spin="securitySubmitting"
              aria-hidden="true"
            />
            <span>
              {{
                securitySubmitting
                  ? t('account_security_submitting')
                  : t('account_security_setup_button')
              }}
            </span>
          </button>

          <p v-if="securityUpdated" class="status-success" role="status">
            <FontAwesomeIcon :icon="faCircleCheck" aria-hidden="true" />
            <span>{{ t('account_security_updated') }}</span>
          </p>
          <p v-if="securityFailure" class="status-error" role="alert">
            {{ tCode(securityFailure.code, securityFailure.params) }}
          </p>
        </form>
      </section>

      <p class="settings-links">
        <RouterLink class="link-button" to="/predict">{{ t('nav_predict') }}</RouterLink>
        <RouterLink class="link-button" to="/history">{{ t('nav_history') }}</RouterLink>
      </p>
    </template>
  </section>
</template>

<style scoped>
.settings-card {
  margin-top: 1rem;
  padding: 1.5rem;
  background-color: var(--color-surface);
  border-radius: var(--radius-small);
}

.settings-title {
  display: flex;
  gap: 0.45rem;
  align-items: center;
  margin: 0 0 0.5rem;
  font-size: 1.15rem;
}

.settings-text {
  margin: 0 0 1rem;
  line-height: 1.7;
  color: var(--color-text-muted);
}

.settings-field {
  margin-bottom: 1rem;
}

.settings-label {
  display: block;
  margin-bottom: 0.35rem;
  font-weight: 600;
}

.settings-input {
  width: 100%;
  padding: 0.65rem 0.8rem;
  font-family: inherit;
  font-size: 1rem;
  color: var(--color-text);
  background-color: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-small);
}

.settings-input:focus {
  border-color: var(--color-primary);
  outline: 2px solid var(--color-primary);
  outline-offset: 1px;
}

.settings-hint {
  margin: 0.4rem 0 0;
  font-size: 0.9rem;
  color: var(--color-text-muted);
}

.failure-text {
  margin: 0;
}

.rule-list {
  margin: 0.5rem 0 0;
  padding-left: 1.2rem;
}

.status-success {
  display: flex;
  gap: 0.4rem;
  align-items: center;
}

.settings-links {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  margin-top: 1.5rem;
}
</style>
