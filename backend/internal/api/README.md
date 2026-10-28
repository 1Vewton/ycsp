# 数据管理层 · 接口层（backend/internal/api）

数据管理层全部的 HTTP 接口。

| 文件 | 作用 |
| --- | --- |
| `router.go` | 注册路由，装配依赖（`Options`） |
| `auth.go` | 账号接口：注册、登录、退出、`GET /api/users/me`，以及会话 Cookie 的下发与清理 |
| `health.go` | `GET /api/health` |
| `compute_proxy.go` | `GET /api/ping-compute`，转发到数据处理层（超时 5 秒） |
| `tasks.go` | 任务接口：创建、查询状态、取结果、取消，以及任务的应答形状 |
| `errors.go` | 统一错误应答 `writeError` 与 panic 兜底 |

状态码统一放在 [../codes/codes.go](../codes/codes.go)（语言无关，文案由前端按语言翻译）。
