// Package codes 是数据管理层的全部状态码。
//
// 面向用户的文案由前端按当前语言翻译（见根 README「多语言」），
// 所以这里只放语言无关的 code；要填进文案的数值放在 params 里。
package codes

const (
	// 应答状态：成功
	BackendHealthy = "backend_healthy"
	ComputeHealthy = "compute_healthy"
	TaskCreated    = "task_created"
	TaskOK         = "task_ok"
	TaskResult     = "task_result"
	TaskCancelled  = "task_cancelled"
)

// 转发数据处理层时的失败。
const (
	ComputeUnreachable = "compute_unreachable" // 参数 url
	ComputeBadStatus   = "compute_bad_status"  // 参数 status_code
	ComputeReadFailed  = "compute_read_failed"
	ComputeInvalidJSON = "compute_invalid_json"
	ComputeNoResult    = "compute_no_result"
)

// 任务接口的错误。
const (
	TaskNotFound       = "task_not_found"       // 参数 task_id
	TaskNotReady       = "task_not_ready"       // 参数 status
	TaskCancelConflict = "task_cancel_conflict" // 参数 status
	TaskFormulaEmpty   = "task_formula_empty"
	TaskFormulaTooLong = "task_formula_too_long" // 参数 max_length
)

// 通用错误。
const (
	RequestInvalidBody = "request_invalid_body"
	InternalError      = "internal_error"
)
