# 数据管理层（backend）

前端与数据处理层之间唯一的通道：对外提供统一接口，对内把计算请求转发给数据处理层，并管理业务数据（Go + Gin + GORM）。

## 对外接口

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| GET | `/api/health` | 本层健康检查 |
| GET | `/api/ping-compute` | 转发到数据处理层 `/health` |
| POST | `/api/users` | 学生自助注册（成功后直接登录） |
| POST | `/api/sessions` | 登录 |
| DELETE | `/api/sessions` | 退出登录（幂等） |
| GET | `/api/users/me` | 当前登录用户（未登录返回 401） |
| POST | `/api/tasks` | 创建预测任务，立刻返回 `pending` |
| GET | `/api/tasks/:id` | 查询任务状态（前端轮询这个） |
| GET | `/api/tasks/:id/result` | 取任务的结构结果（还没算完返回 409） |
| POST | `/api/tasks/:id/cancel` | 取消排队中的任务 |

应答统一为 `{ service, status, code, params? }`，成功时另有 `compute` / `task` / `result` / `user`。
`code` 是语言无关的状态码（见 `internal/codes/codes.go`），面向用户的文案由前端按当前语言翻译。

任务流程（异步）：`POST /api/tasks` 落库为 `pending` 并返回 → 后台 worker 调数据处理层 →
前端轮询 `GET /api/tasks/:id` 直到 `succeeded` / `failed` → 再取 `/result`。
详细规则见 [`internal/tasks/README.md`](./internal/tasks/README.md)。

## 账号与会话

- **学生自助注册**：昵称 2~32 个字符且唯一（昵称就是登录名），密码至少 8 位并含英文字母、
  数字、特殊符号各一个；密码只存 bcrypt 哈希（见 [`internal/auth/README.md`](./internal/auth/README.md)）。
- **会话**：登录成功后生成 256 位随机编号，存 Redis（`ycsp:session:<编号>` → 用户编号），
  Cookie 名 `ycsp_session`，`HttpOnly` + `SameSite=Lax`，默认 7 天；HTTPS 部署时把
  `YCSP_COOKIE_SECURE` 设为 `true`。会话规则见 [`internal/session/README.md`](./internal/session/README.md)。
- **失败不泄露账号是否存在**：昵称不存在与密码错误都返回 `invalid_credentials`。
- **登录失败限流**：同一昵称 15 分钟内失败 5 次、同一地址失败 20 次即锁定 15 分钟，
  期间返回 `429 too_many_attempts`（带 `retry_after_seconds` / `retry_after_minutes` 与标准 `Retry-After` 头）；
  登录成功只清昵称计数。规则见 [`internal/ratelimit/README.md`](./internal/ratelimit/README.md)。
- **不登录也能预测**：任务照常创建，只是 `user_id` 为空；登录后创建的任务会记下归属。

## 数据表

三张表：使用者、预测任务、结构结果。任务与结果分表——失败或取消的任务没有结果，
而状态更新又很频繁，塞进同一行只会放大写冲突。

```text
users 1 ──── n prediction_tasks 1 ──── 1 structure_results
      （user_id 可空：不登录也能预测）        （task_id 唯一）
```

### `users`：使用者

| 字段 | 逻辑类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `id` | 无符号整数 | 主键，自增 | |
| `display_name` | 字符串(64) | 非空，唯一 | 学生昵称，同时是登录名 |
| `password_hash` | 字符串(100) | 非空，默认空串 | bcrypt 哈希：随机盐写在哈希串里，**从不存明文**；空串表示还没设密码，这种账号不能登录（见 [`internal/auth/README.md`](./internal/auth/README.md)） |
| `created_at` | 时间 | 非空 | UTC |

### `prediction_tasks`：一次预测任务

| 字段 | 逻辑类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `id` | 无符号整数 | 主键，自增 | |
| `user_id` | 无符号整数 | 可空，索引，外键 → `users.id`（删用户时置空） | 未登录时为 NULL |
| `formula` | 字符串(64) | 非空 | 用户输入的化学式原文（已去首尾空格） |
| `status` | 字符串(16) | 非空，联合索引 `(status, created_at)` | `pending` / `running` / `succeeded` / `failed` / `cancelled` |
| `error_code` | 字符串(64) | 可空 | 失败时的语言无关错误码，与数据处理层的 code 同名 |
| `error_params` | JSON | 可空 | 错误参数，如 `{"formula":"NaCl2"}` |
| `created_at` | 时间 | 非空 | 任务创建（UTC） |
| `started_at` | 时间 | 可空 | 开始计算 |
| `finished_at` | 时间 | 可空 | 结束（成功 / 失败 / 取消） |
| `duration_ms` | 整数 | 可空 | 计算耗时，用来告诉学生“这次算了多久” |

状态机（只有 `pending` 允许直接取消）：

```text
pending ──► running ──► succeeded
   │            ├─────► failed
   └────────────┴─────► cancelled
```

### `structure_results`：任务产出的结构

