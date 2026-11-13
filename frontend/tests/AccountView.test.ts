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
import * as session from '../src/session'
import AccountView from '../src/views/AccountView.vue'

const account: AccountUser = {
  id: 3,
  display_name: '小明',
  security_question: '最喜欢的元素？',
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

async function mountAccount() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div />' } },
      { path: '/login', name: 'login', component: { template: '<div />' } },
      { path: '/predict', component: { template: '<div />' } },
      { path: '/history', component: { template: '<div />' } },
      { path: '/account', name: 'account', component: AccountView },
    ],
  })
  await router.push('/account')
  await router.isReady()

  const wrapper = mount(AccountView, { global: { plugins: [router] } })
  await flushPromises()
  return { wrapper, router }
}

async function signInAccount(): Promise<void> {
  authMocks.signIn.mockResolvedValue(account)
  await session.signIn('小明', 'Passw0rd!')
}

describe('AccountView 账号设置', () => {
  it('未登录时提示先登录', async () => {
    const { wrapper } = await mountAccount()

    expect(wrapper.text()).toContain(t('account_signed_out_hint'))
    expect(wrapper.text()).toContain(t('account_sign_in_link'))
  })

  it('修改密码：提交当前密码与新密码，成功后清空输入', async () => {
    await signInAccount()
    authMocks.changePassword.mockResolvedValue(account)
    const { wrapper } = await mountAccount()

    await wrapper.find('#account-current-password').setValue('old-pass')
    await wrapper.find('#account-new-password').setValue('new-pass')
    await wrapper.find('#account-new-password-confirmation').setValue('new-pass')
    await wrapper.findAll('form')[0].trigger('submit')
    await flushPromises()

    expect(authMocks.changePassword).toHaveBeenCalledWith('old-pass', 'new-pass')
    expect(wrapper.text()).toContain(t('account_password_changed'))
    expect((wrapper.find('#account-current-password').element as HTMLInputElement).value).toBe('')
  })

  it('两次新密码不一致时本地拦下，不发请求', async () => {
    await signInAccount()
    const { wrapper } = await mountAccount()

    await wrapper.find('#account-current-password').setValue('old-pass')
    await wrapper.find('#account-new-password').setValue('new-pass')
    await wrapper.find('#account-new-password-confirmation').setValue('different')
    await wrapper.findAll('form')[0].trigger('submit')
    await flushPromises()

    expect(authMocks.changePassword).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain(tCode('auth_password_mismatch'))
  })

  it('修改密码失败时逐条显示密码规则', async () => {
    await signInAccount()
    const params = { rules: ['password_rule_missing_digit'], min_length: 8, max_bytes: 72 }
    authMocks.changePassword.mockRejectedValue(
      new ApiError({ code: 'password_too_weak', params }),
    )
    const { wrapper } = await mountAccount()

    await wrapper.find('#account-current-password').setValue('old-pass')
    await wrapper.find('#account-new-password').setValue('abcdefgh')
    await wrapper.find('#account-new-password-confirmation').setValue('abcdefgh')
    await wrapper.findAll('form')[0].trigger('submit')
    await flushPromises()

    const alertText = wrapper.find('[role="alert"]').text()
    expect(alertText).toContain(tCode('password_rule_missing_digit', params))
  })

  it('设置安全问题：提交当前密码、问题与答案', async () => {
    await signInAccount()
    const updated = { ...account, security_question: '最喜欢的晶体？' }
    authMocks.updateSecurityQuestion.mockResolvedValue(updated)
    const { wrapper } = await mountAccount()

    await wrapper.find('#account-security-current-password').setValue('old-pass')
    await wrapper.find('#account-security-question').setValue('最喜欢的晶体？')
    await wrapper.find('#account-security-answer').setValue('NaCl')
    await wrapper.findAll('form')[1].trigger('submit')
    await flushPromises()

    expect(authMocks.updateSecurityQuestion).toHaveBeenCalledWith(
      'old-pass',
      '最喜欢的晶体？',
      'NaCl',
    )
    expect(wrapper.text()).toContain(t('account_security_updated'))
    expect(wrapper.text()).toContain('最喜欢的晶体？')
  })
})
