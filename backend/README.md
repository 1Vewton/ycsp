# 数据管理层（backend）

前端与数据处理层之间唯一的通道：对外提供统一接口，对内把计算请求转发给数据处理层，并管理业务数据（Go + Gin；GORM 待阶段 2 接入）。

## 对外接口

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| GET | `/api/health` | 本层健康检查 |
| GET | `/api/ping-compute` | 转发到数据处理层 `/health` |

应答统一为 `{ service, status, code, params? }`，成功时另有 `compute`。`code` 是语言无关的状态码（见 `internal/api/codes.go`），面向用户的文案由前端按当前语言翻译。

## 配置

配置来自 `backend/.env`（可选：不提供就用默认值；要改就把 `.env.example` 复制为 `.env`）：

| 变量 | 默认值 | 含义 |
| --- | --- | --- |
| `YCSP_BACKEND_HOST` | `127.0.0.1` | 本层监听地址 |
| `YCSP_BACKEND_PORT` | `8000` | 本层监听端口 |
| `YCSP_COMPUTE_BASE_URL` | `http://127.0.0.1:8001` | 数据处理层地址 |

## 如何运行

```powershell
go mod tidy
go run .
```

启动后访问 <http://127.0.0.1:8000/api/health>。
