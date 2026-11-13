> English | [简体中文](./README.zh-CN.md)

# Data management layer · Database (backend/internal/store)

Database connection, GORM models, and automatic migrations. Table design (fields, constraints, state machine) is in [../../README.md](../../README.md) "Data tables".

| File | Purpose |
| --- | --- |
| `models.go` | `User` / `PredictionTask` / `StructureResult`, plus task-status and crystal-family values |
| `store.go` | Connection (switchable engine) and `Migrate` (create tables / add columns) |
| `store_test.go` | Migration, reads/writes and `Preload`, foreign-key behavior (cascade delete / set null), invalid configuration |

## Switching databases

| Variable | Default | Meaning |
| --- | --- | --- |
| `YCSP_ENV` | `development` | `production` rejects SQLite |
| `YCSP_DB_DRIVER` | `sqlite` | `sqlite` (development/test only) / `postgres` / `mysql` |
| `YCSP_DB_DSN` | `ycsp.db` (SQLite only) | SQLite uses a database file path; postgres / mysql use their own connection strings |

SQLite is the zero-install option for local development and tests. Production uses PostgreSQL.
postgres / mysql have no defaults: a connection string contains an address and credentials, and a wrong guess only makes diagnosis harder, so startup fails when the DSN is missing.

## Implementation details

- **Startup Ping**: `Open` pings the database with a 5-second timeout, so a wrong DSN or an unreachable server fails at startup instead of on the first request.
- **Connection pool**: SQLite is fixed at 1 open / 1 idle connection because it allows only one writer; Postgres / MySQL default to 20 open / 10 idle / 30-minute lifetime. `MaxOpenConns` / `MaxIdleConns` / `ConnMaxLifetime` can be overridden through `YCSP_DB_MAX_OPEN_CONNS` / `YCSP_DB_MAX_IDLE_CONNS` / `YCSP_DB_CONN_MAX_LIFETIME_SECONDS`.
- **SQLite foreign-key checks are off by default**: `Open` automatically appends `_pragma=foreign_keys(1)` to the connection string; otherwise `ON DELETE CASCADE` / `SET NULL` do not work (covered by tests). Postgres enforces foreign keys natively.
- **Migrations**: development/test auto-migrate at startup (`YCSP_DB_AUTO_MIGRATE=true`); production defaults to `false`, and `EnsureSchema` verifies that the three tables exist before the server starts. Run `go run . -migrate-only` once before deployment.

## Conventions

- Declare foreign keys only on the "child" side (`PredictionTask.User`, `StructureResult.Task`). Declaring them on both sides creates two constraints, one without cascade, and deletion is blocked by it.
- Passwords are stored only as bcrypt hashes (`users.password_hash`); hashes are created and verified in [`../auth`](../auth/README.md), and the model never contains plaintext.
- Security question answers are stored only as bcrypt hashes (`users.security_answer_hash`); the question itself is stored as text for the student to read.
- JSON columns (`error_params` / `atoms` / `rule` / `caveats`) are stored as text; their structure is defined by the data processing layer and is not described twice in the tables.

Tests: `go test ./internal/store/` uses temporary SQLite by default. Set `YCSP_TEST_DB_DRIVER=postgres` and `YCSP_TEST_DB_DSN=...` to run the same suite against PostgreSQL; each test creates its own schema and drops it afterwards.
