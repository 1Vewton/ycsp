import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { AccountUser } from '../src/api/auth'
import type { TaskHistoryPage } from '../src/api/tasks'

const authMocks = vi.hoisted(() => ({
  fetchCurrentUser: vi.fn(),
  registerAccount: vi.fn(),
  signIn: vi.fn(),
  signOut: vi.fn(),
  changePassword: vi.fn(),
  updateSecurityQuestion: vi.fn(),
  fetchPasswordResetQuestion: vi.fn(),
  resetPassword: vi.fn(),
}))

const taskMocks = vi.hoisted(() => ({
  fetchTaskHistory: vi.fn(),
}))

vi.mock('../src/api/auth', () => authMocks)
vi.mock('../src/api/tasks', () => taskMocks)

import { ApiError } from '../src/api/client'
import { t, tCode } from '../src/locales'
import * as session from '../src/session'
import HistoryView from '../src/views/HistoryView.vue'

const account: AccountUser = {
  id: 3,
  display_name: '小明',
  created_at: '2026-11-02T00:00:00Z',
}

const firstPage: TaskHistoryPage = {
  tasks: [
    {
      id: 7,
      formula: 'NaCl',
      status: 'succeeded',
      created_at: '2026-11-07T08:00:00Z',
      finished_at: '2026-11-07T08:00:01Z',
      duration_ms: 20,
      result: { family: 'ionic', prototype: 'rock_salt', coordination_number: 6 },
    },
    {
      id: 6,
      formula: 'CaF2',
      status: 'failed',
      created_at: '2026-11-06T08:00:00Z',
      finished_at: '2026-11-06T08:00:01Z',
      duration_ms: 30,
      error: { code: 'composition_not_supported', params: { formula: 'CaF2' } },
    },
  ],
  total: 3,
  limit: 20,
  offset: 0,
}

beforeEach(async () => {
  for (const mockFunction of Object.values({ ...authMocks, ...taskMocks })) {
    mockFunction.mockReset()
  }
  authMocks.fetchCurrentUser.mockResolvedValue(null)
  authMocks.signOut.mockResolvedValue(undefined)
  taskMocks.fetchTaskHistory.mockResolvedValue({
    tasks: [],
    total: 0,
    limit: 20,
    offset: 0,
  })

  await session.signOut()
  authMocks.signOut.mockClear()
})

async function mountHistory() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div />' } },
      { path: '/login', name: 'login', component: { template: '<div />' } },
      { path: '/predict', component: { template: '<div />' } },
      { path: '/result/:taskId', name: 'result', component: { template: '<div />' } },
      { path: '/history', name: 'history', component: HistoryView },
    ],
  })
  await router.push('/history')
  await router.isReady()

  const wrapper = mount(HistoryView, { global: { plugins: [router] } })
  await flushPromises()
  return { wrapper, router }
}

async function signInAccount(): Promise<void> {
  authMocks.signIn.mockResolvedValue(account)
  await session.signIn('小明', 'Passw0rd!')
}

describe('HistoryView 预测记录', () => {
  it('未登录时提示先登录', async () => {
    const { wrapper } = await mountHistory()

    expect(wrapper.text()).toContain(t('history_sign_in_hint'))
    expect(wrapper.text()).toContain(t('history_sign_in_link'))
  })

  it('登录后显示任务、状态与结构摘要', async () => {
    await signInAccount()
    taskMocks.fetchTaskHistory.mockResolvedValue(firstPage)
    const { wrapper } = await mountHistory()

    expect(taskMocks.fetchTaskHistory).toHaveBeenCalledWith(20, 0)
    expect(wrapper.text()).toContain('NaCl')
    expect(wrapper.text()).toContain(t('task_status_succeeded'))
    expect(wrapper.text()).toContain(t('prototype_rock_salt'))
    expect(wrapper.text()).toContain('CaF2')
    expect(wrapper.text()).toContain(tCode('composition_not_supported', { formula: 'CaF2' }))
    expect(wrapper.text()).toContain(t('history_showing', { shown: 2, total: 3 }))
  })

  it('还有更多记录时加载下一页并追加', async () => {
    await signInAccount()
    taskMocks.fetchTaskHistory
      .mockResolvedValueOnce(firstPage)
      .mockResolvedValueOnce({
        tasks: [
          {
            id: 5,
            formula: 'ZnO',
            status: 'succeeded',
            created_at: '2026-11-05T08:00:00Z',
            finished_at: '2026-11-05T08:00:01Z',
            duration_ms: 15,
            result: { family: 'ionic', prototype: 'wurtzite', coordination_number: 4 },
          },
        ],
        total: 3,
        limit: 20,
        offset: 2,
      })
    const { wrapper } = await mountHistory()

    const loadMoreButton = wrapper
      .findAll('button')
      .find((button) => button.text() === t('history_load_more'))
    expect(loadMoreButton).toBeDefined()
    await loadMoreButton?.trigger('click')
    await flushPromises()

    expect(taskMocks.fetchTaskHistory).toHaveBeenLastCalledWith(20, 2)
    expect(wrapper.text()).toContain('ZnO')
    expect(wrapper.text()).toContain(t('prototype_wurtzite'))
  })

  it('加载失败时显示可翻译的原因并允许重试', async () => {
    await signInAccount()
    taskMocks.fetchTaskHistory.mockRejectedValue(new ApiError({ code: 'request_failed' }))
    const { wrapper } = await mountHistory()

    expect(wrapper.text()).toContain(tCode('request_failed'))
    expect(
      wrapper.findAll('button').some((button) => button.text() === t('result_retry_button')),
    ).toBe(true)
  })
})
