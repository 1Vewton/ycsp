> [English](./README.md) | 简体中文

# 数据管理层 · 测试工具（backend/internal/testutil）

只给测试用的共享工具，不参与生产构建（只有 `*_test.go` 会引用它）。

| 文件 | 作用 |
| --- | --- |
| `database.go` | `OpenDatabase(t)`：迁移好的临时数据库，`t.Cleanup` 自动关闭 |

默认用临时目录里的 SQLite，测试不需要额外服务。

Postgres：设置 `YCSP_TEST_DB_DRIVER=postgres` 和 `YCSP_TEST_DB_DSN=...`。
每个测试会建独立 schema、把 `search_path` 追加到 DSN、迁移后返回连接，并在 `t.Cleanup` 里删除 schema。
这样 `go test ./...` 的包级并行不会互相污染，也不需要清空共享表。

为什么单独成包：`store` / `tasks` / `api` / `users` 的测试都需要“一个干净的库”，
写在各自文件里会重复四遍建库逻辑。
