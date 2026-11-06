// 三层统一的请求编号：前端每次请求生成，数据管理层、数据处理层原样沿用。
// 出问题时把控制台里的 request_id 发给老师或开发者，就能按编号查三层日志。
// 头名与长度限制与 backend/internal/traceid/traceid.go 保持一致。

export const REQUEST_ID_HEADER = 'X-Request-ID'

export const MAX_REQUEST_ID_LENGTH = 64

/** 生成一个前端请求编号；优先用浏览器提供的 crypto.randomUUID。 */
export function createRequestId(): string {
  const randomPart =
    globalThis.crypto?.randomUUID?.() ??
    `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 12)}`
  return `web-${randomPart}`.slice(0, MAX_REQUEST_ID_LENGTH)
}
