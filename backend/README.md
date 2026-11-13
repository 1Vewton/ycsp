> English | [简体中文](./README.zh-CN.md)

# Data management layer (backend)

The only channel between the frontend and the data processing layer: it exposes a single public API, forwards compute requests inward, and manages business data (Go + Gin + GORM).

## Public API

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/api/health` | Health check for this layer |
| GET | `/api/ping-compute` | Forward to the data processing layer's `/health` |
| POST | `/api/users` | Student self-registration (signs in on success) |
| POST | `/api/sessions` | Sign in |
| DELETE | `/api/sessions` | Sign out (idempotent) |
| GET | `/api/users/me` | Current signed-in user (401 when signed out) |
| PUT | `/api/users/me/password` | Change the password (requires the current password) |
| PUT | `/api/users/me/security-question` | Set or update the security question (requires the current password) |
| POST | `/api/password-resets/question` | Look up the security question by nickname |
| POST | `/api/password-resets` | Reset the password with the security answer (signs in on success) |
| POST | `/api/tasks` | Create a prediction task, returns `pending` immediately |
| GET | `/api/tasks/:id` | Read task status (the frontend polls this) |
| GET | `/api/tasks/:id/result` | Read the task's structure result (409 while still running) |
| POST | `/api/tasks/:id/cancel` | Cancel a queued task |
| GET | `/api/users/me/tasks?limit=20&offset=0` | Prediction history of the current account (newest first) |

Responses use `{ service, status, code, params? }`; on success they also carry `compute` / `task` / `result` / `user`.
`code` is a language-neutral status code (see `internal/codes/codes.go`); user-facing copy is translated by the frontend into the current language.

Task flow (asynchronous): `POST /api/tasks` stores the task as `pending` and returns → a background worker calls the data processing layer →
the frontend polls `GET /api/tasks/:id` until `succeeded` / `failed` → then fetches `/result`.
Detailed rules: [`internal/tasks/README.md`](./internal/tasks/README.md).

## Accounts and sessions

- **Student self-registration**: display name 2–32 characters and unique (the display name is the login name); password at least 8 characters with at least one letter, one digit, and one symbol; only a bcrypt hash is stored (see [`internal/auth/README.md`](./internal/auth/README.md)).
- **Sessions**: after sign-in a 256-bit random ID is generated and stored in Redis (`ycsp:session:<id>` → user ID). The cookie is named `ycsp_session`, `HttpOnly` + `SameSite=Lax`, valid for 7 days by default; set `YCSP_COOKIE_SECURE` to `true` for HTTPS. Session rules: [`internal/session/README.md`](./internal/session/README.md).
- **Failures do not reveal whether an account exists**: an unknown display name and a wrong password both return `invalid_credentials`.
- **Login failure rate limiting**: 5 failures per display name within 15 minutes, or 20 failures per address, locks the account/address for 15 minutes and returns `429 too_many_attempts` (with `retry_after_seconds` / `retry_after_minutes` and the standard `Retry-After` header); a successful sign-in clears only the display-name counter. Rules: [`internal/ratelimit/README.md`](./internal/ratelimit/README.md).
- **Prediction works without signing in**: tasks are created normally with an empty `user_id`; tasks created while signed in record their owner.
- **Password change**: `PUT /api/users/me/password` needs the current password, applies the same password policy and refuses a password identical to the old one. The current session stays signed in.
- **Security question reset**: registration can optionally set a question and answer; signed-in users can set or update them (current password required). The answer is normalized (trimmed, lowercase) and stored only as a bcrypt hash. `POST /api/password-resets` verifies the answer, sets the new password and signs the student in. An unknown account, an account without a question and a wrong answer all return `password_reset_failed` to avoid revealing whether an account exists; reset attempts share the login failure rate limit.
- **Prediction history**: `GET /api/users/me/tasks` lists only the current account's tasks, newest first; anonymous tasks have no owner and are excluded. `limit` defaults to 20 and is capped at 50; `offset` pages through older records.
- **Session limitation**: changing or resetting a password does not invalidate sessions that already exist on other devices; sign out there after using a shared computer.

## Data tables

Three tables: users, prediction tasks, and structure results. Tasks and results are separate tables: failed or cancelled tasks have no result, while status updates are frequent, so putting everything in one row only increases write conflicts.

```text
users 1 ──── n prediction_tasks 1 ──── 1 structure_results
      (user_id nullable: prediction works without signing in)  (task_id unique)
