// 统一的环境变量读取（带默认值），供 vite.config.ts 使用。

/** 读取环境变量；未设置或为空时返回 fallback。 */
export function getEnv(env: Record<string, string>, key: string, fallback: string): string {
  const value = env[key]
  return value ? value : fallback
}

/** 读取数字环境变量；未设置或为空时返回 fallback。 */
export function getEnvNumber(env: Record<string, string>, key: string, fallback: number): number {
  const raw = getEnv(env, key, String(fallback))
  const value = Number(raw)
  if (Number.isNaN(value)) {
    throw new Error(`环境变量 ${key} 必须是数字，收到 "${raw}"。`)
  }
  return value
}
