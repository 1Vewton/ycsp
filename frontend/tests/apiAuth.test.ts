import { beforeEach, describe, expect, it, vi } from 'vitest'

import {
  changePassword,
  fetchCurrentUser,
  fetchPasswordResetQuestion,
  registerAccount,
  resetPassword,
  signIn,
  signOut,
  updateSecurityQuestion,
  type AccountUser,
} from '../src/api/auth'
import { ApiError } from '../src/api/client'

const fetchMock = vi.fn()

function jsonResponse(
  body: unknown,
  ok = true,
  status = 200,
  headers: Record<string, string> = {},
): Response {
  return {
    ok,
    status,
    headers: new Headers(headers),
    json: async () => body,
  } as unknown as Response
}

function requestInit(callIndex = 0): RequestInit {
  return fetchMock.mock.calls[callIndex][1] as RequestInit
}

function requestHeaders(callIndex = 0): Headers {
  return new Headers(requestInit(callIndex).headers)
}

const account: AccountUser = {
  id: 3,
  display_name: '小明',
  created_at: '2026-11-02T00:00:00Z',
}

beforeEach(() => {
  fetchMock.mockReset()
  vi.stubGlobal('fetch', fetchMock)
})

describe('账号请求', () => {
  it('注册：POST /api/users，提交昵称、密码与可选安全问题', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse({ status: 'ok', code: 'user_registered', user: account }),
    )

    await expect(registerAccount('小明', 'Passw0rd!')).resolves.toEqual(account)
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/users',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({
          display_name: '小明',
          password: 'Passw0rd!',
          security_question: '',
          security_answer: '',
        }),
      }),
    )
    expect(requestHeaders().get('Content-Type')).toBe('application/json')
    expect(requestHeaders().get('X-Request-ID')).toMatch(/^web-/)
  })

  it('登录：POST /api/sessions', async () => {
    fetchMock.mockResolvedValue(jsonResponse({ status: 'ok', code: 'signed_in', user: account }))

    await expect(signIn('小明', 'Passw0rd!')).resolves.toEqual(account)
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/sessions',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ display_name: '小明', password: 'Passw0rd!' }),
      }),
    )
    expect(requestHeaders().get('X-Request-ID')).toMatch(/^web-/)
  })

  it('退出：DELETE /api/sessions', async () => {
    fetchMock.mockResolvedValue(jsonResponse({ status: 'ok', code: 'signed_out' }))

    await expect(signOut()).resolves.toBeUndefined()
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/sessions',
      expect.objectContaining({ method: 'DELETE' }),
    )
    expect(requestHeaders().get('X-Request-ID')).toMatch(/^web-/)
  })

  it('取当前用户：GET /api/users/me', async () => {
    fetchMock.mockResolvedValue(jsonResponse({ status: 'ok', code: 'user_ok', user: account }))

    await expect(fetchCurrentUser()).resolves.toEqual(account)
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/users/me',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(requestHeaders().get('X-Request-ID')).toMatch(/^web-/)
  })

  it('修改密码：PUT /api/users/me/password', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse({ status: 'ok', code: 'password_changed', user: account }),
    )

    await expect(changePassword('old-pass', 'new-pass')).resolves.toEqual(account)
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/users/me/password',
      expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({ current_password: 'old-pass', new_password: 'new-pass' }),
      }),
    )
  })

  it('设置安全问题：PUT /api/users/me/security-question', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse({ status: 'ok', code: 'security_question_updated', user: account }),
    )

    await expect(
      updateSecurityQuestion('old-pass', '最喜欢的元素？', '钠'),
    ).resolves.toEqual(account)
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/users/me/security-question',
      expect.objectContaining({
        method: 'PUT',
        body: JSON.stringify({
          current_password: 'old-pass',
          security_question: '最喜欢的元素？',
          security_answer: '钠',
        }),
      }),
    )
  })

  it('取安全问题：POST /api/password-resets/question', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse({
        status: 'ok',
        code: 'password_reset_question',
        security_question: '最喜欢的元素？',
      }),
    )

    await expect(fetchPasswordResetQuestion('小明')).resolves.toBe('最喜欢的元素？')
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/password-resets/question',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ display_name: '小明' }),
      }),
    )
  })

  it('重置密码：POST /api/password-resets', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse({ status: 'ok', code: 'password_reset', user: account }),
    )

    await expect(resetPassword('小明', '钠', 'new-pass')).resolves.toEqual(account)
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/password-resets',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({
          display_name: '小明',
          security_answer: '钠',
          new_password: 'new-pass',
        }),
      }),
    )
  })
})

describe('未登录与异常', () => {
  it('未登录时返回 null：401 是正常状态，不是错误', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse({ status: 'error', code: 'not_signed_in' }, false, 401),
    )

    await expect(fetchCurrentUser()).resolves.toBeNull()
  })

  it('登录服务异常时仍然抛错，由页面提示', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse({ status: 'error', code: 'session_unavailable' }, false, 500),
    )

    await expect(fetchCurrentUser()).rejects.toMatchObject({
      problem: { code: 'session_unavailable' },
    })
  })

  it('成功信封里缺 user 时报 internal_error', async () => {
    fetchMock.mockResolvedValue(jsonResponse({ status: 'ok', code: 'signed_in' }))

    await expect(signIn('小明', 'Passw0rd!')).rejects.toBeInstanceOf(ApiError)
  })

  it('密码规则与参数原样保留，供页面逐条翻译', async () => {
    const params = {
      rules: ['password_rule_missing_digit', 'password_rule_missing_symbol'],
      min_length: 8,
      max_bytes: 72,
    }
    fetchMock.mockResolvedValue(
      jsonResponse(
        { status: 'error', code: 'password_too_weak', params },
        false,
        400,
        { 'X-Request-ID': 'api-auth-123' },
      ),
    )

    await expect(registerAccount('小明', 'abcdefgh')).rejects.toMatchObject({
      requestId: 'api-auth-123',
      problem: { code: 'password_too_weak', params },
    })
  })
})
