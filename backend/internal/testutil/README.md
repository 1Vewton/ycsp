> English | [简体中文](./README.zh-CN.md)

# Data management layer · Test helpers (backend/internal/testutil)

Shared helpers used only by tests; they are not part of the production build (only `*_test.go` files import them).

| File | Purpose |
| --- | --- |
| `database.go` | `OpenDatabase(t)`: a migrated temporary database, closed automatically by `t.Cleanup` |

Default: a SQLite file in a temporary directory, so tests need no extra service.

PostgreSQL: set `YCSP_TEST_DB_DRIVER=postgres` and `YCSP_TEST_DB_DSN=...`. Each test then creates its own schema, appends `search_path` to the DSN, migrates it, and drops it in `t.Cleanup`. This keeps `go test ./...` package-level parallelism isolated without truncating shared tables.

Why a separate package: the tests for `store` / `tasks` / `api` / `users` all need "a clean database"; writing that logic in each file would duplicate database setup four times.