```

### `users`: platform users

| Field | Logical type | Constraint | Description |
| --- | --- | --- | --- |
| `id` | unsigned integer | primary key, auto-increment | |
| `display_name` | string(64) | not null, unique | Student display name, also the login name |
| `password_hash` | string(100) | not null, default empty string | bcrypt hash: the random salt is embedded in the hash string, **plaintext is never stored**; an empty string means no password has been set, and such an account cannot sign in (see [`internal/auth/README.md`](./internal/auth/README.md)) |
| `security_question` | string(128) | not null, default empty string | Question shown when the student resets the password; empty means self-service reset is not available |
| `security_answer_hash` | string(100) | not null, default empty string | bcrypt hash of the normalized answer; the plaintext answer is never stored |
| `created_at` | time | not null | UTC |

### `prediction_tasks`: one prediction task

| Field | Logical type | Constraint | Description |
| --- | --- | --- | --- |
| `id` | unsigned integer | primary key, auto-increment | |
| `user_id` | unsigned integer | nullable, indexed, foreign key → `users.id` (set to NULL when the user is deleted) | NULL when not signed in |
| `formula` | string(64) | not null | The formula exactly as entered (leading/trailing spaces removed) |
| `trace_id` | string(64) | nullable, indexed | Request ID that created the task; the worker reuses it when calling the data processing layer (see "Log tracing") |
| `status` | string(16) | not null, composite index `(status, created_at)` | `pending` / `running` / `succeeded` / `failed` / `cancelled` |
| `error_code` | string(64) | nullable | Language-neutral error code on failure, same name as in the data processing layer |
| `error_params` | JSON | nullable | Error parameters, e.g. `{"formula":"NaCl2"}` |
| `created_at` | time | not null | Task creation (UTC) |
| `started_at` | time | nullable | Computation started |
| `finished_at` | time | nullable | Finished (succeeded / failed / cancelled) |
| `duration_ms` | integer | nullable | Computation duration, used to tell students "how long this took" |

State machine (only `pending` can be cancelled directly):

```text
pending ──► running ──► succeeded
   │            ├─────► failed
   └────────────┴─────► cancelled
