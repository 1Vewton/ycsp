// 任务接口的前端客户端：只负责发请求、把应答换成结构化数据。
//
// 应答形状见 backend/README.md《对外接口》：{ service, status, code, params?, task? }。
// 统一信封与错误转换在 ./client.ts；页面用 code 从语言包取文案——请求层不写死任何一句话。

import { ApiError, problemFromError, requestApi, type ApiProblem } from './client'

// 兼容原来的引用方式：页面与测试从任务模块拿 ApiError / TaskProblem 也能工作。
export { ApiError, problemFromError }
export type { ApiProblem }
export type TaskProblem = ApiProblem

/** 任务的五种状态，与后端 prediction_tasks.status 一致。 */
export type TaskStatus = 'pending' | 'running' | 'succeeded' | 'failed' | 'cancelled'

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

/** 任务成功后的结构结果；rule 与 caveats 留给结果解释使用。 */
export interface StructureResult {
  family: string
  prototype: string
  coordination_number: number
  cell: StructureCell
  atoms: StructureAtom[]
  rule: TaskProblem
  caveats: TaskProblem[]
}

const TASKS_PATH = '/api/tasks'

interface TaskEnvelope {
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
  const envelope = await requestApi<TaskEnvelope>(url, init)
  if (envelope.task === undefined) {
    throw new ApiError({ code: 'internal_error' })
  }
  return envelope.task
}

async function requestResult(url: string, init: RequestInit): Promise<StructureResult> {
  const envelope = await requestApi<TaskEnvelope>(url, init)
  if (envelope.result === undefined) {
    throw new ApiError({ code: 'internal_error' })
  }
  return envelope.result
}
