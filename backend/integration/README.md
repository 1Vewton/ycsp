> English | [简体中文](./README.zh-CN.md)

# Data management layer · Cross-layer integration tests (backend/integration)

A build-tagged test that runs the **real data processing layer** and the Go task service together.
It answers task 4.3: when Python rejects a formula, the Go task must keep the same `code` + `params`, and the frontend language packs must contain a Chinese explanation for that code.

## What it checks

- `POST /api/tasks` sends `NaCl2`, `Xx2` and `H2O` to the real Python `/predict`.
- The task ends as `failed` with `formula_charge_unbalanced`, `element_not_supported` and `composition_not_supported` respectively.
- The failure keeps the upstream `params` (formula / symbol / position).
- `frontend/src/locales/zh.ts` has the key and the expected Chinese sentence; `en.ts` has the key.
- Rendering the code into the result page is covered by `frontend/tests/ResultView.test.ts`.

## Run locally

Start the data processing layer in one terminal:

```powershell
cd compute
uv run uvicorn app.main:app --host 127.0.0.1 --port 8001
```

Then run the tagged test in another terminal:

```powershell
cd backend
$env:YCSP_TEST_COMPUTE_URL = "http://127.0.0.1:8001"
go test -tags=integration ./integration/...
```

Without `YCSP_TEST_COMPUTE_URL` the test skips; the regular `go test ./...` does not compile this directory because it needs the `integration` build tag.

CI runs this as a separate job that starts compute first and then runs the tagged test.
