import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { AccountUser } from '../src/api/auth'

// 只替掉 HTTP 层；session 模块的真实逻辑（缓存、并发合并）由本文件验证。
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

const account: AccountUser = {
  id: 3,
  display_name: '小明',
  created_at: '2026-11-02T00:00:00Z',
}

// session 的当前账号是模块级状态；每个用例重新加载一份，互不影响。
async function loadSession() {
  vi.resetModules()
  return import('../src/session')
}

beforeEach(() => {
  for (const mockFunction of Object.values(authMocks)) {
    mockFunction.mockReset()
  }
  authMocks.fetchCurrentUser.mockResolvedValue(null)
})

describe('restoreSession', () => {
  it('取回已登录账号，并发调用只请求一次', async () => {
    authMocks.fetchCurrentUser.mockResolvedValue(account)
    const session = await loadSession()

    await Promise.all([session.restoreSession(), session.restoreSession()])

    expect(authMocks.fetchCurrentUser).toHaveBeenCalledTimes(1)
    expect(session.signedInUser.value).toEqual(account)

    // 已经检查过，再调用不会重复打接口。
    await session.restoreSession()
    expect(authMocks.fetchCurrentUser).toHaveBeenCalledTimes(1)
  })

  it('未登录时保持 null', async () => {
    const session = await loadSession()

    await session.restoreSession()

    expect(session.signedInUser.value).toBeNull()
  })

  it('检查失败不抛出：按未登录处理，细节留在控制台', async () => {
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    authMocks.fetchCurrentUser.mockRejectedValue(new Error('network down'))
    const session = await loadSession()

    await expect(session.restoreSession()).resolves.toBeUndefined()

    expect(session.signedInUser.value).toBeNull()
    expect(errorSpy).toHaveBeenCalled()
  })
})

describe('登录、注册与退出', () => {
  it('登录成功后更新当前账号', async () => {
    authMocks.signIn.mockResolvedValue(account)
    const session = await loadSession()

    await expect(session.signIn('小明', 'Passw0rd!')).resolves.toEqual(account)
    expect(session.signedInUser.value).toEqual(account)
  })

  it('注册成功等同于登录', async () => {
    authMocks.registerAccount.mockResolvedValue(account)
    const session = await loadSession()

    await session.registerAccount('小明', 'Passw0rd!')

    expect(session.signedInUser.value).toEqual(account)
  })

  it('修改密码后保持当前账号', async () => {
    authMocks.changePassword.mockResolvedValue(account)
    const session = await loadSession()

    await expect(session.changePassword('old-pass', 'new-pass')).resolves.toEqual(account)

    expect(authMocks.changePassword).toHaveBeenCalledWith('old-pass', 'new-pass')
    expect(session.signedInUser.value).toEqual(account)
  })

  it('更新安全问题后刷新账号信息', async () => {
    const updated = { ...account, security_question: '最喜欢的元素？' }
    authMocks.updateSecurityQuestion.mockResolvedValue(updated)
    const session = await loadSession()

    await expect(
      session.updateSecurityQuestion('old-pass', '最喜欢的元素？', '钠'),
    ).resolves.toEqual(updated)

    expect(session.signedInUser.value).toEqual(updated)
  })

  it('重置密码成功后直接登录', async () => {
    authMocks.resetPassword.mockResolvedValue(account)
    const session = await loadSession()

    await expect(session.resetPassword('小明', '钠', 'new-pass')).resolves.toEqual(account)

    expect(session.signedInUser.value).toEqual(account)
  })

  it('退出成功后清空当前账号', async () => {
    authMocks.fetchCurrentUser.mockResolvedValue(account)
    authMocks.signOut.mockResolvedValue(undefined)
    const session = await loadSession()
    await session.restoreSession()

    await session.signOut()

    expect(session.signedInUser.value).toBeNull()
  })

  it('退出失败时保留账号，避免界面与 Cookie 状态不一致', async () => {
    authMocks.fetchCurrentUser.mockResolvedValue(account)
    authMocks.signOut.mockRejectedValue(new Error('session service down'))
    const session = await loadSession()
    await session.restoreSession()

    await expect(session.signOut()).rejects.toThrow()
    expect(session.signedInUser.value).toEqual(account)
  })
})
