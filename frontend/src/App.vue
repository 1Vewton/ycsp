<script setup lang="ts">
import { watchEffect } from 'vue'
import { RouterLink, RouterView } from 'vue-router'

import { faCubes } from '@fortawesome/free-solid-svg-icons/faCubes'
import { faFlask } from '@fortawesome/free-solid-svg-icons/faFlask'
import { faHouse } from '@fortawesome/free-solid-svg-icons/faHouse'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import { AVAILABLE_LOCALES, currentLocale, LOCALE_NAMES, setLocale, t } from './locales'

// 浏览器标签页标题跟随界面语言。
watchEffect(() => {
  document.title = t('page_title')
})
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
</style>
