// 数据管理层应答的公共处理：统一信封、错误转换与“把错误变成可翻译的问题”。
// 任务与账号接口共用这一份，避免各写一遍导致行为漂移。

import { createRequestId, REQUEST_ID_HEADER } from './requestId'

/** 一条可翻译的问题：code 决定说哪句话，params 是要填进句子的值。 */
export interface ApiProblem {
  code: string
  params?: Record<string, unknown>
}

/** 应答里与具体业务无关的字段（见 backend/README.md《对外接口》）。 */
export interface ApiSummary {
  status?: string
  code?: string
  params?: Record<string, unknown>
}

export type ApiEnvelope<T extends object> = ApiSummary & Partial<T>

/** 后端拒绝请求或网络失败时抛出的错误。requestId 用来在控制台与三层日志里对齐。 */
export class ApiError extends Error {
  readonly problem: ApiProblem
  readonly requestId: string

  constructor(problem: ApiProblem, requestId = '') {
    super(problem.code)
    this.name = 'ApiError'
    this.problem = problem
    this.requestId = requestId
  }
}

/** 把任意异常换成可翻译的问题；技术细节只留在控制台。 */
export function problemFromError(error: unknown): ApiProblem {
  if (error instanceof ApiError) {
    return error.problem
  }
  console.error('[ycsp] unexpected error', error)
  return { code: 'request_failed' }
}

/**
 * 发请求并读出统一信封：HTTP 失败或 status 不是 ok 时抛 ApiError。
 *
 * 每次请求都会生成 X-Request-ID：后端原样沿用并写回响应头，
 * 异步任务也会把它带进数据处理层，三层日志可按同一个编号搜索。
 * 会话 Cookie 由浏览器自动携带（HttpOnly，脚本读不到）；前端与数据管理层同源。
 */
export async function requestApi<T extends object>(
  url: string,
  init: RequestInit,
): Promise<ApiEnvelope<T>> {
  const requestId = createRequestId()
  const headers = new Headers(init.headers)
  headers.set(REQUEST_ID_HEADER, requestId)

  const method = init.method ?? 'GET'
  const startedAt = Date.now()
  console.debug(`[ycsp] request start request_id=${requestId} method=${method} path=${url}`)

  let response: Response
  try {
    response = await fetch(url, { ...init, headers })
  } catch (error) {
    console.error(
      `[ycsp] request failed request_id=${requestId} method=${method} path=${url}`,
      error,
    )
    throw new ApiError({ code: 'request_failed' }, requestId)
  }

  // 后端会回写编号；如果代理把它丢了，就用前端自己生成的那个。
  const responseRequestId = response.headers?.get(REQUEST_ID_HEADER) ?? requestId
  const envelope = await readEnvelope<T>(response, responseRequestId)
  console.debug(
    `[ycsp] request done request_id=${responseRequestId} method=${method} path=${url}` +
      ` status=${response.status} duration_ms=${Date.now() - startedAt}`,
  )

  if (!response.ok || envelope.status !== 'ok') {
    throw new ApiError(
      {
        code: envelope.code ?? 'request_failed',
        params: envelope.params,
      },
      responseRequestId,
    )
  }
  return envelope
}

// 应答不是 JSON 时（例如代理返回了一页 HTML），也给出可翻译的错误，
// 而不是让页面抛出一个学生看不懂的语法异常。
async function readEnvelope<T extends object>(
  response: Response,
  requestId: string,
): Promise<ApiEnvelope<T>> {
  try {
    return (await response.json()) as ApiEnvelope<T>
  } catch (error) {
    console.error(`[ycsp] response is not JSON request_id=${requestId}`, error)
    // 读不出 JSON 时没有业务字段可填；调用方会因 status 不是 ok 而抛 internal_error。
    return { status: 'error', code: 'internal_error' } as unknown as ApiEnvelope<T>
  }
}

