# 数据处理层 · 应用包（compute/app）

用 `python -m app.main` 启动（见上级 [compute/README.md](../README.md)）。

| 路径 | 作用 |
| --- | --- |
| `main.py` | 入口：组装 FastAPI 应用并启动 uvicorn |
| `api/` | HTTP 接口、序列化与统一错误应答，见 [api/README.md](./api/README.md) |
| `domain/` | 晶体学数据类型与预测规则，见 [domain/README.md](./domain/README.md) |
| `constants/` | 元素数据、结构原型数据、判据阈值与状态码，见 [constants/README.md](./constants/README.md) |
| `env/` | 环境配置（带默认值），见 [env/README.md](./env/README.md) |
| `__init__.py` | 包元信息：服务名、版本 |
