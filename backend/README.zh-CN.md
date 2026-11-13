> [English](./README.md) | 简体中文

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
| PUT | `/api/users/me/password` | 修改密码（需要当前密码） |
| PUT | `/api/users/me/security-question` | 设置或更新安全问题（需要当前密码） |
| POST | `/api/password-resets/question` | 按昵称取安全问题 |
| POST | `/api/password-resets` | 用安全问题答案重置密码（成功后直接登录） |
| POST | `/api/tasks` | 创建预测任务，立刻返回 `pending` |
| GET | `/api/tasks/:id` | 查询任务状态（前端轮询这个） |
| GET | `/api/tasks/:id/result` | 取任务的结构结果（还没算完返回 409） |
| POST | `/api/tasks/:id/cancel` | 取消排队中的任务 |
| GET | `/api/users/me/tasks?limit=20&offset=0` | 当前账号的预测记录（最新在前） |

应答统一为 `{ service, status, code, params? }`，成功时另有 `compute` / `task` / `result` / `user`。
`code` 是语言无关的状态码（见 `internal/codes/codes.go`），面向用户的文案由前端按当前语言翻译。

任务流程（异步）：`POST /api/tasks` 落库为 `pending` 并返回 → 后台 worker 调数据处理层 →
前端轮询 `GET /api/tasks/:id` 直到 `succeeded` / `failed` → 再取 `/result`。
详细规则见 [`internal/tasks/README.zh-CN.md`](./internal/tasks/README.zh-CN.md)。

## 账号与会话

- **学生自助注册**：昵称 2~32 个字符且唯一（昵称就是登录名），密码至少 8 位并含英文字母、
  数字、特殊符号各一个；密码只存 bcrypt 哈希（见 [`internal/auth/README.zh-CN.md`](./internal/auth/README.zh-CN.md)）。
- **会话**：登录成功后生成 256 位随机编号，存 Redis（`ycsp:session:<编号>` → 用户编号），
  Cookie 名 `ycsp_session`，`HttpOnly` + `SameSite=Lax`，默认 7 天；HTTPS 部署时把
  `YCSP_COOKIE_SECURE` 设为 `true`。会话规则见 [`internal/session/README.zh-CN.md`](./internal/session/README.zh-CN.md)。
- **失败不泄露账号是否存在**：昵称不存在与密码错误都返回 `invalid_credentials`。
- **登录失败限流**：同一昵称 15 分钟内失败 5 次、同一地址失败 20 次即锁定 15 分钟，
  期间返回 `429 too_many_attempts`（带 `retry_after_seconds` / `retry_after_minutes` 与标准 `Retry-After` 头）；
  登录成功只清昵称计数。规则见 [`internal/ratelimit/README.zh-CN.md`](./internal/ratelimit/README.zh-CN.md)。
- **不登录也能预测**：任务照常创建，只是 `user_id` 为空；登录后创建的任务会记下归属。
- **修改密码**：`PUT /api/users/me/password` 需要当前密码，沿用注册时的密码策略，并拒绝与旧密码相同的新密码；当前会话保持登录。
- **安全问题重置**：注册时可以选填问题与答案；登录后也可以设置或更新（需要当前密码）。答案会去掉首尾空格并转小写，只存 bcrypt 哈希。`POST /api/password-resets` 校验答案后设置新密码并直接登录。账号不存在、没设安全问题、答案不对都统一返回 `password_reset_failed`，避免泄露账号是否存在；重置失败与登录失败共用限流。
- **预测记录**：`GET /api/users/me/tasks` 只列当前账号的任务，最新在前；匿名任务没有归属，不会出现。`limit` 默认 20、最多 50，`offset` 用于翻页。
- **会话限制**：修改或重置密码不会让其他设备上已经存在的会话失效；在公用电脑上用完请退出登录。

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
| `password_hash` | 字符串(100) | 非空，默认空串 | bcrypt 哈希：随机盐写在哈希串里，**从不存明文**；空串表示还没设密码，这种账号不能登录（见 [`internal/auth/README.zh-CN.md`](./internal/auth/README.zh-CN.md)） |
| `security_question` | 字符串(128) | 非空，默认空串 | 忘记密码时显示的安全问题；空串表示没有启用自助重置 |
| `security_answer_hash` | 字符串(100) | 非空，默认空串 | 规范化后的答案的 bcrypt 哈希；从不存明文答案 |
| `created_at` | 时间 | 非空 | UTC |

