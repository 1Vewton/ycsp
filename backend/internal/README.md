> English | [简体中文](./README.zh-CN.md)

# Data management layer · internal packages (backend/internal)

Implementation code that is not exposed publicly (Go's `internal/` rule: only this module can import it).

| Package | Purpose |
| --- | --- |
| `api/` | HTTP interface layer, see [api/README.md](./api/README.md) |
| `auth/` | Password hashing, password policy, and validation (bcrypt), see [auth/README.md](./auth/README.md) |
| `codes/` | Status codes (language-neutral; copy is translated by the frontend), see [codes/README.md](./codes/README.md) |
| `compute/` | Data processing layer client, see [compute/README.md](./compute/README.md) |
| `ratelimit/` | Login failure rate limiting (per display name / address), see [ratelimit/README.md](./ratelimit/README.md) |
| `session/` | Login session storage (Redis / memory), see [session/README.md](./session/README.md) |
| `store/` | Database connection, GORM models, and migrations, see [store/README.md](./store/README.md) |
| `tasks/` | Task queueing and execution (asynchronous), see [tasks/README.md](./tasks/README.md) |
| `traceid/` | Request ID generation, validation, and propagation (three-layer log tracing), see [traceid/README.md](./traceid/README.md) |
| `users/` | Account registration and login validation, see [users/README.md](./users/README.md) |
| `env/` | Environment variable reading (with defaults), see [env/README.md](./env/README.md) |
| `testutil/` | Shared test-only helpers (temporary database), see [testutil/README.md](./testutil/README.md) |
