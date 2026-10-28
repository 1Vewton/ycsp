# 数据管理层 · 内部实现（backend/internal）

不对外暴露的实现代码（Go 的 `internal/` 约定：只能被本模块导入）。

| 子目录 | 作用 |
| --- | --- |
| `api/` | HTTP 接口层，见 [api/README.md](./api/README.md) |
| `auth/` | 密码哈希、密码策略与校验（bcrypt），见 [auth/README.md](./auth/README.md) |
| `codes/` | 状态码（语言无关，文案由前端翻译），见 [codes/README.md](./codes/README.md) |
| `compute/` | 数据处理层客户端，见 [compute/README.md](./compute/README.md) |
| `ratelimit/` | 登录失败限流（按昵称 / 地址），见 [ratelimit/README.md](./ratelimit/README.md) |
| `session/` | 登录会话存储（Redis / 内存），见 [session/README.md](./session/README.md) |
| `store/` | 数据库连接、GORM 模型与迁移，见 [store/README.md](./store/README.md) |
| `tasks/` | 任务排队与执行（异步），见 [tasks/README.md](./tasks/README.md) |
| `users/` | 账号注册与登录校验，见 [users/README.md](./users/README.md) |
| `env/` | 环境变量读取（带默认值），见 [env/README.md](./env/README.md) |
| `testutil/` | 只给测试用的共享工具（临时数据库），见 [testutil/README.md](./testutil/README.md) |