### `prediction_tasks`：一次预测任务

| 字段 | 逻辑类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `id` | 无符号整数 | 主键，自增 | |
| `user_id` | 无符号整数 | 可空，索引，外键 → `users.id`（删用户时置空） | 未登录时为 NULL |
| `formula` | 字符串(64) | 非空 | 用户输入的化学式原文（已去首尾空格） |
| `trace_id` | 字符串(64) | 可空，索引 | 创建任务的请求编号；worker 调用数据处理层时沿用（见《日志追踪》） |
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

> 数据库引擎（G0.2）已定：连接层可切换（`YCSP_DB_DRIVER` / `YCSP_DB_DSN`）。
> **SQLite 只用于本地开发和测试；生产使用 PostgreSQL**（`YCSP_ENV=production` 会拒绝 SQLite）。
> 上表是逻辑类型，具体列类型由 GORM 按驱动映射；JSON 先按文本存（PostgreSQL 以后可换 `JSONB`）。
> 模型与迁移见 [`internal/store/`](./internal/store/README.zh-CN.md)；开发/测试默认启动时自动建表，生产用 `-migrate-only` 单独迁移。

## 配置

配置来自 `backend/.env`（可选：不提供就用默认值；要改就把 `.env.example` 复制为 `.env`）：

| 变量 | 默认值 | 含义 |
| --- | --- | --- |
| `YCSP_ENV` | `development` | 运行环境：`development` / `test` / `production`。`production` 禁止 SQLite |
| `YCSP_BACKEND_HOST` | `127.0.0.1` | 本层监听地址 |
| `YCSP_BACKEND_PORT` | `8000` | 本层监听端口 |
| `YCSP_COMPUTE_BASE_URL` | `http://127.0.0.1:8001` | 数据处理层地址 |
| `YCSP_DB_DRIVER` | `sqlite` | 数据库引擎：`sqlite`（仅开发/测试）/ `postgres` / `mysql` |
| `YCSP_DB_DSN` | `ycsp.db`（仅 sqlite） | sqlite 是文件路径；postgres / mysql 必须显式给连接串 |
| `YCSP_DB_MAX_OPEN_CONNS` | `0` | 连接池最大打开连接数；0 表示按驱动默认（SQLite 固定 1） |
| `YCSP_DB_MAX_IDLE_CONNS` | `0` | 连接池最大空闲连接数；0 表示按驱动默认（SQLite 固定 1） |
| `YCSP_DB_CONN_MAX_LIFETIME_SECONDS` | `0` | 连接最长存活秒数；0 表示按驱动默认（服务端数据库 30 分钟） |
| `YCSP_DB_AUTO_MIGRATE` | 开发/测试 `true`，生产 `false` | 启动时是否自动建表/补列 |
| `YCSP_TASK_WORKERS` | `0` | 任务 worker 数；0 表示按驱动自动（SQLite 1，Postgres/MySQL 4），范围 1~16 |
| `YCSP_SESSION_DRIVER` | `redis` | 会话存储：`redis` / `memory`（后者仅单机开发：重启即掉登录） |
| `YCSP_REDIS_URL` | `redis://127.0.0.1:6379/0` | Redis 地址；启动时连不上会直接失败并提示 |
| `YCSP_SESSION_LIFETIME_SECONDS` | `604800`（7 天） | 会话有效期 |
| `YCSP_COOKIE_SECURE` | `false` | 只在 HTTPS 部署时设为 `true` |
| `YCSP_LOGIN_MAX_ATTEMPTS` | `5` | 同一昵称在窗口内允许的失败次数 |
| `YCSP_LOGIN_IP_MAX_ATTEMPTS` | `20` | 同一地址在窗口内允许的失败次数 |
| `YCSP_LOGIN_BLOCK_SECONDS` | `900`（15 分钟） | 失败计数的窗口长度，也是锁定时的等待时间 |
| `YCSP_TRUSTED_PROXIES` | 空 | 可信反向代理的 IP / CIDR（逗号分隔）。只有直连地址命中这里才读 `X-Forwarded-For`；nginx 同机填 `127.0.0.1/32`。禁止 `0.0.0.0/0` |

