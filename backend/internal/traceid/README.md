> English | [简体中文](./README.zh-CN.md)

# Data management layer · Request IDs (backend/internal/traceid)

One request uses the same `X-Request-ID` in the frontend, the data management layer, and the data processing layer, tying the three layers' logs together.

| File | Purpose |
| --- | --- |
| `traceid.go` | Generate random IDs, validate externally supplied IDs, store/read IDs in `context.Context` |
| `traceid_test.go` | Generation, validation (reject newlines and overlong values), context round-trip |

## Rules

- The header name is always `X-Request-ID`, at most 64 characters.
- Only letters, digits, and `-_.:` are accepted; newlines and similar characters could pollute logs, so such IDs are rejected and regenerated.
- The frontend generates `web-…`, the data management layer falls back to `api-…`, and the data processing layer falls back to `compute-…`.
- Asynchronous tasks store the ID from task creation in `prediction_tasks.trace_id`; the worker keeps using it when calling the data processing layer, so a task has one ID from creation to completion.
