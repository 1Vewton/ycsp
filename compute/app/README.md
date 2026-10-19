# 数据处理层 · 应用包（compute/app）

FastAPI 应用与接口，用 `python -m app.main` 启动（见上级 [compute/README.md](../README.md)）。

| 文件 | 作用 |
| --- | --- |
| `main.py` | FastAPI 应用与接口（阶段 0 只有 `GET /health`） |
| `env/` | 环境配置（带默认值），见 [env/README.md](./env/README.md) |
| `structure.py` | 内部数据结构：元素 / 原子 / 晶胞（长度用 Å，角度用 °） |
| `__init__.py` | 标记 `app` 为 Python 包 |