### 生产数据库（PostgreSQL）

SQLite 只用于本地开发和测试：它固定单连接、写入串行，不适合生产。生产至少这样配置：

```dotenv
YCSP_ENV=production
YCSP_DB_DRIVER=postgres
YCSP_DB_DSN=host=db.internal port=5432 user=ycsp password=*** dbname=ycsp sslmode=require
YCSP_DB_AUTO_MIGRATE=false
YCSP_TASK_WORKERS=4
```

部署顺序：

1. 先跑一次 `go run . -migrate-only` 建表/补列（这个参数不受 `YCSP_DB_AUTO_MIGRATE` 影响）。
2. 再正常启动 `go run .`；生产默认不会在启动时自动改表。缺少表时启动会失败，并提示先跑 `-migrate-only`。
3. 连接池先用默认值（20 打开 / 10 空闲 / 30 分钟上限），有压测数据后再调
   `YCSP_DB_MAX_OPEN_CONNS` / `YCSP_DB_MAX_IDLE_CONNS` / `YCSP_DB_CONN_MAX_LIFETIME_SECONDS`。
4. `YCSP_TASK_WORKERS` 默认值：SQLite 1，Postgres / MySQL 4；范围 1~16。

`YCSP_ENV=production` 时如果仍是 `YCSP_DB_DRIVER=sqlite`，服务会拒绝启动。
本地开发和测试继续用 SQLite，零安装、不需要额外服务。

### 反向代理部署（nginx）

不配置时，后端只按**直连地址**限流：nginx 后面的所有学生会共用 nginx 的 IP，
一个人连续失败就可能把全班锁住。把 nginx 的地址加进 `YCSP_TRUSTED_PROXIES` 后，
后端才会从 `X-Forwarded-For` 里取真实客户端地址。

nginx 与后端在同一台机器时：

```nginx
server {
    listen 443 ssl;
    server_name crystal.example.com;

    root /srv/ycsp/frontend/dist;
    index index.html;

    location /api/ {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        # 关键：把直连客户端地址追加到 X-Forwarded-For
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

- **只填自己的代理地址 / 网段**：同机填 `127.0.0.1/32`；nginx 在容器里时填容器网络网段
  （如 `172.18.0.0/16`）或 nginx 的固定 IP。填 `0.0.0.0/0` 会被启动直接拒绝——
  那等于让任何客户端伪造地址绕过限流。
- **从右往左取第一个不受信地址**：客户端自己在请求里塞的 `X-Forwarded-For` 在左侧，会被忽略；
  nginx 追加的真实地址在右侧，才会被采用。
- **nginx 必须追加或覆盖，不能原样透传**：用 `$proxy_add_x_forwarded_for`（追加）或
  `$remote_addr`（覆盖）。若写成 `$http_x_forwarded_for`，客户端就能完全控制这个头。
- **后端端口不要对公网开放**：只让 nginx 访问 8000。直连后端的请求即使带伪造头，
  也会因为“直连地址不是可信代理”而被忽略。
- **容器部署**：后端需要监听 `0.0.0.0`（`YCSP_BACKEND_HOST=0.0.0.0`）才能被 nginx 容器访问，
  但不要把它映射到宿主机公网端口；让 nginx 与后端走同一个 Docker 网络，可信地址填该网络网段。
- **前面还有 CDN / 负载均衡时**：它们的地址也要列进 `YCSP_TRUSTED_PROXIES`；
  或者先用 nginx 的 `real_ip` 模块恢复真实 IP，再用 `proxy_set_header X-Forwarded-For $remote_addr;`
  只把恢复后的地址传给后端（此时后端只信任 nginx）。
- 转发头格式不合法时退回直连地址（安全优先），此时会暂时共用代理 IP 的计数，需要检查 nginx 配置。
- `X-Request-ID` 是普通请求头，nginx 默认会转发；后端也会在缺失时生成，因此不必额外配置。

如果希望 **nginx 访问日志也能按同一个编号检索**，可以在 `http` 块里定义一个 map
（有前端编号就沿用，没有就用 nginx 自己生成的），并在 `location /api/` 里传给后端：

```nginx
# http 块
map $http_x_request_id $ycsp_request_id {
    default $http_x_request_id;
    ""      $request_id;
}
log_format ycsp '$remote_addr - $request_method $uri $status request_id=$ycsp_request_id';
access_log /var/log/nginx/ycsp.access.log ycsp;

