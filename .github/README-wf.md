# GitHub Actions（.github）

持续集成配置：每次 push / PR 自动检查三层能否安装依赖、通过格式检查、编译与构建。

| 文件 | 作用 |
| --- | --- |
| `workflows/ci.yml` | 三层并行检查（Python / Go / 前端），任一失败即失败 |

## 每层检查什么

| 层 | 检查项 |
| --- | --- |
| compute | `uv sync --frozen`（锁文件与 `pyproject.toml` 一致）+ 能否 `import app.main` + `uv run pytest --cov=app`（覆盖率低于 90% 失败） |
| backend | `gofmt` 无差异、`go vet`、`go build`、`go test -cover ./...` |
| frontend | `npm ci`、`npm run type-check`、`npm run build` |

## 本地对齐 CI

```powershell
cd compute;  uv sync --frozen; uv run python -c "import app.main"; uv run pytest --cov=app --cov-report=term-missing
cd backend;  gofmt -l .; go vet ./...; go build ./...; go test -cover ./...
cd frontend; npm ci; npm run type-check; npm run build
```

> CI 环境没有 `.env`，三层都使用代码里的默认值——这也顺带保证了「默认值能跑通」。

> action 版本固定到具体 release（浮动大版本 tag 并非都存在，例如 `setup-uv` 只发布了 `v10.3.0`）；升级时请先查各自的 release 再改。
