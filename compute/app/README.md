> English | [简体中文](./README.zh-CN.md)

# Data processing layer · Application package (compute/app)

Start it with `python -m app.main` (see the parent [compute/README.md](../README.md)).

| Path | Purpose |
| --- | --- |
| `main.py` | Entry point: assemble the FastAPI app and start uvicorn |
| `api/` | HTTP endpoints, serialization, and unified error responses, see [api/README.md](./api/README.md) |
| `domain/` | Crystallography data types and prediction rules, see [domain/README.md](./domain/README.md) |
| `constants/` | Element data, structure prototype data, thresholds, and status codes, see [constants/README.md](./constants/README.md) |
| `env/` | Environment configuration (with defaults), see [env/README.md](./env/README.md) |
| `logging_setup.py` | Log format: every application log carries a request_id |
| `__init__.py` | Package metadata: service name and version |
