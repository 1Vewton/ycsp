# 数据管理层 · 数据库（backend/internal/store）

数据库连接、GORM 模型与自动迁移。表结构的设计（字段、约束、状态机）见 [../../README.md](../../README.md)《数据表》。

| 文件 | 作用 |
| --- | --- |
| `models.go` | `User` / `PredictionTask` / `StructureResult`，以及任务状态与晶体类型的取值 |
| `store.go` | 连接（引擎可切换）与 `Migrate`（建表 / 补列） |
| `store_test.go` | 迁移、存取与 `Preload`、外键行为（级联删除 / 置空）、错误配置 |

## 切换数据库

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `YCSP_DB_DRIVER` | `sqlite` | `sqlite` / `postgres` / `mysql` |
| `YCSP_DB_DSN` | `ycsp.db`（仅 sqlite） | sqlite 是数据库文件路径；postgres / mysql 填各自连接串 |

postgres / mysql 没有默认值：连接串含地址与账号，猜错只会更难排查，所以缺 DSN 时启动直接报错。

## 两个实现细节

- **SQLite 的外键检查默认是关的**：`Open` 会自动给连接串补 `_pragma=foreign_keys(1)`，
  否则 `ON DELETE CASCADE` / `SET NULL` 都不生效（有测试守着）。
- **SQLite 连接数限制为 1**：它同一时刻只允许一个写入者，连接开多了会频繁报 `database is locked`。
  教学场景并发很低，这个限制够用；换成 PostgreSQL 时只有 sqlite 分支受限。

## 约定

- 外键只在“子表”一侧声明（`PredictionTask.User`、`StructureResult.Task`）。
  父子两侧都写会生成两条约束，其中一条不带级联，删除时会被它挡住。
- JSON 列（`error_params` / `atoms` / `rule` / `caveats`）用文本存，
  具体结构由数据处理层定义，表里不重复描述。

测试：`go test ./internal/store/`