| 字段 | 逻辑类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `id` | 无符号整数 | 主键，自增 | |
| `task_id` | 无符号整数 | 非空，唯一索引，外键 → `prediction_tasks.id`（级联删除） | 一个任务一行 |
| `family` | 字符串(16) | 非空 | `ionic` / `covalent` / `metallic` |
| `prototype` | 字符串(32) | 非空 | `rock_salt` 等原型编号 |
| `coordination_number` | 整数 | 非空 | 配位数 |
| `edge_a_angstrom`、`edge_b_angstrom`、`edge_c_angstrom` | 浮点 | 非空 | 三条棱长（Å） |
| `angle_alpha_degree`、`angle_beta_degree`、`angle_gamma_degree` | 浮点 | 非空 | 三个夹角（°） |
| `atoms` | JSON | 非空 | `[{"symbol":"Na","fractional_coordinates":[0,0,0]}, …]` |
| `rule` | JSON | 非空 | `{"code":"rule_radius_ratio","params":{…}}` |
| `caveats` | JSON | 非空（可为 `[]`） | `[{"code":"caveat_near_boundary","params":{…}}]` |
| `created_at` | 时间 | 非空 | UTC |

### 为什么这样设计

- **晶胞参数拆成列，原子与说明存 JSON**：晶胞参数要参与校验与展示，拆列便于直接查询；
  原子列表和“为什么是这个结构”的 code 由数据处理层定义，数据库不重复描述其结构。
- **`user_id` 可空**：阶段 2 还没有登录，不登录也能预测；账号体系定了再补。
- **一个任务一行结果**：当前候选结构恒为 1；以后要多候选，就加 `candidate_index` 并去掉唯一索引。
- **删用户不删任务**（`user_id` 置空）：教学数据保留；删任务则级联删结果。

### 与接口的对应

| 接口 | 落库动作 |
| --- | --- |
| `POST /api/tasks` | 插入 `prediction_tasks`（`pending`） |
| `GET /api/tasks/:id` | 读 `prediction_tasks.status` 与错误字段 |
| `GET /api/tasks/:id/result` | 读 `structure_results` |
| `POST /api/tasks/:id/cancel` | 把 `pending` 改成 `cancelled` |

> 数据库引擎（G0.2）已定：连接层可切换（`YCSP_DB_DRIVER` / `YCSP_DB_DSN`），当前默认 SQLite。
> 上表是逻辑类型，具体列类型由 GORM 按驱动映射；JSON 先按文本存（PostgreSQL 以后可换 `JSONB`）。
> 模型与迁移见 [`internal/store/`](./internal/store/README.md)，启动时自动建表。

## 配置

配置来自 `backend/.env`（可选：不提供就用默认值；要改就把 `.env.example` 复制为 `.env`）：

| 变量 | 默认值 | 含义 |
| --- | --- | --- |
| `YCSP_BACKEND_HOST` | `127.0.0.1` | 本层监听地址 |
| `YCSP_BACKEND_PORT` | `8000` | 本层监听端口 |
| `YCSP_COMPUTE_BASE_URL` | `http://127.0.0.1:8001` | 数据处理层地址 |
| `YCSP_DB_DRIVER` | `sqlite` | 数据库引擎：`sqlite` / `postgres` / `mysql` |
| `YCSP_DB_DSN` | `ycsp.db`（仅 sqlite） | sqlite 是文件路径；postgres / mysql 必须显式给连接串 |
| `YCSP_SESSION_DRIVER` | `redis` | 会话存储：`redis` / `memory`（后者仅单机开发：重启即掉登录） |
| `YCSP_REDIS_URL` | `redis://127.0.0.1:6379/0` | Redis 地址；启动时连不上会直接失败并提示 |
| `YCSP_SESSION_LIFETIME_SECONDS` | `604800`（7 天） | 会话有效期 |
| `YCSP_COOKIE_SECURE` | `false` | 只在 HTTPS 部署时设为 `true` |
| `YCSP_LOGIN_MAX_ATTEMPTS` | `5` | 同一昵称在窗口内允许的失败次数 |
| `YCSP_LOGIN_IP_MAX_ATTEMPTS` | `20` | 同一地址在窗口内允许的失败次数 |
| `YCSP_LOGIN_BLOCK_SECONDS` | `900`（15 分钟） | 失败计数的窗口长度，也是锁定时的等待时间 |

## 如何运行

```powershell
go mod tidy
go run .
```

启动后访问 <http://127.0.0.1:8000/api/health>。

- 首次启动会自动创建 SQLite 文件（默认 `backend/ycsp.db`）并建表；想回到空库，删掉这个文件即可。
- 注册与登录需要 Redis（默认 `redis://127.0.0.1:6379/0`）；本地没有 Redis 时设 `YCSP_SESSION_DRIVER=memory`。
- 任务接口要先启动数据处理层：联系不上时任务会以 `compute_unreachable` 失败（不会一直卡在 pending）。
- 拉不到 Go 依赖时（`proxy.golang.org` 不通）：先 `go env -w GOPROXY=https://goproxy.cn,direct`，再 `go mod tidy`。
- 测试：`go test ./...`；加 `-cover` 看覆盖率（当前 `env` 100%、`store` 92.9%、`auth` 91.4%、`session` 84.1%、
  `api` 81.6%，任务与账号包在 75%~85%；只有启动入口 `main()` 不测，因为它会真的占用端口）。
