> English | [简体中文](./README.zh-CN.md)

# Data processing layer · HTTP API (compute/app/api)

HTTP endpoints called by the data management layer.

| File | Purpose |
| --- | --- |
| `health.py` | `GET /health` route |
| `predict.py` | `POST /predict` route: wraps the domain prediction result in the unified response |
| `request_context.py` | Request IDs: reuse/generate `X-Request-ID`, write it back to the response header, log requests with the ID |
| `serialization.py` | Result to dictionary: unit-cell parameters, atom coordinates, floats rounded for display |
| `display.py` | Display precision (how many decimal places) and rounding functions |
| `errors.py` | Framework-level errors (invalid request body, unexpected exception) also become unified responses, so the UI never shows an English error |

Status codes live in [../constants/codes.py](../constants/codes.py) (language-neutral; the frontend translates the copy).

