import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { AccountUser } from '../src/api/auth'

// 只替掉 HTTP 层：登录页与 session 的真实逻辑一起验证。
const authMocks = vi.hoisted(() => ({
  fetchCurrentUser: vi.fn(),
  registerAccount: vi.fn(),
  signIn: vi.fn(),
  signOut: vi.fn(),
}))

vi.mock('../src/api/auth', () => authMocks)

import { ApiError } from '../src/api/client'
import { t, tCode } from '../src/locales'
import * as session from '../src/session'
import LoginView from '../src/views/LoginView.vue'

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

  // session 是模块级状态：先退出一次，让每个用例都从“已检查、未登录”开始。
  await session.signOut()
  // 这次退出只是准备状态，不算用例里的操作。
  authMocks.signOut.mockClear()
})

async function mountLogin(path = '/login') {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div />' } },
      { path: '/predict', component: { template: '<div />' } },
      { path: '/result/:taskId', component: { template: '<div />' } },
      { path: '/login', component: LoginView },
    ],
  })
  await router.push(path)
  await router.isReady()

  const wrapper = mount(LoginView, { global: { plugins: [router] } })
  await flushPromises()
  return { wrapper, router }
}

async function fillCredentials(
  wrapper: Awaited<ReturnType<typeof mountLogin>>['wrapper'],
  displayName = '小明',
  password = 'Passw0rd!',
): Promise<void> {
  await wrapper.find('#login-display-name').setValue(displayName)
  await wrapper.find('#login-password').setValue(password)
}

describe('LoginView 表单', () => {
  it('默认是登录模式；注册模式的密码框语义正确', async () => {
    const { wrapper } = await mountLogin()
    const modeButtons = wrapper.findAll('.mode-button')

    expect(wrapper.text()).toContain(t('login_heading'))
    expect(modeButtons[0].attributes('aria-pressed')).toBe('true')
    expect(modeButtons[1].attributes('aria-pressed')).toBe('false')
    expect(wrapper.find('#login-password-confirmation').exists()).toBe(false)
    expect(wrapper.find('#login-password').attributes('autocomplete')).toBe('current-password')

    await modeButtons[1].trigger('click')

    expect(wrapper.find('#login-password-confirmation').exists()).toBe(true)
    expect(wrapper.text()).toContain(t('login_password_policy_hint'))
    expect(wrapper.find('#login-password').attributes('autocomplete')).toBe('new-password')
  })

  it('什么都没填就提交时给出提示', async () => {
    const { wrapper } = await mountLogin()

    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(wrapper.find('[role="alert"]').text()).toContain(tCode('auth_fields_required'))
    expect(authMocks.signIn).not.toHaveBeenCalled()
  })

  it('登录成功：去掉昵称首尾空格，然后去预测页', async () => {
    authMocks.signIn.mockResolvedValue(account)
    const { wrapper, router } = await mountLogin()

    await fillCredentials(wrapper, '  小明  ')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(authMocks.signIn).toHaveBeenCalledWith('小明', 'Passw0rd!')
    expect(router.currentRoute.value.fullPath).toBe('/predict')
  })

  it('登录失败时显示后端给的可翻译原因', async () => {
    authMocks.signIn.mockRejectedValue(new ApiError({ code: 'invalid_credentials' }))
    const { wrapper } = await mountLogin()

    await fillCredentials(wrapper)
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(wrapper.find('[role="alert"]').text()).toContain(tCode('invalid_credentials'))
  })
})

describe('LoginView 注册', () => {
  it('两次密码不一致时本地拦下，不发请求', async () => {
    const { wrapper } = await mountLogin()
    await wrapper.findAll('.mode-button')[1].trigger('click')

    await fillCredentials(wrapper)
    await wrapper.find('#login-password-confirmation').setValue('Different1!')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(authMocks.registerAccount).not.toHaveBeenCalled()
    expect(wrapper.find('[role="alert"]').text()).toContain(tCode('auth_password_mismatch'))
  })

  it('注册成功：调用注册接口并去预测页', async () => {
    authMocks.registerAccount.mockResolvedValue(account)
    const { wrapper, router } = await mountLogin()
    await wrapper.findAll('.mode-button')[1].trigger('click')

    await fillCredentials(wrapper, ' 小明 ')
    await wrapper.find('#login-password-confirmation').setValue('Passw0rd!')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(authMocks.registerAccount).toHaveBeenCalledWith('小明', 'Passw0rd!')
    expect(router.currentRoute.value.fullPath).toBe('/predict')
  })

  it('密码不合规时逐条显示后端返回的规则', async () => {
    const params = {
      rules: ['password_rule_missing_digit', 'password_rule_missing_symbol'],
      min_length: 8,
      max_bytes: 72,
    }
    authMocks.registerAccount.mockRejectedValue(
      new ApiError({ code: 'password_too_weak', params }),
    )
    const { wrapper } = await mountLogin()
    await wrapper.findAll('.mode-button')[1].trigger('click')

    await fillCredentials(wrapper, '小明', 'abcdefgh')
    await wrapper.find('#login-password-confirmation').setValue('abcdefgh')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    const alertText = wrapper.find('[role="alert"]').text()
    expect(alertText).toContain(tCode('password_too_weak', params))
    expect(alertText).toContain(tCode('password_rule_missing_digit', params))
    expect(alertText).toContain(tCode('password_rule_missing_symbol', params))
  })
})

describe('LoginView 跳转与已登录状态', () => {
  it('登录后回到 redirect 指定的站内页面', async () => {
    authMocks.signIn.mockResolvedValue(account)
    const { wrapper, router } = await mountLogin('/login?redirect=/result/7')

    await fillCredentials(wrapper)
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(router.currentRoute.value.fullPath).toBe('/result/7')
  })

  it('站外 redirect 被忽略，只回站内页面', async () => {
    authMocks.signIn.mockResolvedValue(account)
    const { wrapper, router } = await mountLogin('/login?redirect=//evil.example/steal')

    await fillCredentials(wrapper)
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(router.currentRoute.value.fullPath).toBe('/predict')
  })

  it('redirect 指回登录页本身时改去预测页，避免登录后原地打转', async () => {
    authMocks.signIn.mockResolvedValue(account)
    const { wrapper, router } = await mountLogin('/login?redirect=/login%3Fredirect%3D%2Fpredict')

    await fillCredentials(wrapper)
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(router.currentRoute.value.fullPath).toBe('/predict')
  })

  it('已登录时显示账号面板，退出后回到登录表单', async () => {
    authMocks.signIn.mockResolvedValue(account)
    await session.signIn('小明', 'Passw0rd!')

    const { wrapper } = await mountLogin()

    expect(wrapper.text()).toContain(t('auth_signed_in_title'))
    expect(wrapper.text()).toContain(account.display_name)

    await wrapper.find('.secondary-button').trigger('click')
    await flushPromises()

    expect(authMocks.signOut).toHaveBeenCalledOnce()
    expect(wrapper.find('#login-display-name').exists()).toBe(true)
  })
})

