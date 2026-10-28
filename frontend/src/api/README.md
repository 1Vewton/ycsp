# 前端 · 接口（frontend/src/api）

与数据管理层的 HTTP 请求封装；页面只拿结构化数据，不自己拼地址。

| 文件 | 作用 |
| --- | --- |
| `tasks.ts` | 创建 / 查询 / 取消预测任务、读取结构结果；把 `{ status, code, params }` 统一转成 `ApiError` |

## 约定

- 只调数据管理层的 `/api/...`，不直接连数据处理层；开发时由 Vite 代理（见 [../../README.md](../../README.md)）。
- 失败一律抛 `ApiError`（`code` + `params`），页面按 code 从 [`../locales`](../locales/README.md) 取文案，请求层不写死句子。
- 网络不通、应答不是 JSON 时用 `request_failed` / `internal_error`；技术细节只打印到控制台，不往界面上抛。
