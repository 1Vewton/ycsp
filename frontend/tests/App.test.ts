import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { AccountUser } from '../src/api/auth'

const authMocks = vi.hoisted(() => ({
  fetchCurrentUser: vi.fn(),
  registerAccount: vi.fn(),
  signIn: vi.fn(),
  signOut: vi.fn(),
}))

vi.mock('../src/api/auth', () => authMocks)

import App from '../src/App.vue'
import { ApiError } from '../src/api/client'
import { t, tCode } from '../src/locales'
import * as session from '../src/session'

const account: AccountUser = {
  id: 3,
  display_name: '小明',
  created_at: '2026-11-02T00:00:00Z',
}

beforeEach(async () => {
  for (const mockFunction of Object.values(authMocks)) {
    mockFunction.mockReset()
  }
  authMocks.fetchCurrentUser.mockResolvedValue(null)
  authMocks.signOut.mockResolvedValue(undefined)

  await session.signOut()
  authMocks.signOut.mockClear()
})

async function mountApp(path = '/predict') {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'home', component: { template: '<div>home</div>' } },
      { path: '/predict', name: 'predict', component: { template: '<div>predict</div>' } },
      { path: '/login', name: 'login', component: { template: '<div>login</div>' } },
    ],
  })
  await router.push(path)
  await router.isReady()

  const wrapper = mount(App, { global: { plugins: [router] } })
  await flushPromises()
  return { wrapper, router }
}

describe('App 顶栏的登录状态', () => {
  it('未登录时显示登录入口，并把当前页面作为登录后的返回地址', async () => {
    const { wrapper, router } = await mountApp('/predict')

    const loginLink = wrapper.findAll('a').find((link) => link.text() === t('nav_login'))

    expect(loginLink).toBeDefined()
    expect(loginLink?.attributes('href')).toBe(
      router.resolve({ name: 'login', query: { redirect: '/predict' } }).fullPath,
    )
  })

  it('已登录时显示昵称与退出按钮，退出后回到登录入口', async () => {
    authMocks.signIn.mockResolvedValue(account)
    await session.signIn('小明', 'Passw0rd!')

    const { wrapper } = await mountApp()
    expect(wrapper.text()).toContain(account.display_name)

    const logoutButton = wrapper
      .findAll('button')
      .find((button) => button.text() === t('nav_logout'))
    expect(logoutButton).toBeDefined()
    await logoutButton?.trigger('click')
    await flushPromises()

    expect(authMocks.signOut).toHaveBeenCalledOnce()
    expect(wrapper.findAll('a').some((link) => link.text() === t('nav_login'))).toBe(true)
  })

  it('退出失败时保留登录状态，并在顶栏给出可翻译的原因', async () => {
    authMocks.signIn.mockResolvedValue(account)
    authMocks.signOut.mockRejectedValue(new ApiError({ code: 'session_unavailable' }))
    await session.signIn('小明', 'Passw0rd!')

    const { wrapper } = await mountApp()
    const logoutButton = wrapper
      .findAll('button')
      .find((button) => button.text() === t('nav_logout'))
    await logoutButton?.trigger('click')
    await flushPromises()

    expect(wrapper.find('[role="alert"]').text()).toContain(tCode('session_unavailable'))
    expect(wrapper.text()).toContain(account.display_name)
  })
})
