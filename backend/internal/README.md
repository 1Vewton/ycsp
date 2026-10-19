# 数据管理层 · 内部实现（backend/internal）

不对外暴露的实现代码（Go 的 `internal/` 约定：只能被本模块导入）。

| 子目录 | 作用 |
| --- | --- |
| `api/` | HTTP 接口层，见 [api/README.md](./api/README.md) |
| `env/` | 环境变量读取（带默认值），见 [env/README.md](./env/README.md) |
