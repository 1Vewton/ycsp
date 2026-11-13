// 登录状态：全站共用一份“当前账号”，页面只读它决定显示登录入口还是昵称。
// 会话本身存在后端的 HttpOnly Cookie 里；这里只缓存账号信息，刷新后由 restoreSession 重新取回。

import { readonly, ref } from 'vue'

import {
  changePassword as requestChangePassword,
  fetchCurrentUser,
  fetchPasswordResetQuestion,
  registerAccount as requestRegistration,
  resetPassword as requestPasswordReset,
  signIn as requestSignIn,
  signOut as requestSignOut,
  updateSecurityQuestion as requestSecurityQuestionUpdate,
  type AccountUser,
} from './api/auth'

const activeUser = ref<AccountUser | null>(null)
const isRestoring = ref(false)
const hasRestored = ref(false)

// App 与登录页可能同时挂载；共用同一个请求，避免重复打 /api/users/me。
let restorePromise: Promise<void> | null = null

/** 当前登录用户；未登录为 null。 */
export const signedInUser = readonly(activeUser)

/** 是否正在检查登录状态（首次打开页面时）。 */
export const isRestoringSession = readonly(isRestoring)

/** 从后端取回当前账号；成功执行一次后，由登录 / 退出负责更新。 */
export function restoreSession(): Promise<void> {
  if (hasRestored.value) {
    return Promise.resolve()
  }
  if (restorePromise === null) {
    isRestoring.value = true
    restorePromise = fetchCurrentUser()
      .then((user) => {
        activeUser.value = user
      })
      .catch((error: unknown) => {
        // 检查登录状态失败不挡学生预测：当作未登录，细节只留在控制台。
        console.error(error)
        activeUser.value = null
      })
      .finally(() => {
        isRestoring.value = false
        hasRestored.value = true
      })
  }
  return restorePromise
}

/** 登录成功后全站立刻显示当前账号。 */
export async function signIn(displayName: string, password: string): Promise<AccountUser> {
  const user = await requestSignIn(displayName, password)
  activeUser.value = user
  hasRestored.value = true
  return user
}

/** 注册：后端注册成功即登录。 */
export async function registerAccount(
  displayName: string,
  password: string,
  securityQuestion = '',
  securityAnswer = '',
): Promise<AccountUser> {
  const user = await requestRegistration(
    displayName,
    password,
    securityQuestion,
    securityAnswer,
  )
  rememberSignedInUser(user)
  return user
}

/** 退出登录：只有后端确认成功后才清掉本地账号，避免界面与 Cookie 状态不一致。 */
export async function signOut(): Promise<void> {
  await requestSignOut()
  activeUser.value = null
  hasRestored.value = true
}

/** 登录后修改密码；当前会话保持不变。 */
export async function changePassword(
  currentPassword: string,
  newPassword: string,
): Promise<AccountUser> {
  const user = await requestChangePassword(currentPassword, newPassword)
  rememberSignedInUser(user)
  return user
}

/** 登录后设置或更新安全问题；需要当前密码。 */
export async function updateSecurityQuestion(
  currentPassword: string,
  securityQuestion: string,
  securityAnswer: string,
): Promise<AccountUser> {
  const user = await requestSecurityQuestionUpdate(
    currentPassword,
    securityQuestion,
    securityAnswer,
  )
  rememberSignedInUser(user)
  return user
}

/** 忘记密码时按昵称取安全问题。 */
export function lookupSecurityQuestion(displayName: string): Promise<string> {
  return fetchPasswordResetQuestion(displayName)
}

/** 用安全问题答案重置密码；后端会直接登录，这里同步全站账号状态。 */
export async function resetPassword(
  displayName: string,
  securityAnswer: string,
  newPassword: string,
): Promise<AccountUser> {
  const user = await requestPasswordReset(displayName, securityAnswer, newPassword)
  rememberSignedInUser(user)
  return user
}

function rememberSignedInUser(user: AccountUser): void {
  activeUser.value = user
  hasRestored.value = true
}
