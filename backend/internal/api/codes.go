// 数据管理层的接口状态码。
//
// 面向用户的文案由前端按当前语言翻译（见根 README「多语言」），
// 所以后端只返回语言无关的 code 与 params，不返回写死某种语言的文字。
package api

const (
	CodeBackendHealthy     = "backend_healthy"      // 本层健康检查通过
	CodeComputeHealthy     = "compute_healthy"      // 已连通数据处理层
	CodeComputeUnreachable = "compute_unreachable"  // 联系不上数据处理层（参数 url）
	CodeComputeBadStatus   = "compute_bad_status"   // 数据处理层返回非 200（参数 status_code）
	CodeComputeReadFailed  = "compute_read_failed"  // 读取上游应答失败
	CodeComputeInvalidJSON = "compute_invalid_json" // 上游应答不是合法 JSON
)
