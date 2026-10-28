# 数据管理层 · 测试工具（backend/internal/testutil）

只给测试用的共享工具，不参与生产构建（只有 `*_test.go` 会引用它）。

| 文件 | 作用 |
| --- | --- |
| `database.go` | `OpenDatabase(t)`：临时目录里的 SQLite + 迁移好的三张表，`t.Cleanup` 自动关闭 |

为什么单独成包：`store` / `tasks` / `api` / `users` 的测试都需要“一个干净的库”，
写在各自文件里会重复四遍建库逻辑。
