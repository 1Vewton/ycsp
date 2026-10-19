# 数据管理层 · 环境变量（backend/internal/env）

统一读取环境变量，未设置时返回默认值。

| 函数 | 作用 |
| --- | --- |
| `Load(path)` | 读取 `.env`；读不到只提示，不中断 |
| `Get(key, fallback)` | 字符串配置 |
| `GetInt(key, fallback)` | 整数配置（如端口） |
