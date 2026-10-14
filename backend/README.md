# 数据管理层（backend）

前端与数据处理层之间唯一的通道：对外提供统一接口，对内把计算请求转发给数据处理层，并管理业务数据（Go + Gin；GORM 待阶段 2 接入）。

## 对外接口

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| GET | `/api/health` | 本层健康检查 |
| GET | `/api/ping-compute` | 转发到数据处理层 `/health` |

## 配置

| 环境变量 | 默认值 |
| --- | --- |
| `YCSP_BACKEND_PORT` | `8000` |
| `YCSP_COMPUTE_BASE_URL` | `http://127.0.0.1:8001` |

## 如何运行

```powershell
go mod tidy
go run .
```

启动后访问 <http://127.0.0.1:8000/api/health>。