```

### `structure_results`: structure produced by a task

| Field | Logical type | Constraint | Description |
| --- | --- | --- | --- |
| `id` | unsigned integer | primary key, auto-increment | |
| `task_id` | unsigned integer | not null, unique index, foreign key → `prediction_tasks.id` (cascade delete) | one row per task |
| `family` | string(16) | not null | `ionic` / `covalent` / `metallic` |
| `prototype` | string(32) | not null | prototype ID such as `rock_salt` |
| `coordination_number` | integer | not null | coordination number |
| `edge_a_angstrom`, `edge_b_angstrom`, `edge_c_angstrom` | float | not null | three edge lengths (Å) |
| `angle_alpha_degree`, `angle_beta_degree`, `angle_gamma_degree` | float | not null | three angles (°) |
| `atoms` | JSON | not null | `[{"symbol":"Na","fractional_coordinates":[0,0,0]}, …]` |
| `rule` | JSON | not null | `{"code":"rule_radius_ratio","params":{…}}` |
| `caveats` | JSON | not null (may be `[]`) | `[{"code":"caveat_near_boundary","params":{…}}]` |
| `created_at` | time | not null | UTC |

### Why this design

- **Unit-cell parameters as columns, atoms and explanations as JSON**: unit-cell parameters participate in validation and display, so columns make direct queries easy;
  the atom list and the "why this structure" codes are defined by the data processing layer, and the database does not describe their structure twice.
- **`user_id` nullable**: prediction works without signing in, so the task can exist without an owner.
- **One result row per task**: the current candidate count is always 1; if multiple candidates are needed later, add `candidate_index` and drop the unique index.
- **Deleting a user keeps their tasks** (`user_id` set to NULL): teaching data is preserved; deleting a task cascades to its result.

### Mapping to the API

| API | Database action |
| --- | --- |
| `POST /api/tasks` | insert a `prediction_tasks` row (`pending`) |
| `GET /api/tasks/:id` | read `prediction_tasks.status` and error fields |
| `GET /api/tasks/:id/result` | read `structure_results` |
| `POST /api/tasks/:id/cancel` | change `pending` to `cancelled` |

> The database engine (G0.2) is settled: the connection layer is switchable (`YCSP_DB_DRIVER` / `YCSP_DB_DSN`).
> **SQLite is for local development and tests only; production uses PostgreSQL** (`YCSP_ENV=production` rejects SQLite).
> The table shows logical types; GORM maps them to concrete column types per driver; JSON is stored as text for now (PostgreSQL can move to `JSONB` later).
> Models and migrations: [`internal/store/`](./internal/store/README.md); development/test auto-migrate at startup, production uses `-migrate-only` as a separate step.

## Configuration

Configuration comes from `backend/.env` (optional: defaults are used when absent; to change them, copy `.env.example` to `.env`):

| Variable | Default | Meaning |
| --- | --- | --- |
| `YCSP_ENV` | `development` | Runtime environment: `development` / `test` / `production`. `production` rejects SQLite |
| `YCSP_BACKEND_HOST` | `127.0.0.1` | Listen address for this layer |
| `YCSP_BACKEND_PORT` | `8000` | Listen port for this layer |
| `YCSP_COMPUTE_BASE_URL` | `http://127.0.0.1:8001` | Data processing layer address |
| `YCSP_DB_DRIVER` | `sqlite` | Database engine: `sqlite` (development/test only) / `postgres` / `mysql` |
| `YCSP_DB_DSN` | `ycsp.db` (SQLite only) | SQLite uses a file path; postgres / mysql require an explicit connection string |
| `YCSP_DB_MAX_OPEN_CONNS` | `0` | Maximum open connections; 0 uses the driver default (SQLite is fixed at 1) |
| `YCSP_DB_MAX_IDLE_CONNS` | `0` | Maximum idle connections; 0 uses the driver default (SQLite is fixed at 1) |
| `YCSP_DB_CONN_MAX_LIFETIME_SECONDS` | `0` | Maximum connection lifetime in seconds; 0 uses the driver default (server databases: 30 minutes) |
| `YCSP_DB_AUTO_MIGRATE` | `true` for development/test, `false` for production | Whether to create/update tables at startup |
| `YCSP_TASK_WORKERS` | `0` | Task workers; 0 chooses by driver (SQLite 1, Postgres/MySQL 4), range 1–16 |
| `YCSP_SESSION_DRIVER` | `redis` | Session store: `redis` / `memory` (the latter is for single-machine development only: a restart signs everyone out) |
| `YCSP_REDIS_URL` | `redis://127.0.0.1:6379/0` | Redis address; startup fails with a hint when Redis is unreachable |
| `YCSP_SESSION_LIFETIME_SECONDS` | `604800` (7 days) | Session lifetime |
| `YCSP_COOKIE_SECURE` | `false` | Set to `true` only for HTTPS deployments |
| `YCSP_LOGIN_MAX_ATTEMPTS` | `5` | Failed attempts allowed per display name within the window |
| `YCSP_LOGIN_IP_MAX_ATTEMPTS` | `20` | Failed attempts allowed per address within the window |
| `YCSP_LOGIN_BLOCK_SECONDS` | `900` (15 minutes) | Failure-counting window, also the lockout wait time |
| `YCSP_TRUSTED_PROXIES` | empty | Trusted reverse-proxy IPs / CIDRs (comma-separated). `X-Forwarded-For` is read only when the direct peer matches; for nginx on the same machine use `127.0.0.1/32`. `0.0.0.0/0` is forbidden |

