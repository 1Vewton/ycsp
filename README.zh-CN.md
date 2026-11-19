> [English](./README.md) | 简体中文

# ycsp

面向青少年的 CSP（晶体结构预测）教学平台：输入物质的化学成分，预测它最可能形成的晶体结构。学生不需要任何计算或编程基础。

## 项目定位

面向青少年，用户不预设任何计算或编程基础；以“中学生能否独立看懂并操作”为唯一验收标准。

## 整体架构

| 分层 | 职责 | 技术栈 |
| --- | --- | --- |
| 前端 | 学生界面与晶体结构可视化 | TypeScript + Vue |
| 数据管理 | 统一对外接口、数据存储、转发计算请求 | Go + Gin + GORM |
| 数据处理 | 晶体结构预测与数值计算 | Python + FastAPI（数值计算按需引入 NumPy） |

请求流向：`前端(Vue) → 数据管理(Go) → 数据处理(Python)`；前端只与数据管理通信，由它再调用数据处理。

本地开发和测试时数据管理层用 SQLite（零安装）；生产使用 PostgreSQL，配置与迁移步骤见 [backend README](./backend/README.zh-CN.md)。

## 多语言

产品要能切换界面语言，让不同语言的用户都能独立操作：

- 默认中文，并支持英文（中/英可切换）。
- 文案与代码分离：前端用语言包（见 `frontend/src/locales/`），不在组件里写死文字。
- 后端只返回错误码与参数，由前端按当前语言翻译，避免把某种语言的提示写死在后端。

## 安装与启动

需要 Go 1.25+、Node.js 22+、[uv](https://docs.astral.sh/uv/)（数据处理层用它管理 Python 依赖与虚拟环境），
以及 [Redis](https://redis.io/)（保存登录会话；本地没有 Redis 时可临时用内存会话，见下）。

端口与地址都有默认值，直接就能跑；需要改的话，把对应层的 `.env.example` 复制为 `.env` 再编辑：

```powershell
Copy-Item compute\.env.example compute\.env
Copy-Item backend\.env.example backend\.env
Copy-Item frontend\.env.example frontend\.env
```

然后在三个终端分别启动：

```powershell
# 数据处理层（默认 8001）
cd compute
uv run python -m app.main

# 数据管理层（8000）
cd backend
go mod tidy; go run .

# 前端（5173）
cd frontend
npm install; npm run dev
```

打开 <http://localhost:5173/>，在首页点“检查后台服务”；若显示“后台服务正常，可以开始预测了。”即三层已打通。

> 本地没有 Redis 时，把 `backend/.env` 里的 `YCSP_SESSION_DRIVER` 改成 `memory` 即可跑通注册与登录
> （会话只存在进程内存里，重启后需要重新登录）。

## Docker Compose（开发）

需要 Docker Desktop 或 Docker Engine + Compose v2。这是开发栈：源码通过 bind mount 挂进容器，前端 Vite、compute uvicorn、backend Air 都会热重载；数据库用容器里的 Postgres，宿主不需要安装 Go / Node / uv / Redis。

```powershell
docker compose up --build
```

首次构建会下载 Go / Python / npm 依赖；之后启动更快。打开 <http://localhost:5173/>，在首页点“检查后台服务”。

| 服务 | 地址 | 说明 |
| --- | --- | --- |
| `frontend` | <http://localhost:5173> | Vite 开发服务器 |
| `backend` | <http://localhost:8000> | Go API |
| `compute` | <http://localhost:8001> | FastAPI；接口文档 `/docs` |
| `postgres` | `127.0.0.1:5432` | 开发账号/密码 `ycsp` / `ycsp`，数据库 `ycsp` |
| `redis` | `127.0.0.1:6380` | 登录会话（容器内仍是 6379；宿主端口避开本地 Redis） |

常用命令：

```powershell
docker compose logs -f backend
docker compose logs -f compute
docker compose down        # 保留数据库与依赖卷
docker compose down -v     # 连同 Postgres、Redis 和依赖卷一起删除
```

说明：

- 这套 Compose 用 PostgreSQL，不是 SQLite，开发时也走生产数据库路径。
- 账号密码和端口只用于本地开发，不要用于生产。
- 改了 `go.mod` / `uv.lock` / `package.json` 后，依赖卷可能还是旧的：先 `docker compose down -v`，再 `docker compose up --build`。
- 如果宿主机端口被占用，改 `docker-compose.yml` 里 `ports` 的左侧端口。

## 目录导航

| 目录 | 分层 | 文档 |
| --- | --- | --- |
| `frontend/` | 前端 | [README](./frontend/README.zh-CN.md) |
| `backend/` | 数据管理 | [README](./backend/README.zh-CN.md) |
| `compute/` | 数据处理 | [README](./compute/README.zh-CN.md) |
| `slides/` | 演示文稿 | [README](./slides/README.zh-CN.md) |
| `docker-compose.yml` | 开发环境 | 见上方《Docker Compose（开发）》 |

CI 配置见 [`.github/README-workflow.zh-CN.md`](./.github/README-workflow.zh-CN.md)，开发计划见 [TASKS.md](./TASKS.md)。
