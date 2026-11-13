// 账号接口的前端客户端：注册、登录、退出与“当前用户”。
// 失败时抛 ApiError，页面用 code 从语言包取文案（见 ./client.ts）。

import { ApiError, requestApi } from './client'

/** 当前登录账号；后端不会返回密码哈希或安全答案哈希。 */
export interface AccountUser {
  id: number
  display_name: string
  security_question?: string
  created_at: string
}

interface UserEnvelope {
  user?: AccountUser
}

interface ResetQuestionEnvelope {
  security_question?: string
}

/** 注册并直接登录（后端会下发会话 Cookie）。 */
export async function registerAccount(
  displayName: string,
  password: string,
  securityQuestion = '',
  securityAnswer = '',
): Promise<AccountUser> {
  const envelope = await requestApi<UserEnvelope>('/api/users', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      display_name: displayName,
      password,
      security_question: securityQuestion,
      security_answer: securityAnswer,
    }),
  })
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

/** 登录后修改密码；成功返回更新后的账号。 */
export async function changePassword(
  currentPassword: string,
  newPassword: string,
): Promise<AccountUser> {
  const envelope = await requestApi<UserEnvelope>('/api/users/me/password', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      current_password: currentPassword,
      new_password: newPassword,
    }),
  })
  return requireUser(envelope)
}

/** 登录后设置或更新安全问题；需要当前密码。 */
export async function updateSecurityQuestion(
  currentPassword: string,
  securityQuestion: string,
  securityAnswer: string,
): Promise<AccountUser> {
  const envelope = await requestApi<UserEnvelope>('/api/users/me/security-question', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      current_password: currentPassword,
      security_question: securityQuestion,
      security_answer: securityAnswer,
    }),
  })
  return requireUser(envelope)
}

/** 忘记密码时按昵称取安全问题；账号不存在或没设置问题时后端返回 password_reset_unavailable。 */
export async function fetchPasswordResetQuestion(displayName: string): Promise<string> {
  const envelope = await requestApi<ResetQuestionEnvelope>('/api/password-resets/question', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ display_name: displayName }),
  })
  if (envelope.security_question === undefined) {
    throw new ApiError({ code: 'internal_error' })
  }
  return envelope.security_question
}

/** 用安全问题答案重置密码；成功后后端直接登录。 */
export async function resetPassword(
  displayName: string,
  securityAnswer: string,
  newPassword: string,
): Promise<AccountUser> {
  const envelope = await requestApi<UserEnvelope>('/api/password-resets', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      display_name: displayName,
      security_answer: securityAnswer,
      new_password: newPassword,
    }),
  })
  return requireUser(envelope)
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