### Production database (PostgreSQL)

SQLite is for local development and tests only: it is fixed at one connection and serializes writes, which is not suitable for production. A production configuration looks like:

```dotenv
YCSP_ENV=production
YCSP_DB_DRIVER=postgres
YCSP_DB_DSN=host=db.internal port=5432 user=ycsp password=*** dbname=ycsp sslmode=require
YCSP_DB_AUTO_MIGRATE=false
YCSP_TASK_WORKERS=4
```

Deployment order:

1. Run `go run . -migrate-only` once to create/update tables (this flag ignores `YCSP_DB_AUTO_MIGRATE`).
2. Start `go run .` normally; production does not auto-migrate at startup by default. If a table is missing, startup fails and tells you to run `-migrate-only`.
3. Keep the pool defaults (20 open / 10 idle / 30-minute lifetime) until you have load-test data, then tune
   `YCSP_DB_MAX_OPEN_CONNS` / `YCSP_DB_MAX_IDLE_CONNS` / `YCSP_DB_CONN_MAX_LIFETIME_SECONDS`.
4. `YCSP_TASK_WORKERS` defaults to 1 for SQLite and 4 for Postgres / MySQL; the range is 1–16.

With `YCSP_ENV=production`, the service refuses to start if `YCSP_DB_DRIVER=sqlite`.
Local development and tests keep using SQLite: zero install, no extra service.

### Reverse proxy deployment (nginx)

Without configuration, the backend rate-limits by **direct peer address**: all students behind nginx share nginx's IP,
so one student failing repeatedly can lock the whole class out. Add nginx's address to `YCSP_TRUSTED_PROXIES`
and the backend will read the real client address from `X-Forwarded-For`.

When nginx and the backend are on the same machine:

```nginx
server {
    listen 443 ssl;
    server_name crystal.example.com;

    root /srv/ycsp/frontend/dist;
    index index.html;

    location /api/ {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        # Important: append the direct client address to X-Forwarded-For
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

```dotenv
# backend/.env
YCSP_TRUSTED_PROXIES=127.0.0.1/32
YCSP_COOKIE_SECURE=true
```

- **List only your own proxy addresses / ranges**: on the same machine use `127.0.0.1/32`; when nginx runs in a container use the container network range
  (e.g. `172.18.0.0/16`) or nginx's fixed IP. `0.0.0.0/0` is rejected at startup —
  that would let any client forge an address and bypass rate limiting.
- **Take the first untrusted address from right to left**: values a client stuffs into `X-Forwarded-For` sit on the left and are ignored;
  nginx appends the real address on the right, which is used.
- **nginx must append or overwrite, never pass through**: use `$proxy_add_x_forwarded_for` (append) or
  `$remote_addr` (overwrite). Using `$http_x_forwarded_for` lets the client fully control the header.
- **Do not expose the backend port to the public internet**: let only nginx reach 8000. Requests that hit the backend directly are ignored
  even if they carry a forged header, because the direct peer is not a trusted proxy.
- **Container deployment**: the backend must listen on `0.0.0.0` (`YCSP_BACKEND_HOST=0.0.0.0`) to be reachable from an nginx container,
  but do not map it to a public host port; put nginx and the backend on the same Docker network and list that network range as trusted.
- **When a CDN / load balancer sits in front**: add their addresses to `YCSP_TRUSTED_PROXIES` as well;
  or restore the real IP with nginx's `real_ip` module first, then pass only the restored address to the backend with
  `proxy_set_header X-Forwarded-For $remote_addr;` (the backend then trusts only nginx).
- If the forwarded header is malformed, fall back to the direct peer address (safety first); during that time the proxy IP's counter is shared, so check the nginx configuration.
- `X-Request-ID` is an ordinary request header that nginx forwards by default; the backend also generates one when missing, so no extra configuration is needed.

If you want **nginx access logs searchable by the same ID**, define a map in the `http` block
(reuse the frontend ID when present, otherwise use nginx's own) and pass it to the backend in `location /api/`:

```nginx
# http block
map $http_x_request_id $ycsp_request_id {
    default $http_x_request_id;
    ""      $request_id;
}
log_format ycsp '$remote_addr - $request_method $uri $status request_id=$ycsp_request_id';
access_log /var/log/nginx/ycsp.access.log ycsp;

