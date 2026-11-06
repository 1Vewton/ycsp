import { beforeEach, describe, expect, it, vi } from 'vitest'

import {
  ApiError,
  cancelTask,
  createTask,
  fetchTask,
  fetchTaskResult,
  problemFromError,
} from '../src/api/tasks'
import { rockSaltResult, taskSnapshot } from './fixtures'

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

beforeEach(() => {
  fetchMock.mockReset()
  vi.stubGlobal('fetch', fetchMock)
})

describe('任务请求', () => {
  it('创建任务：POST /api/tasks，请求体只有化学式', async () => {
    const task = taskSnapshot('pending')
    fetchMock.mockResolvedValue(jsonResponse({ status: 'ok', code: 'task_created', task }))

    await expect(createTask('NaCl')).resolves.toEqual(task)
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/tasks',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify({ formula: 'NaCl' }),
      }),
    )
    expect(requestHeaders().get('Content-Type')).toBe('application/json')
    expect(requestHeaders().get('X-Request-ID')).toMatch(/^web-/)
  })

  it('查询状态：GET /api/tasks/:id', async () => {
    const task = taskSnapshot('running')
    fetchMock.mockResolvedValue(jsonResponse({ status: 'ok', code: 'task_ok', task }))

    await expect(fetchTask(7)).resolves.toEqual(task)
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/tasks/7',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(requestHeaders().get('X-Request-ID')).toMatch(/^web-/)
  })

  it('取消任务：POST /api/tasks/:id/cancel', async () => {
    const task = taskSnapshot('cancelled')
    fetchMock.mockResolvedValue(jsonResponse({ status: 'ok', code: 'task_cancelled', task }))

    await expect(cancelTask(7)).resolves.toEqual(task)
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/tasks/7/cancel',
      expect.objectContaining({ method: 'POST' }),
    )
    expect(requestHeaders().get('X-Request-ID')).toMatch(/^web-/)
  })

  it('取结构结果：GET /api/tasks/:id/result', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse({ status: 'ok', code: 'task_result', result: rockSaltResult }),
    )

    await expect(fetchTaskResult(7)).resolves.toEqual(rockSaltResult)
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/tasks/7/result',
      expect.objectContaining({ method: 'GET' }),
    )
    expect(requestHeaders().get('X-Request-ID')).toMatch(/^web-/)
  })
})

describe('失败如何处理', () => {
  it('后端错误信封换成 ApiError，保留 code 与 params', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(
        { status: 'error', code: 'formula_charge_unbalanced', params: { formula: 'NaCl2' } },
        false,
        400,
        { 'X-Request-ID': 'api-test-123' },
      ),
    )

    await expect(createTask('NaCl2')).rejects.toMatchObject({
      name: 'ApiError',
      requestId: 'api-test-123',
      problem: { code: 'formula_charge_unbalanced', params: { formula: 'NaCl2' } },
    })
  })

  it('HTTP 200 但 status 不是 ok 也算失败', async () => {
    fetchMock.mockResolvedValue(jsonResponse({ status: 'error', code: 'task_not_ready' }))

    await expect(fetchTaskResult(7)).rejects.toBeInstanceOf(ApiError)
  })

  it('应答不是 JSON（例如代理返回一页 HTML）时报 internal_error', async () => {
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    fetchMock.mockResolvedValue({
      ok: true,
      status: 200,
      headers: new Headers(),
      json: async () => {
        throw new SyntaxError('Unexpected token <')
      },
    } as unknown as Response)

    await expect(fetchTask(7)).rejects.toMatchObject({ problem: { code: 'internal_error' } })
    expect(errorSpy).toHaveBeenCalled()
  })

  it('网络不通时报 request_failed', async () => {
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    fetchMock.mockRejectedValue(new TypeError('Failed to fetch'))

    await expect(fetchTask(7)).rejects.toMatchObject({ problem: { code: 'request_failed' } })
    expect(errorSpy).toHaveBeenCalled()
  })

  it('成功信封里缺 task / result 时报 internal_error，不让页面拿到 undefined', async () => {
    fetchMock.mockResolvedValue(jsonResponse({ status: 'ok', code: 'task_ok' }))

    await expect(fetchTask(7)).rejects.toMatchObject({ problem: { code: 'internal_error' } })
    await expect(fetchTaskResult(7)).rejects.toMatchObject({
      problem: { code: 'internal_error' },
    })
  })
})

describe('problemFromError', () => {
  it('ApiError 直接取出 code 与 params', () => {
    const problem = { code: 'task_not_found', params: { task_id: 7 } }

    expect(problemFromError(new ApiError(problem))).toEqual(problem)
  })

  it('未知异常只留在控制台，界面拿到 request_failed', () => {
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

    expect(problemFromError(new TypeError('boom'))).toEqual({ code: 'request_failed' })
    expect(errorSpy).toHaveBeenCalledOnce()
  })
})
