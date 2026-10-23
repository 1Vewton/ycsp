# 数据管理层 · 数据处理层客户端（backend/internal/compute）

调用数据处理层 `POST /predict`，把结果交给任务服务落库。接口契约见
[`compute/README.md`](../../../compute/README.md)《对外接口》。

| 文件 | 作用 |
| --- | --- |
| `client.go` | 发请求、超时（10 秒）、解析统一应答 |
| `types.go` | `Prediction` / `Candidate` / `Cell`；`rule`、`caveats`、`atoms` 保留原始 JSON |
| `errors.go` | 四类失败 + `Describe`：把客户端错误翻成前端可用的 `code` 与 `params` |

## 错误怎么分

| 情况 | 客户端错误 | 落库的 code |
| --- | --- | --- |
| 上游 400 + `{status:error, code}`（如化学式电荷不平） | `PredictionError` | **原样透传**上游 code，前端已有文案 |
| 连不上 / 超时 | `TransportError` | `compute_unreachable`（参数 `url`） |
| 读不出应答 | `ProtocolError` | `compute_read_failed` |
| 状态码不是 200 | `ProtocolError` | `compute_bad_status`（参数 `status_code`） |
| 应答不是合法 JSON | `ProtocolError` | `compute_invalid_json` |
| 200 但没有候选结构 | `ProtocolError` | `compute_no_result` |

`rule` / `caveats` / `atoms` 用 `json.RawMessage` 保存：落库与回传时字节不变，
不需要在后端重新描述一遍结构。

测试：`go test ./internal/compute/`（用 `httptest` 造假的数据处理层，不需要真的起 Python）。
