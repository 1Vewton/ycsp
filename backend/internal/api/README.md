> English | [简体中文](./README.zh-CN.md)

# Data management layer · HTTP API (backend/internal/api)

All HTTP endpoints of the data management layer.

| File | Purpose |
| --- | --- |
| `router.go` | Register routes and assemble dependencies (`Options`) |
| `auth.go` | Account endpoints: register, sign in, sign out, `GET /api/users/me`, password change, security question setup, password reset by security answer, and session cookie issue/clear |
| `health.go` | `GET /api/health` |
| `compute_proxy.go` | `GET /api/ping-compute`, forwarded to the data processing layer (5-second timeout, passes `X-Request-ID` through) |
| `tasks.go` | Task endpoints: create, status, result, cancel, history, and task response shapes |
| `request_log.go` | Request IDs and access logs: reuse/generate `X-Request-ID`, log method, path, status, and duration |
| `client_address.go` | Client address for login rate limiting: direct peer by default; parses `X-Forwarded-For` only when trusted proxies are configured |
| `errors.go` | Unified error response `writeError` and panic recovery |

Status codes live in [../codes/codes.go](../codes/codes.go) (language-neutral; the frontend translates the copy).
