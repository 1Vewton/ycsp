> [English](./README.md) | 简体中文

# 数据管理层 · 数据库（backend/internal/store）

数据库连接、GORM 模型与自动迁移。表结构的设计（字段、约束、状态机）见 [../../README.zh-CN.md](../../README.zh-CN.md)《数据表》。

| 文件 | 作用 |
| --- | --- |
| `models.go` | `User` / `PredictionTask` / `StructureResult`，以及任务状态与晶体类型的取值 |
| `store.go` | 连接（引擎可切换）与 `Migrate`（建表 / 补列） |
| `store_test.go` | 迁移、存取与 `Preload`、外键行为（级联删除 / 置空）、错误配置 |

## 切换数据库

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `YCSP_ENV` | `development` | `production` 会拒绝 SQLite |
| `YCSP_DB_DRIVER` | `sqlite` | `sqlite`（仅开发/测试）/ `postgres` / `mysql` |
| `YCSP_DB_DSN` | `ycsp.db`（仅 sqlite） | sqlite 是数据库文件路径；postgres / mysql 填各自连接串 |

SQLite 是本地开发和测试的零安装选项；生产使用 PostgreSQL。
postgres / mysql 没有默认值：连接串含地址与账号，猜错只会更难排查，所以缺 DSN 时启动直接报错。

## 实现细节

- **启动 Ping**：`Open` 会用 5 秒超时 Ping 数据库，DSN 写错或服务不可达时在启动阶段就失败，而不是等第一次请求。
- **连接池**：SQLite 固定 1 个打开 / 1 个空闲连接（它只允许一个写入者）；Postgres / MySQL 默认 20 个打开、10 个空闲、30 分钟上限。可用 `YCSP_DB_MAX_OPEN_CONNS` / `YCSP_DB_MAX_IDLE_CONNS` / `YCSP_DB_CONN_MAX_LIFETIME_SECONDS` 覆盖。
- **SQLite 的外键检查默认是关的**：`Open` 会自动给连接串补 `_pragma=foreign_keys(1)`，
  否则 `ON DELETE CASCADE` / `SET NULL` 都不生效（有测试守着）。Postgres 原生强制外键。
- **迁移**：开发/测试启动时自动建表（`YCSP_DB_AUTO_MIGRATE=true`）；生产默认 `false`，启动前用 `EnsureSchema` 校验三张表都在，缺表会直接失败并提示先跑 `go run . -migrate-only`。

## 约定

- 外键只在“子表”一侧声明（`PredictionTask.User`、`StructureResult.Task`）。
  父子两侧都写会生成两条约束，其中一条不带级联，删除时会被它挡住。
- 密码只存 bcrypt 哈希（`users.password_hash`），哈希在 [`../auth`](../auth/README.zh-CN.md) 里生成与校验，模型里不出现明文。
- 安全问题答案只存 bcrypt 哈希（`users.security_answer_hash`）；问题本身按文本存，供学生本人查看。
- JSON 列（`error_params` / `atoms` / `rule` / `caveats`）用文本存，
  具体结构由数据处理层定义，表里不重复描述。

测试：`go test ./internal/store/` 默认用临时 SQLite。设置 `YCSP_TEST_DB_DRIVER=postgres` 和 `YCSP_TEST_DB_DSN=...` 后，同一套测试会跑 Postgres；每个测试建独立 schema 并在结束时删除。
