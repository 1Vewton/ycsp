> English | [简体中文](./README.zh-CN.md)

# Data management layer · Data processing client (backend/internal/compute)

Calls the data processing layer's `POST /predict` and hands the result to the task service for persistence. The contract is documented in
[`compute/README.md`](../../../compute/README.md) "Public API".

| File | Purpose |
| --- | --- |
| `client.go` | Sends requests, configurable timeout (10 s by default), passes `X-Request-ID` through, parses the unified response |
| `types.go` | `Prediction` / `Candidate` / `Cell`; `rule`, `caveats`, and `atoms` keep their raw JSON |
| `errors.go` | Failure kinds + `Describe`: turns client errors into frontend-usable `code` and `params` |

## How errors are classified

| Situation | Client error | Stored code |
| --- | --- | --- |
| Upstream 400 + `{status:error, code}` (e.g. unbalanced charges) | `PredictionError` | **Pass through** the upstream code; the frontend already has copy |
| Connection failure | `TransportError` | `compute_unreachable` (param `url`) |
| Timeout | `TimeoutError` | `compute_timeout` |
| Cannot read the response | `ProtocolError` | `compute_read_failed` |
| Status code is not 200 | `ProtocolError` | `compute_bad_status` (param `status_code`) |
| Response is not valid JSON | `ProtocolError` | `compute_invalid_json` |
| 200 but no candidate structure | `ProtocolError` | `compute_no_result` |

`rule` / `caveats` / `atoms` are kept as `json.RawMessage`: bytes are unchanged when stored and returned,
so the backend does not need to describe the structure again.

Tests: `go test ./internal/compute/` (uses `httptest` to fake the data processing layer; no real Python process needed).
