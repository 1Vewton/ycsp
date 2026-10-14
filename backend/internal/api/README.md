# 数据管理层 · 接口层（backend/internal/api）

数据管理层全部的 HTTP 接口。

| 文件 | 作用 |
| --- | --- |
| `router.go` | 注册路由 |
| `health.go` | `GET /api/health` |
| `compute_proxy.go` | `GET /api/ping-compute`，转发到数据处理层（超时 5 秒） |
