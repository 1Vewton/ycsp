import { enableAutoUnmount, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

// 只替掉三个请求函数，保留 ApiError / problemFromError 的真实实现：
// 这样失败路径与生产代码走同一套“错误 → 可翻译问题”的转换。
vi.mock('../src/api/tasks', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../src/api/tasks')>()
  return {
    ...actual,
    fetchTask: vi.fn(),
    fetchTaskResult: vi.fn(),
    cancelTask: vi.fn(),
  }
})

import { ApiError, cancelTask, fetchTask, fetchTaskResult } from '../src/api/tasks'
import { t, tCode } from '../src/locales'
import ResultView from '../src/views/ResultView.vue'
import { rockSaltResult, taskSnapshot } from './fixtures'

const fetchTaskMock = vi.mocked(fetchTask)
const fetchTaskResultMock = vi.mocked(fetchTaskResult)
const cancelTaskMock = vi.mocked(cancelTask)

enableAutoUnmount(afterEach)

// 结果页每 1 秒轮询一次；用假时钟把等待变成显式推进，测试不用真的等。
beforeEach(() => {
  vi.useFakeTimers()
  fetchTaskMock.mockReset()
  fetchTaskResultMock.mockReset()
  cancelTaskMock.mockReset()
})

afterEach(() => {
  vi.useRealTimers()
})

async function mountResultView(taskId = '7') {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div />' } },
      { path: '/predict', component: { template: '<div />' } },
      { path: '/result/:taskId', component: ResultView },
    ],
  })
  await router.push(`/result/${taskId}`)
  await router.isReady()

  const wrapper = mount(ResultView, {
    global: {
      plugins: [router],
      // Three.js 需要 WebGL；组件测试只关心“结果有没有交给查看器”。
      stubs: { StructureViewer: true },
    },
  })
  await settle()
  return wrapper
}

// 假时钟下不能等真实定时器：推进 0 ms 并让挂起的 Promise 链跑完。
async function settle() {
  await vi.advanceTimersByTimeAsync(0)
}

describe('ResultView', () => {
  it('排队中可取消，轮询到 running → succeeded 后取结构结果', async () => {
    fetchTaskMock
      .mockResolvedValueOnce(taskSnapshot('pending'))
      .mockResolvedValueOnce(taskSnapshot('running'))
      .mockResolvedValueOnce(taskSnapshot('succeeded', { duration_ms: 150 }))
    fetchTaskResultMock.mockResolvedValue(rockSaltResult)

    const wrapper = await mountResultView()

    expect(wrapper.text()).toContain(t('task_status_pending'))
    expect(wrapper.find('.cancel-button').exists()).toBe(true)
    expect(fetchTaskMock).toHaveBeenCalledWith(7)

    await vi.advanceTimersByTimeAsync(1000)
    expect(wrapper.text()).toContain(t('task_status_running'))
    expect(wrapper.find('.cancel-button').exists()).toBe(false)

    await vi.advanceTimersByTimeAsync(1000)
    expect(wrapper.text()).toContain(t('task_status_succeeded'))
    expect(fetchTaskResultMock).toHaveBeenCalledWith(7)
    expect(wrapper.text()).toContain(t('result_duration_seconds', { seconds: '0.15' }))
    expect(wrapper.find('structure-viewer-stub').exists()).toBe(true)

    // 已结束的任务不再轮询。
    await vi.advanceTimersByTimeAsync(5000)
    expect(fetchTaskMock).toHaveBeenCalledTimes(3)
  })

  it('失败时显示可翻译的原因，并隐藏三步进度', async () => {
    fetchTaskMock.mockResolvedValueOnce(
      taskSnapshot('failed', {
        error: { code: 'composition_not_supported', params: { formula: 'H2O' } },
        duration_ms: 20,
      }),
    )

    const wrapper = await mountResultView()

    expect(wrapper.find('[role="alert"]').text()).toContain(
      tCode('composition_not_supported', { formula: 'H2O' }),
    )
    expect(wrapper.text()).toContain(t('result_failed_hint'))
    expect(wrapper.find('.task-steps').exists()).toBe(false)
  })

  it('计算超时时显示“等太久”的提示，而不是内部 code', async () => {
    fetchTaskMock.mockResolvedValueOnce(
      taskSnapshot('failed', {
        error: { code: 'compute_timeout' },
        duration_ms: 10500,
      }),
    )

    const wrapper = await mountResultView()

    expect(wrapper.find('[role="alert"]').text()).toContain(tCode('compute_timeout'))
    expect(wrapper.text()).not.toContain('compute_timeout')
  })

  it('任务不存在时提示重试，点“再试一次”重新拉状态', async () => {
    fetchTaskMock
      .mockRejectedValueOnce(new ApiError({ code: 'task_not_found', params: { task_id: '7' } }))
      .mockResolvedValueOnce(taskSnapshot('succeeded'))
    fetchTaskResultMock.mockResolvedValue(rockSaltResult)

    const wrapper = await mountResultView()

    expect(wrapper.find('[role="alert"]').text()).toContain(
      tCode('task_not_found', { task_id: '7' }),
    )

    await wrapper.find('.retry-button').trigger('click')
    await settle()

    expect(fetchTaskMock).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain(t('task_status_succeeded'))
  })

  it('地址里的任务编号不是正整数时，不发请求也不显示重试按钮', async () => {
    const wrapper = await mountResultView('abc')

    expect(fetchTaskMock).not.toHaveBeenCalled()
    expect(wrapper.find('[role="alert"]').text()).toContain(
      tCode('task_not_found', { task_id: 'abc' }),
    )
    expect(wrapper.find('.retry-button').exists()).toBe(false)
  })

  it('取消排队中的任务后显示已取消，并停止轮询', async () => {
    fetchTaskMock.mockResolvedValueOnce(taskSnapshot('pending'))
    cancelTaskMock.mockResolvedValueOnce(taskSnapshot('cancelled'))

    const wrapper = await mountResultView()
    await wrapper.find('.cancel-button').trigger('click')
    await settle()

    expect(cancelTaskMock).toHaveBeenCalledWith(7)
    expect(wrapper.text()).toContain(t('task_status_cancelled'))

    await vi.advanceTimersByTimeAsync(3000)
    expect(fetchTaskMock).toHaveBeenCalledTimes(1)
  })

  it('结构结果取失败时保留三维区域与重试，不影响任务本身', async () => {
    fetchTaskMock.mockResolvedValue(taskSnapshot('succeeded'))
    fetchTaskResultMock
      .mockRejectedValueOnce(
        new ApiError({
          code: 'compute_unreachable',
          params: { url: 'http://127.0.0.1:8001' },
        }),
      )
      .mockResolvedValueOnce(rockSaltResult)

    const wrapper = await mountResultView()

    expect(wrapper.text()).toContain(t('structure_title'))
    expect(wrapper.text()).toContain(
      tCode('compute_unreachable', { url: 'http://127.0.0.1:8001' }),
    )
    expect(wrapper.text()).toContain(t('task_status_succeeded'))

    await wrapper.find('.structure-retry').trigger('click')
    await settle()

    expect(fetchTaskResultMock).toHaveBeenCalledTimes(2)
    expect(wrapper.find('structure-viewer-stub').exists()).toBe(true)
  })
})
