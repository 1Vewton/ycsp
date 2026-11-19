> English | [简体中文](./README-workflow.zh-CN.md)

# GitHub Actions continuous integration (.github)

CI configuration: every push / PR automatically checks that all three layers can install dependencies, pass formatting checks, compile, and build.

> Why it is not called README: GitHub renders a directory's README file as that directory's description,
> and also treats a README under `.github/` as a candidate for the repository home page; this document is only about CI and should not take the README's place.

| File | Purpose |
| --- | --- |
| `workflows/ci.yml` | Three parallel checks (Python / Go / frontend); any failure fails the run |

## What each layer checks

| Layer | Checks |
| --- | --- |
| compute | `uv sync --frozen` (lock file matches `pyproject.toml`) + `import app.main` works + `uv run pytest --cov=app` (fails below 90% coverage) |
| backend | no `gofmt` diff, `go vet`, `go build`, `go test -cover ./...` (SQLite), a second `go test ./...` against a PostgreSQL service, `go test -race ./...` (concurrent task creation and timeout paths) |
| integration | starts the real Python compute service with uv, then runs `go test -tags=integration ./integration/...`: Python errors must reach the Go task and the frontend language packs must contain Chinese copy |
| frontend | `npm ci`, `npm run check-codes` (every backend code has Chinese and English copy), `npm run test` (Vitest + jsdom), `npm run type-check`, `npm run build`, `npm run check-csp` (the built bundle contains no `eval` / `Function` constructor, so strict CSP works) |

## Aligning locally with CI

```powershell
cd compute;  uv sync --frozen; uv run python -c "import app.main"; uv run pytest --cov=app --cov-report=term-missing
cd backend;  gofmt -l .; go vet ./...; go build ./...; go test -cover ./...; go test -race ./...
cd frontend; npm ci; npm run test; npm run type-check; npm run build; npm run check-csp
```

PostgreSQL is part of the backend CI job. To repeat it locally, point the test suite at a test database:

```powershell
cd backend
$env:YCSP_TEST_DB_DRIVER="postgres"
$env:YCSP_TEST_DB_DSN="host=127.0.0.1 port=5432 user=ycsp password=ycsp dbname=ycsp_test sslmode=disable"
go test -count=1 ./...
```

Each test creates its own schema and drops it afterwards, so the test database does not need to be empty between runs.

The cross-layer test starts with the real compute service:

```powershell
# terminal 1
cd compute
uv run uvicorn app.main:app --host 127.0.0.1 --port 8001

# terminal 2
cd backend
$env:YCSP_TEST_COMPUTE_URL="http://127.0.0.1:8001"
go test -tags=integration ./integration/...
```

> CI has no `.env`, so all three layers use the defaults in code — this also guarantees that "the defaults work".

> Action versions are pinned to specific releases (floating major tags do not always exist; for example, `setup-uv` only published `v10.3.0`); check each release before upgrading.
