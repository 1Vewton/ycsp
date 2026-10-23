# 数据管理层 · 任务服务（backend/internal/tasks）

把“创建任务 → 排队 → 调数据处理层 → 落库”串起来。接口层只管 HTTP，排队与执行都在这里。

| 文件 | 作用 |
| --- | --- |
| `service.go` | `Create` / `Get` / `Result` / `Cancel`、队列、启动时补算残留任务 |
| `worker.go` | 抢占任务 → 调 `Predictor` → 写结果或错误 |

## 为什么是异步

接口先返回 `pending`，前端轮询 `GET /api/tasks/:id`。计算现在虽然只要一百多毫秒，
但把等待放到后台后，以后换成更慢的算法前端不用改，也符合“长耗时要有反馈”的约定。

## 并发怎么保证不打架

- **抢占用条件更新**：`WHERE id = ? AND status = 'pending'`，影响行数为 0 就跳过。
  “取消”与“开始计算”同时发生时，已取消的任务不会被又跑起来。
- **结果与状态同事务**：写 `structure_results` 与把状态改成 `succeeded` 在同一个事务里，
  不会出现“有结果但状态还是 running”。
- **启动时补算**：上次进程退出时停在 `running` 的任务退回 `pending`，`pending` 的统一重新排队，重启不丢任务。

## 失败怎么记

`compute.Describe` 把数据处理层的错误翻成 `code` + `params` 存进任务
（如 `compute_unreachable`、`formula_charge_unbalanced`），前端按 code 翻译。

## 已知限制

- 队列容量 64；满了任务留在 `pending`，等重启补算（教学场景够用）。
- 同一时刻 1 个 worker：SQLite 只允许一个写入者。
- 只有 `pending` 能取消；已经开始的任务取消会返回 `task_cancel_conflict`。

测试：`go test ./internal/tasks/`（假 Predictor + 临时 SQLite，不依赖 Python）。
