// 账号接口的前端客户端：注册、登录、退出与“当前用户”。
// 失败时抛 ApiError，页面用 code 从语言包取文案（见 ./client.ts）。

import { ApiError, requestApi } from './client'

/** 当前登录账号；后端不会返回密码哈希。 */
export interface AccountUser {
  id: number
  display_name: string
  created_at: string
}

interface UserEnvelope {
  user?: AccountUser
}

/** 注册并直接登录（后端会下发会话 Cookie）。 */
export async function registerAccount(
  displayName: string,
  password: string,
): Promise<AccountUser> {
  const envelope = await requestApi<UserEnvelope>(
    '/api/users',
    credentialsRequest(displayName, password),
  )
  return requireUser(envelope)
}

/** 登录；连续失败达到阈值时后端返回 too_many_attempts。 */
export async function signIn(displayName: string, password: string): Promise<AccountUser> {
  const envelope = await requestApi<UserEnvelope>(
    '/api/sessions',
    credentialsRequest(displayName, password),
  )
  return requireUser(envelope)
}

/** 退出登录；后端是幂等的，本来没登录也算成功。 */
export async function signOut(): Promise<void> {
  await requestApi<object>('/api/sessions', { method: 'DELETE' })
}

/** 取当前登录用户；未登录返回 null（这是正常情况，不算错误）。 */
export async function fetchCurrentUser(): Promise<AccountUser | null> {
  try {
    const envelope = await requestApi<UserEnvelope>('/api/users/me', { method: 'GET' })
    return requireUser(envelope)
  } catch (error) {
    if (error instanceof ApiError && error.problem.code === 'not_signed_in') {
      return null
    }
    throw error
  }
}

function credentialsRequest(displayName: string, password: string): RequestInit {
  return {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ display_name: displayName, password }),
  }
}

function requireUser(envelope: UserEnvelope): AccountUser {
  if (envelope.user === undefined) {
    throw new ApiError({ code: 'internal_error' })
  }
  return envelope.user
}
