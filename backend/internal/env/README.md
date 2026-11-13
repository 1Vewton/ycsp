> English | [简体中文](./README.zh-CN.md)

# Data management layer · Environment variables (backend/internal/env)

Reads environment variables uniformly and returns defaults when a variable is unset.

| Function | Purpose |
| --- | --- |
| `Load(path)` | Read `.env`; a missing file only logs a hint, it does not abort |
| `Get(key, fallback)` | String setting |
| `GetInt(key, fallback)` | Integer setting (e.g. port) |
