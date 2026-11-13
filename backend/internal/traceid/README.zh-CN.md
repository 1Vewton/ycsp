> [English](./README.md) | 简体中文

# 数据管理层 · 请求编号（backend/internal/traceid）

一次请求在前端、数据管理层、数据处理层使用同一个 `X-Request-ID`，用于把三层日志串起来。

| 文件 | 作用 |
| --- | --- |
| `traceid.go` | 生成随机编号、校验外部传入的编号、在 `context.Context` 中存取编号 |
| `traceid_test.go` | 生成、校验（拒绝换行与超长）、context 往返 |

## 规则

- 头名统一为 `X-Request-ID`，最长 64 个字符。
- 只接受字母、数字与 `-_.:`；换行等字符可能污染日志，一律拒绝并重新生成。
- 前端生成 `web-…`，数据管理层兜底生成 `api-…`，数据处理层兜底生成 `compute-…`。
- 异步任务会把创建任务时的编号存进 `prediction_tasks.trace_id`，worker 调用数据处理层时继续使用，
  所以任务从创建到计算完成只有一个编号。