# add inside location /api/
proxy_set_header X-Request-ID $ycsp_request_id;
```

## Log tracing

All three layers use `X-Request-ID` to tie one request together:

| Layer | ID source | Log example |
| --- | --- | --- |
| Frontend | Generates `web-…` for every request and puts it in the header | `[ycsp] request start request_id=web-… method=POST path=/api/tasks` |
| Data management | Reuses the frontend ID; generates `api-…` when missing or invalid and writes it back to the response header | `http: request_id=web-… client_ip=203.0.113.7 method=POST path=/api/tasks status=201 duration_ms=8` |
| Data processing | Reuses the data management ID; generates `compute-…` when missing or invalid and writes it back to the response header | `request_id=web-… request method=POST path=/predict status=200 duration_ms=3` |

- **Asynchronous prediction**: the ID from task creation is stored in `prediction_tasks.trace_id`; the worker keeps using it when calling the data processing layer
  and logs `tasks: start / succeeded / failed trace_id=… task_id=…`.
- **How to search**: copy the `request_id` from the browser console (or the `trace_id` from task logs) and search for the same value in all three layers' logs.
- **What is logged**: only method, path, status code, duration, and prediction summary; request bodies, passwords, and cookies are never logged.
- **Log-injection protection**: IDs are at most 64 characters and may contain only letters, digits, and `-_.:`; IDs containing newlines or other characters are replaced.

## How to run

```powershell
go mod tidy
go run .
```

Then open <http://127.0.0.1:8000/api/health>.

- Local development uses SQLite by default: the first start creates `backend/ycsp.db` and its tables; delete the file to return to an empty database.
- Production uses PostgreSQL: set `YCSP_ENV=production` + `YCSP_DB_DRIVER=postgres` + `YCSP_DB_DSN`, run `go run . -migrate-only` once, then `go run .` (see "Production database (PostgreSQL)").
- Registration and login need Redis (default `redis://127.0.0.1:6379/0`); without a local Redis set `YCSP_SESSION_DRIVER=memory`.
- Task endpoints require the data processing layer: when it is unreachable or times out, tasks fail with `compute_unreachable` / `compute_timeout` instead of staying `pending` forever.
- When Go dependencies cannot be fetched (`proxy.golang.org` is unreachable): run `go env -w GOPROXY=https://goproxy.cn,direct` first, then `go mod tidy`.
- Behind nginx, configure `YCSP_TRUSTED_PROXIES`; otherwise login rate limiting counts nginx's IP (see \"Reverse proxy deployment (nginx)\").
- Tests: `go test ./...`; add `-cover` for coverage and `-race` for the concurrent task creation / timeout paths (currently `env` 100%, `store` 92.9%, `auth` 91.4%, `session` 84.1%,
  `api` 84.0%, `traceid` 90%; the tasks and users packages are 74%–85%. Only the startup entry `main()` is not tested, because it would really bind the port).
- To verify PostgreSQL locally, prepare a test database, set
  `YCSP_TEST_DB_DRIVER=postgres` and `YCSP_TEST_DB_DSN=...`, then run `go test ./...`;
  each test creates its own schema and drops it afterwards. The backend CI job already runs the suite once against PostgreSQL.
