> [English](./README.md) | 简体中文

# 数据管理层 · 内部实现（backend/internal）

不对外暴露的实现代码（Go 的 `internal/` 约定：只能被本模块导入）。

| 子目录 | 作用 |
| --- | --- |
| `api/` | HTTP 接口层，见 [api/README.zh-CN.md](./api/README.zh-CN.md) |
| `auth/` | 密码哈希、密码策略与校验（bcrypt），见 [auth/README.zh-CN.md](./auth/README.zh-CN.md) |
| `codes/` | 状态码（语言无关，文案由前端翻译），见 [codes/README.zh-CN.md](./codes/README.zh-CN.md) |
| `compute/` | 数据处理层客户端，见 [compute/README.zh-CN.md](./compute/README.zh-CN.md) |
| `ratelimit/` | 登录失败限流（按昵称 / 地址），见 [ratelimit/README.zh-CN.md](./ratelimit/README.zh-CN.md) |
| `session/` | 登录会话存储（Redis / 内存），见 [session/README.zh-CN.md](./session/README.zh-CN.md) |
| `store/` | 数据库连接、GORM 模型与迁移，见 [store/README.zh-CN.md](./store/README.zh-CN.md) |
| `tasks/` | 任务排队与执行（异步），见 [tasks/README.zh-CN.md](./tasks/README.zh-CN.md) |
| `traceid/` | 请求编号的生成、校验与传递（三层日志追踪），见 [traceid/README.zh-CN.md](./traceid/README.zh-CN.md) |
| `users/` | 账号注册与登录校验，见 [users/README.zh-CN.md](./users/README.zh-CN.md) |
| `env/` | 环境变量读取（带默认值），见 [env/README.zh-CN.md](./env/README.zh-CN.md) |
| `testutil/` | 只给测试用的共享工具（临时数据库），见 [testutil/README.zh-CN.md](./testutil/README.zh-CN.md) |
