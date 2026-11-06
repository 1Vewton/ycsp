// 登录状态：全站共用一份“当前账号”，页面只读它决定显示登录入口还是昵称。
// 会话本身存在后端的 HttpOnly Cookie 里；这里只缓存账号信息，刷新后由 restoreSession 重新取回。

import { readonly, ref } from 'vue'

import {
  fetchCurrentUser,
  registerAccount as requestRegistration,
  signIn as requestSignIn,
  signOut as requestSignOut,
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
): Promise<AccountUser> {
  const user = await requestRegistration(displayName, password)
  activeUser.value = user
  hasRestored.value = true
  return user
}

/** 退出登录：只有后端确认成功后才清掉本地账号，避免界面与 Cookie 状态不一致。 */
export async function signOut(): Promise<void> {
  await requestSignOut()
  activeUser.value = null
  hasRestored.value = true
}
