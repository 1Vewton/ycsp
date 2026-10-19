# 数据处理层 · 环境配置（compute/app/env）

统一读取环境变量，未设置时返回默认值（`.env` 位于 `compute/` 根目录，可缺省）。

| 名称 | 作用 |
| --- | --- |
| `get_env(key, fallback)` | 字符串配置 |
| `get_env_int(key, fallback)` | 整数配置（如端口） |
| `SERVICE_HOST` / `SERVICE_PORT` | 本层监听地址与端口（已解析好的值） |
