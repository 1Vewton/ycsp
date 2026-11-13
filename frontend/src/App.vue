<script setup lang="ts">
import { onMounted, ref, watchEffect } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'

import { faCubes } from '@fortawesome/free-solid-svg-icons/faCubes'
import { faFlask } from '@fortawesome/free-solid-svg-icons/faFlask'
import { faHouse } from '@fortawesome/free-solid-svg-icons/faHouse'
import { faListUl } from '@fortawesome/free-solid-svg-icons/faListUl'
import { faRightFromBracket } from '@fortawesome/free-solid-svg-icons/faRightFromBracket'
import { faRightToBracket } from '@fortawesome/free-solid-svg-icons/faRightToBracket'
import { faUser } from '@fortawesome/free-solid-svg-icons/faUser'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import { problemFromError, type ApiProblem } from './api/client'
import { AVAILABLE_LOCALES, currentLocale, LOCALE_NAMES, setLocale, t, tCode } from './locales'
import { isRestoringSession, restoreSession, signOut, signedInUser } from './session'

const route = useRoute()
const isSigningOut = ref(false)
const signOutFailure = ref<ApiProblem | null>(null)

// 浏览器标签页标题跟随界面语言。
watchEffect(() => {
  document.title = t('page_title')
})

onMounted(() => {
  void restoreSession()
})

async function leaveAccount(): Promise<void> {
  isSigningOut.value = true
  signOutFailure.value = null
  try {
    await signOut()
  } catch (error) {
    signOutFailure.value = problemFromError(error)
  } finally {
    isSigningOut.value = false
  }
}
</script>

<template>
  <div class="app-shell">
    <header class="app-header">
      <div class="app-header-inner">
        <RouterLink class="app-brand" to="/">
          <FontAwesomeIcon :icon="faCubes" aria-hidden="true" />
          <span>{{ t('page_title') }}</span>
        </RouterLink>

        <nav class="app-nav" :aria-label="t('main_navigation_label')">
          <RouterLink class="nav-link" to="/">
            <FontAwesomeIcon :icon="faHouse" aria-hidden="true" />
            <span>{{ t('nav_home') }}</span>
          </RouterLink>
          <RouterLink class="nav-link" to="/predict">
            <FontAwesomeIcon :icon="faFlask" aria-hidden="true" />
            <span>{{ t('nav_predict') }}</span>
          </RouterLink>
          <RouterLink class="nav-link" to="/history">
            <FontAwesomeIcon :icon="faListUl" aria-hidden="true" />
            <span>{{ t('nav_history') }}</span>
          </RouterLink>

          <template v-if="!isRestoringSession">
            <RouterLink
              v-if="!signedInUser"
              class="nav-link"
              :to="{ name: 'login', query: { redirect: route.fullPath } }"
            >
              <FontAwesomeIcon :icon="faRightToBracket" aria-hidden="true" />
              <span>{{ t('nav_login') }}</span>
            </RouterLink>
            <template v-else>
              <RouterLink
                class="nav-link"
                :to="{ name: 'account' }"
                :aria-label="t('nav_account_label', { name: signedInUser.display_name })"
                :title="signedInUser.display_name"
              >
                <FontAwesomeIcon :icon="faUser" aria-hidden="true" />
                <span class="nav-account-name">{{ signedInUser.display_name }}</span>
              </RouterLink>
              <button
                type="button"
                class="nav-link nav-signout"
                :disabled="isSigningOut"
                @click="leaveAccount"
              >
                <FontAwesomeIcon :icon="faRightFromBracket" aria-hidden="true" />
                <span>{{ isSigningOut ? t('nav_signing_out') : t('nav_logout') }}</span>
              </button>
            </template>
          </template>
        </nav>

        <div class="locale-switcher" role="group" :aria-label="t('language_switcher_label')">
          <button
            v-for="locale in AVAILABLE_LOCALES"
            :key="locale"
            type="button"
            class="locale-button"
            :class="{ 'locale-button-active': currentLocale === locale }"
            :aria-pressed="currentLocale === locale"
            @click="setLocale(locale)"
          >
            {{ LOCALE_NAMES[locale] }}
          </button>
        </div>
      </div>
      <p v-if="signOutFailure" class="header-alert" role="alert">
        {{ tCode(signOutFailure.code, signOutFailure.params) }}
      </p>
    </header>

    <main class="app-main">
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.app-header {
  background-color: var(--color-surface);
  border-bottom: 1px solid var(--color-divider);
}

.app-header-inner {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem 1rem;
  align-items: center;
  justify-content: space-between;
  max-width: var(--content-width);
  margin: 0 auto;
  padding: 0.9rem 1.5rem;
}

.app-brand {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: var(--color-primary);
  text-decoration: none;
}

.app-nav {
  display: flex;
  gap: 0.25rem;
}

.nav-link {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.3rem 0.6rem;
  color: var(--color-text-muted);
  text-decoration: none;
  border-radius: var(--radius-small);
}

.nav-link:hover,
.nav-link.router-link-active {
  color: var(--color-primary);
  background-color: var(--color-primary-soft);
}

.nav-link.router-link-active {
  font-weight: 600;
}

.locale-switcher {
  display: flex;
  gap: 0.4rem;
}

.locale-button {
  padding: 0.3rem 0.75rem;
  font-family: inherit;
  font-size: 0.875rem;
  color: var(--color-primary);
  background-color: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-pill);
  cursor: pointer;
}

.locale-button:hover {
  background-color: var(--color-primary-soft);
}

.locale-button-active {
  color: #ffffff;
  background-color: var(--color-primary);
  border-color: var(--color-primary);
}

.nav-account-name {
  max-width: 8rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.nav-signout {
  font-family: inherit;
  font-size: inherit;
  background-color: transparent;
  border: none;
  cursor: pointer;
}

.nav-signout:disabled {
  cursor: progress;
  opacity: 0.6;
}

.header-alert {
  max-width: var(--content-width);
  margin: 0 auto;
  padding: 0 1.5rem 0.75rem;
  color: var(--color-error-text);
}
</style>
