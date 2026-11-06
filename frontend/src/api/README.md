# 前端 · 接口（frontend/src/api）

与数据管理层的 HTTP 请求封装；页面只拿结构化数据，不自己拼地址。

| 文件 | 作用 |
| --- | --- |
| `client.ts` | 统一应答信封与失败转换（`ApiError` / `problemFromError`）；任务与账号接口共用 |
| `requestId.ts` | 生成 `X-Request-ID`（`web-…`）并统一头名与长度上限 |
| `tasks.ts` | 创建 / 查询 / 取消预测任务、读取结构结果 |
| `auth.ts` | 注册、登录、退出、取当前账号；未登录时 `fetchCurrentUser()` 返回 `null` |

## 约定

- 只调数据管理层的 `/api/...`，不直接连数据处理层；开发时由 Vite 代理（见 [../../README.md](../../README.md)）。
- 每次请求带 `X-Request-ID`：数据管理层与数据处理层原样沿用，控制台以
  `[ycsp] request ... request_id=...` 记录；失败时 `ApiError.requestId` 保留同一个编号。
- 失败一律抛 `ApiError`（`code` + `params`），页面按 code 从 [`../locales`](../locales/README.md) 取文案，请求层不写死句子。
- 网络不通、应答不是 JSON 时用 `request_failed` / `internal_error`；技术细节只打印到控制台，不往界面上抛。
- 会话 Cookie 由后端以 `HttpOnly` 下发，前端不读也不存；同源请求由浏览器自动携带。
