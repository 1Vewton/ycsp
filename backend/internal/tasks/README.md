> English | [简体中文](./README.zh-CN.md)

# Data management layer · Task service (backend/internal/tasks)

Connects "create task → queue → call data processing layer → persist". The API layer handles HTTP only; queueing and execution live here.

| File | Purpose |
| --- | --- |
| `service.go` | `Create` / `Get` / `Result` / `Cancel` / `History`, the queue, requeueing leftover tasks at startup, and the periodic sweep of pending tasks |
| `worker.go` | Claim a task → call `Predictor` → write the result or error |

## Why asynchronous

The API returns `pending` first, and the frontend polls `GET /api/tasks/:id`. Although computation currently takes only about a hundred milliseconds, moving the wait to the background means the frontend will not need changes when a slower algorithm arrives, and it matches the rule that "long-running work must give feedback".

## How concurrency avoids conflicts

- **Claim with a conditional update**: `WHERE id = ? AND status = 'pending'`; if zero rows are affected, skip. If "cancel" and "start computing" happen at the same time, a cancelled task will not run again.
- **Result and status in one transaction**: writing `structure_results` and setting the status to `succeeded` happen in the same transaction, so there is never "a result exists but the status is still running".
- **Requeue at startup**: tasks left in `running` by the previous process go back to `pending`, and `pending` tasks are requeued; a restart loses no task (tasks beyond the first batch follow on the next sweep).
- **Periodic sweep**: every second the service reads a batch of `pending` tasks and re-enqueues them; an in-memory set keeps each task in the queue at most once, so a task dropped by a full queue is retried without a restart.

## How failures are recorded

`compute.Describe` turns data processing errors into `code` + `params` and stores them in the task (e.g. `compute_unreachable`, `compute_timeout`, `formula_charge_unbalanced`); the frontend translates by code.
A compute request that exceeds the client timeout is stored as `compute_timeout`; the worker records the failure and moves on to the next task.

The request ID from task creation is stored in `prediction_tasks.trace_id`; the worker keeps using it when calling the data processing layer and logs `tasks: start / succeeded / failed trace_id=… task_id=…`, so one ID ties the three layers' logs together.

## History

`History(userID, limit, offset)` returns the signed-in account's tasks newest first, together with the result summary of succeeded tasks.
Anonymous tasks have no owner and never appear. `limit` defaults to 20 and is capped at 50; `offset` pages through older records.

## Worker count

`NewWithWorkers(database, predictor, workers)` starts `workers` goroutines (1–16); `New` keeps the default of 1 for local development and tests.
The database layer does not decide this: `main` chooses the default by driver (SQLite 1, Postgres / MySQL 4) and lets `YCSP_TASK_WORKERS` override it.
Multiple workers are safe because claiming uses the conditional update above; on SQLite the writes still serialize, so raising the count mainly overlaps time spent waiting on the data processing layer.

## Known limitations

- Queue capacity is 64 (`QueueCapacity`); when full, tasks stay `pending` and the sweep retries them about once per second, so no restart is needed.
- Only `pending` can be cancelled; cancelling a running task returns `task_cancel_conflict`.
- `requeueUnfinished` resets `running` tasks at startup. That is correct for a single backend process; multiple backend instances need a lease / owner column or an external queue.

Tests: `go test ./internal/tasks/` (fake Predictor + temporary SQLite by default; set `YCSP_TEST_DB_DRIVER=postgres` and `YCSP_TEST_DB_DSN=...` to run against PostgreSQL).
