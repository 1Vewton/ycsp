import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { AccountUser } from '../src/api/auth'

const authMocks = vi.hoisted(() => ({
  fetchCurrentUser: vi.fn(),
  registerAccount: vi.fn(),
  signIn: vi.fn(),
  signOut: vi.fn(),
  changePassword: vi.fn(),
  updateSecurityQuestion: vi.fn(),
  fetchPasswordResetQuestion: vi.fn(),
  resetPassword: vi.fn(),
}))

vi.mock('../src/api/auth', () => authMocks)

import { ApiError } from '../src/api/client'
import { t, tCode } from '../src/locales'
import ResetPasswordView from '../src/views/ResetPasswordView.vue'

const account: AccountUser = {
  id: 3,
  display_name: '小明',
  created_at: '2026-11-02T00:00:00Z',
}

beforeEach(() => {
  for (const mockFunction of Object.values(authMocks)) {
    mockFunction.mockReset()
  }
})

async function mountReset() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div />' } },
      { path: '/login', name: 'login', component: { template: '<div />' } },
      { path: '/history', name: 'history', component: { template: '<div />' } },
      { path: '/reset-password', name: 'reset-password', component: ResetPasswordView },
    ],
  })
  await router.push('/reset-password')
  await router.isReady()

  const wrapper = mount(ResetPasswordView, { global: { plugins: [router] } })
  await flushPromises()
  return { wrapper, router }
}

describe('ResetPasswordView 安全问题重置', () => {
  it('按昵称取到问题后显示问题与答案输入框', async () => {
    authMocks.fetchPasswordResetQuestion.mockResolvedValue('最喜欢的元素？')
    const { wrapper } = await mountReset()

    await wrapper.find('#reset-display-name').setValue('小明')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(authMocks.fetchPasswordResetQuestion).toHaveBeenCalledWith('小明')
    expect(wrapper.text()).toContain('最喜欢的元素？')
    expect(wrapper.find('#reset-answer').exists()).toBe(true)
    expect(wrapper.find('#reset-new-password').exists()).toBe(true)
  })

  it('答案与新密码正确时调用重置接口并跳到记录页', async () => {
    authMocks.fetchPasswordResetQuestion.mockResolvedValue('最喜欢的元素？')
    authMocks.resetPassword.mockResolvedValue(account)
    const { wrapper, router } = await mountReset()

    await wrapper.find('#reset-display-name').setValue('小明')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    await wrapper.find('#reset-answer').setValue('钠')
    await wrapper.find('#reset-new-password').setValue('new-pass')
    await wrapper.find('#reset-password-confirmation').setValue('new-pass')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(authMocks.resetPassword).toHaveBeenCalledWith('小明', '钠', 'new-pass')
    expect(router.currentRoute.value.fullPath).toBe('/history')
  })

  it('账号没有安全问题或不存在时显示可翻译的提示', async () => {
    authMocks.fetchPasswordResetQuestion.mockRejectedValue(
      new ApiError({ code: 'password_reset_unavailable' }),
    )
    const { wrapper } = await mountReset()

    await wrapper.find('#reset-display-name').setValue('小明')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain(tCode('password_reset_unavailable'))
    expect(wrapper.find('#reset-answer').exists()).toBe(false)
  })

  it('两次新密码不一致时本地拦下，不发请求', async () => {
    authMocks.fetchPasswordResetQuestion.mockResolvedValue('最喜欢的元素？')
    const { wrapper } = await mountReset()

    await wrapper.find('#reset-display-name').setValue('小明')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    await wrapper.find('#reset-answer').setValue('钠')
    await wrapper.find('#reset-new-password').setValue('new-pass')
    await wrapper.find('#reset-password-confirmation').setValue('different')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(authMocks.resetPassword).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain(t('auth_password_mismatch'))
  })
})
