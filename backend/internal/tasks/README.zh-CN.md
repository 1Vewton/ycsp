> [English](./README.md) | 简体中文

# 数据管理层 · 任务服务（backend/internal/tasks）

把“创建任务 → 排队 → 调数据处理层 → 落库”串起来。接口层只管 HTTP，排队与执行都在这里。

| 文件 | 作用 |
| --- | --- |
| `service.go` | `Create` / `Get` / `Result` / `Cancel` / `History`、队列、启动时补算残留任务、周期性补扫 pending 任务 |
| `worker.go` | 抢占任务 → 调 `Predictor` → 写结果或错误 |

## 为什么是异步

接口先返回 `pending`，前端轮询 `GET /api/tasks/:id`。计算现在虽然只要一百多毫秒，
但把等待放到后台后，以后换成更慢的算法前端不用改，也符合“长耗时要有反馈”的约定。

## 并发怎么保证不打架

- **抢占用条件更新**：`WHERE id = ? AND status = 'pending'`，影响行数为 0 就跳过。
  “取消”与“开始计算”同时发生时，已取消的任务不会被又跑起来。
- **结果与状态同事务**：写 `structure_results` 与把状态改成 `succeeded` 在同一个事务里，
  不会出现“有结果但状态还是 running”。
- **启动时补算**：上次进程退出时停在 `running` 的任务退回 `pending`，`pending` 的重新排队（超出一批的由下一次补扫继续），重启不丢任务。
- **周期性补扫**：每秒从数据库取一批 `pending` 任务重新入队；内存里记录已入队编号，保证同一任务最多排队一次，队列满时被丢掉的任务不需要重启就会重试。

## 失败怎么记

`compute.Describe` 把数据处理层的错误翻成 `code` + `params` 存进任务
（如 `compute_unreachable`、`compute_timeout`、`formula_charge_unbalanced`），前端按 code 翻译。
计算请求超过客户端超时会落成 `compute_timeout`；worker 记下失败后继续处理下一条，不会一直卡住。

任务创建时的请求编号会存进 `prediction_tasks.trace_id`；worker 调用数据处理层时继续使用它，
并记录 `tasks: start / succeeded / failed trace_id=… task_id=…`，按同一个编号能把三层日志串起来。

## 历史记录

`History(userID, limit, offset)` 按账号取预测记录，最新在前，并带上成功任务的结构摘要。
匿名任务没有归属，永远不会出现在这里。`limit` 默认 20、最多 50，`offset` 用于翻页。

## Worker 数量

`NewWithWorkers(database, predictor, workers)` 会启动 `workers` 个 goroutine（1~16）；`New` 保留默认值 1，给本地开发和测试用。
数据库层不决定这个数：`main` 按驱动选默认值（SQLite 1，Postgres / MySQL 4），并允许 `YCSP_TASK_WORKERS` 覆盖。
多 worker 是安全的，因为抢占仍然用上面的条件更新；SQLite 的写入仍会串行，调大 worker 主要是在等待数据处理层时重叠计算。

## 已知限制

- 队列容量 64（`QueueCapacity`）；满了任务留在 `pending`，由补扫约每秒重试一次，不再需要重启。
- 只有 `pending` 能取消；已经开始的任务取消会返回 `task_cancel_conflict`。
- `requeueUnfinished` 会在启动时把所有 `running` 任务退回 `pending`。这对单个 backend 进程是正确的；多个 backend 实例需要租约 / owner 字段或外部队列。

测试：`go test ./internal/tasks/`（假 Predictor + 默认临时 SQLite；设置 `YCSP_TEST_DB_DRIVER=postgres` 和 `YCSP_TEST_DB_DSN=...` 可跑 PostgreSQL）。
