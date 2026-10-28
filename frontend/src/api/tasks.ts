// 任务接口的前端客户端：只负责发请求、把应答换成结构化数据。
//
// 应答形状见 backend/README.md《对外接口》：{ service, status, code, params?, task? }。
// 失败时抛出 ApiError，页面用 code 从语言包取文案——请求层不写死任何一句话。

/** 任务的五种状态，与后端 prediction_tasks.status 一致。 */
export type TaskStatus = 'pending' | 'running' | 'succeeded' | 'failed' | 'cancelled'

/** 一条可翻译的问题：code 决定说哪句话，params 是要填进句子的值。 */
export interface TaskProblem {
  code: string
  params?: Record<string, unknown>
}

/** 任务快照；字段名保持后端 JSON 的写法，少一层转换。 */
export interface TaskSnapshot {
  id: number
  formula: string
  status: TaskStatus
  created_at: string
  started_at: string | null
  finished_at: string | null
  duration_ms: number | null
  error?: TaskProblem
}

/** 后端拒绝请求或网络失败时抛出的错误。 */
export class ApiError extends Error {
  readonly problem: TaskProblem

  constructor(problem: TaskProblem) {
    super(problem.code)
    this.name = 'ApiError'
    this.problem = problem
  }
}

/** 把任意异常换成可翻译的问题；技术细节只留在控制台。 */
export function problemFromError(error: unknown): TaskProblem {
  if (error instanceof ApiError) {
    return error.problem
  }
  console.error(error)
  return { code: 'request_failed' }
}

const TASKS_PATH = '/api/tasks'

interface TaskEnvelope {
  status?: string
  code?: string
  params?: Record<string, unknown>
  task?: TaskSnapshot
}

/** 创建预测任务；成功时返回刚建好的任务（状态为 pending）。 */
export function createTask(formula: string): Promise<TaskSnapshot> {
  return requestTask(TASKS_PATH, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ formula }),
  })
}

/** 查询任务状态。 */
export function fetchTask(taskId: number): Promise<TaskSnapshot> {
  return requestTask(`${TASKS_PATH}/${taskId}`, { method: 'GET' })
}

/** 取消排队中的任务；已经开始或结束的任务会以 task_cancel_conflict 失败。 */
export function cancelTask(taskId: number): Promise<TaskSnapshot> {
  return requestTask(`${TASKS_PATH}/${taskId}/cancel`, { method: 'POST' })
}

async function requestTask(url: string, init: RequestInit): Promise<TaskSnapshot> {
  let response: Response
  try {
    response = await fetch(url, init)
  } catch (error) {
    console.error(error)
    throw new ApiError({ code: 'request_failed' })
  }

  const envelope = await readEnvelope(response)
  if (!response.ok || envelope.status !== 'ok') {
    throw new ApiError({
      code: envelope.code ?? 'request_failed',
      params: envelope.params,
    })
  }
  if (envelope.task === undefined) {
    throw new ApiError({ code: 'internal_error' })
  }
  return envelope.task
}

// 应答不是 JSON 时（例如代理返回了一页 HTML），也给出可翻译的错误，
// 而不是让页面抛出一个学生看不懂的语法异常。
async function readEnvelope(response: Response): Promise<TaskEnvelope> {
  try {
    return (await response.json()) as TaskEnvelope
  } catch (error) {
    console.error(error)
    return { status: 'error', code: 'internal_error' }
  }
}
