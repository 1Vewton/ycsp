> [English](./README.md) | 简体中文

# 数据管理层 · 接口层（backend/internal/api）

数据管理层全部的 HTTP 接口。

| 文件 | 作用 |
| --- | --- |
| `router.go` | 注册路由，装配依赖（`Options`） |
| `auth.go` | 账号接口：注册、登录、退出、`GET /api/users/me`、修改密码、设置安全问题、用安全问题答案重置密码，以及会话 Cookie 的下发与清理 |
| `health.go` | `GET /api/health` |
| `compute_proxy.go` | `GET /api/ping-compute`，转发到数据处理层（超时 5 秒，透传 `X-Request-ID`） |
| `tasks.go` | 任务接口：创建、查询状态、取结果、取消、历史记录，以及任务的应答形状 |
| `request_log.go` | 请求编号与访问日志：沿用/生成 `X-Request-ID`，记录方法、路径、状态码与耗时 |
| `client_address.go` | 登录限流取客户端地址：默认只认直连地址；配置可信代理后才解析 `X-Forwarded-For` |
| `errors.go` | 统一错误应答 `writeError` 与 panic 兜底 |

状态码统一放在 [../codes/codes.go](../codes/codes.go)（语言无关，文案由前端按语言翻译）。
