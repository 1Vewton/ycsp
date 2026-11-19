> [English](./README-workflow.md) | 简体中文

# GitHub Actions 持续集成（.github）

持续集成配置：每次 push / PR 自动检查三层能否安装依赖、通过格式检查、编译与构建。

> 为什么不叫 README：GitHub 会把目录里的 README 文件渲染成该目录的说明，
> 也会把 `.github/` 下的 README 当作仓库首页 README 的候选；这份文档只讲 CI，不该顶替 README 的位置。

| 文件 | 作用 |
| --- | --- |
| `workflows/ci.yml` | 三层并行检查（Python / Go / 前端），任一失败即失败 |

## 每层检查什么

| 层 | 检查项 |
| --- | --- |
| compute | `uv sync --frozen`（锁文件与 `pyproject.toml` 一致）+ 能否 `import app.main` + `uv run pytest --cov=app`（覆盖率低于 90% 失败） |
| backend | `gofmt` 无差异、`go vet`、`go build`、`go test -cover ./...`（SQLite）、再对 PostgreSQL 服务跑一次 `go test ./...`、`go test -race ./...`（并发创建与超时路径） |
| integration | 用 uv 启动真实 Python compute，再跑 `go test -tags=integration ./integration/...`：Python 错误必须落到 Go 任务上，且前端语言包有中文文案 |
| frontend | `npm ci`、`npm run check-codes`（后端每个 code 都有中英文案）、`npm run test`（Vitest + jsdom）、`npm run type-check`、`npm run build`、`npm run check-csp`（构建产物不含 `eval` / `Function` 构造器，严格 CSP 可用） |

## 本地对齐 CI

```powershell
cd compute;  uv sync --frozen; uv run python -c "import app.main"; uv run pytest --cov=app --cov-report=term-missing
cd backend;  gofmt -l .; go vet ./...; go build ./...; go test -cover ./...; go test -race ./...
cd frontend; npm ci; npm run test; npm run type-check; npm run build; npm run check-csp
```

PostgreSQL 跑测已经包含在 backend CI job 里。本机复现时，把测试套件指向一个测试库：

```powershell
cd backend
$env:YCSP_TEST_DB_DRIVER="postgres"
$env:YCSP_TEST_DB_DSN="host=127.0.0.1 port=5432 user=ycsp password=ycsp dbname=ycsp_test sslmode=disable"
go test -count=1 ./...
```

每个测试会建独立 schema 并在结束时删除，所以测试库不需要在两次运行之间清空。

跨层测试要先启动真实 compute：

```powershell
# 终端 1
cd compute
uv run uvicorn app.main:app --host 127.0.0.1 --port 8001

# 终端 2
cd backend
$env:YCSP_TEST_COMPUTE_URL="http://127.0.0.1:8001"
go test -tags=integration ./integration/...
```

> CI 环境没有 `.env`，三层都使用代码里的默认值——这也顺带保证了「默认值能跑通」。

> action 版本固定到具体 release（浮动大版本 tag 并非都存在，例如 `setup-uv` 只发布了 `v10.3.0`）；升级时请先查各自的 release 再改。