# location /api/ 里追加
proxy_set_header X-Request-ID $ycsp_request_id;
```

## 日志追踪

三层统一用 `X-Request-ID` 串起一次请求：

| 层 | 编号来源 | 日志示例 |
| --- | --- | --- |
| 前端 | 每次请求生成 `web-…`，放进请求头 | `[ycsp] request start request_id=web-… method=POST path=/api/tasks` |
| 数据管理层 | 沿用前端编号；没有或不合法时生成 `api-…`，并写回响应头 | `http: request_id=web-… client_ip=203.0.113.7 method=POST path=/api/tasks status=201 duration_ms=8` |
| 数据处理层 | 沿用数据管理层编号；没有或不合法时生成 `compute-…`，并写回响应头 | `request_id=web-… request method=POST path=/predict status=200 duration_ms=3` |

- **异步预测**：创建任务时的编号写进 `prediction_tasks.trace_id`；worker 调用数据处理层时继续使用，
  并记录 `tasks: start / succeeded / failed trace_id=… task_id=…`。
- **怎么查**：把前端控制台里的 `request_id`（或任务日志里的 `trace_id`）复制出来，在三层日志里搜索同一个值。
- **记录范围**：只记方法、路径、状态码、耗时与预测摘要；不记请求体、密码与 Cookie。
- **防日志注入**：编号最长 64 个字符，只允许字母、数字与 `-_.:`；含换行等字符的编号会被替换。

## 如何运行

```powershell
go mod tidy
go run .
```

启动后访问 <http://127.0.0.1:8000/api/health>。

- 本地开发默认用 SQLite：首次启动会自动创建 `backend/ycsp.db` 并建表；想回到空库，删掉这个文件即可。
- 生产用 PostgreSQL：设置 `YCSP_ENV=production` + `YCSP_DB_DRIVER=postgres` + `YCSP_DB_DSN`；先 `go run . -migrate-only`，再 `go run .`（见《生产数据库（PostgreSQL）》）。
- 注册与登录需要 Redis（默认 `redis://127.0.0.1:6379/0`）；本地没有 Redis 时设 `YCSP_SESSION_DRIVER=memory`。
- 任务接口要先启动数据处理层：联系不上或超时时任务会以 `compute_unreachable` / `compute_timeout` 失败（不会一直卡在 pending）。
- 拉不到 Go 依赖时（`proxy.golang.org` 不通）：先 `go env -w GOPROXY=https://goproxy.cn,direct`，再 `go mod tidy`。
- 部署在 nginx 后面时配置 `YCSP_TRUSTED_PROXIES`，否则登录限流会按 nginx 的 IP 统计（见《反向代理部署（nginx）》）。
- 测试：`go test ./...`；加 `-cover` 看覆盖率，加 `-race` 检查并发创建与超时路径（当前 `env` 100%、`store` 92.9%、`auth` 91.4%、`session` 84.1%、
  `api` 84.0%、`traceid` 90%，任务与账号包在 74%~85%；只有启动入口 `main()` 不测，因为它会真的占用端口）。
- 想在本机验证 Postgres：准备一个测试库，然后设
  `YCSP_TEST_DB_DRIVER=postgres` 和 `YCSP_TEST_DB_DSN=...` 再跑 `go test ./...`；
  每个测试会自动建独立 schema 并在结束时删除。CI 的 backend job 已包含一次 Postgres 跑测。
