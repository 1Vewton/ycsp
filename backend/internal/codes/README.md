# 数据管理层 · 状态码（backend/internal/codes）

数据管理层返回给前端的全部 code。**只有 code 与 params，没有任何语言的文案**：
文案在前端按当前语言翻译（见根 README「多语言」）。

| 分组 | 举例 | 说明 |
| --- | --- | --- |
| 成功 | `backend_healthy`、`task_created`、`task_result` | 表示这次请求成功了 |
| 转发失败 | `compute_unreachable`、`compute_bad_status` | 数据处理层没响应或应答异常 |
| 任务错误 | `task_not_found`、`task_not_ready`、`task_cancel_conflict` | 任务编号不存在 / 结果没算好 / 状态不允许取消 |
| 输入错误 | `task_formula_empty`、`task_formula_too_long` | 化学式为空或过长 |
| 通用 | `request_invalid_body`、`internal_error` | 请求体不合法、未预期异常 |

约定：

- 新增 code 必须同时在前端语言包 `frontend/src/messages.ts` 补中、英两种文案。
- 需要填进句子的数值放进 params（例如 `{"max_length": 64}`），不要把数字写死在文案里。
- 数据处理层自己返回的 code（如 `formula_charge_unbalanced`）原样透传，前端已有对应文案。
