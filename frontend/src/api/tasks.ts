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

/** 结构结果里的一个原子：元素符号 + 分数坐标（相对棱长的比例，无量纲）。 */
export interface StructureAtom {
  symbol: string
  fractional_coordinates: [number, number, number]
}

/** 晶胞的 6 个参数：棱长单位 Å，夹角单位 °。 */
export interface StructureCell {
  edge_a_angstrom: number
  edge_b_angstrom: number
  edge_c_angstrom: number
  angle_alpha_degree: number
  angle_beta_degree: number
  angle_gamma_degree: number
}

/** 任务成功后的结构结果；rule 与 caveats 留给结果解释使用（3.5）。 */
export interface StructureResult {
  family: string
  prototype: string
  coordination_number: number
  cell: StructureCell
  atoms: StructureAtom[]
  rule: TaskProblem
  caveats: TaskProblem[]
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
  result?: StructureResult
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

/** 取任务的结构结果；任务还没算完时后端返回 409（task_not_ready）。 */
export function fetchTaskResult(taskId: number): Promise<StructureResult> {
  return requestResult(`${TASKS_PATH}/${taskId}/result`, { method: 'GET' })
}

async function requestTask(url: string, init: RequestInit): Promise<TaskSnapshot> {
  const envelope = await requestEnvelope(url, init)
  if (envelope.task === undefined) {
    throw new ApiError({ code: 'internal_error' })
  }
  return envelope.task
}

async function requestResult(url: string, init: RequestInit): Promise<StructureResult> {
  const envelope = await requestEnvelope(url, init)
  if (envelope.result === undefined) {
    throw new ApiError({ code: 'internal_error' })
  }
  return envelope.result
}

async function requestEnvelope(url: string, init: RequestInit): Promise<TaskEnvelope> {
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
  return envelope
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
