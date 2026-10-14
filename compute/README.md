# 数据处理层（compute）

按化学组成预测晶体结构，并以 HTTP 接口提供给数据管理层调用（Python + FastAPI + NumPy）。

## 对外接口

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| GET | `/health` | 健康检查（实现见 [app/README.md](./app/README.md)） |

预测接口 `POST /predict` 待阶段 1，见 [TASKS.md](../TASKS.md)。

## 如何运行

依赖用 [uv](https://docs.astral.sh/uv/) 管理：版本已锁在 `uv.lock` 里，首次运行会自动建好 `.venv`。

```powershell
uv run python -m app.main
```

启动后访问 <http://127.0.0.1:8001/health>。
