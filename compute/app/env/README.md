> English | [简体中文](./README.zh-CN.md)

# Data processing layer · Environment configuration (compute/app/env)

Reads environment variables uniformly and returns defaults when a variable is unset (`.env` lives in `compute/` and is optional).

| Name | Purpose |
| --- | --- |
| `get_env(key, fallback)` | String setting |
| `get_env_int(key, fallback)` | Integer setting (e.g. port) |
| `SERVICE_HOST` / `SERVICE_PORT` | Listen address and port for this layer (already parsed values) |
