import { describe, expect, it } from 'vitest'

import {
  createRequestId,
  MAX_REQUEST_ID_LENGTH,
  REQUEST_ID_HEADER,
} from '../src/api/requestId'

describe('createRequestId', () => {
  it('生成 web- 前缀、长度不超限的请求编号', () => {
    const requestId = createRequestId()

    expect(requestId).toMatch(/^web-[A-Za-z0-9._:-]+$/)
    expect(requestId.length).toBeLessThanOrEqual(MAX_REQUEST_ID_LENGTH)
  })

  it('每次生成的编号不同', () => {
    expect(createRequestId()).not.toBe(createRequestId())
  })

  it('头名与数据管理层、数据处理层一致', () => {
    expect(REQUEST_ID_HEADER).toBe('X-Request-ID')
  })
})
