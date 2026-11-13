> English | [简体中文](./README.zh-CN.md)

# Data management layer · Login rate limiting (backend/internal/ratelimit)

Limits consecutive login failures to block brute-force password guessing. Counters use the same backend as sessions: Redis (default) or process memory (local development).

| File | Purpose |
| --- | --- |
| `counter.go` | `Counter` interface and `Policy` thresholds |
| `redis.go` | `RedisCounter`: `INCR` + expiry window, shared across instances |
| `memory.go` | `MemoryCounter`: process memory, for single-machine development |
| `guard.go` | `Guard`: decides whether to reject, records failures, clears on successful sign-in |

## Rate-limit policy (defaults, configurable in `.env`)

| Dimension | Key | Default | Meaning |
| --- | --- | --- | --- |
| Same display name | `ycsp:login:account:<lowercase name>` | 5 per 15 minutes | Stops guessing one account |
| Same address | `ycsp:login:address:<IP>` | 20 per 15 minutes | Stops credential stuffing across accounts |

- Once the threshold is reached, **even a correct password is rejected** until the window expires.
- A successful sign-in clears only the **display-name** counter: the address counter remains, otherwise compromising one account could reset the whole machine's limit.
- Display names are lowercased so `Alice` / `alice` cannot bypass the limit.

## Why a fixed window

`INCR` + `EXPIRE` are small atomic Redis operations, enough to stop scripted password guessing without Lua scripts or extra data structures. The trade-off is that up to about twice the allowed attempts may pass around a window boundary — acceptable for a teaching scenario.

## Known limitations

- Counters are keyed by the **direct peer address** and do not trust `X-Forwarded-For` (anti-spoofing). Behind nginx you must configure `YCSP_TRUSTED_PROXIES`, otherwise all students share the proxy IP's failure count; see [../../README.md](../../README.md) "Reverse proxy deployment (nginx)".
- `X-Forwarded-For` is parsed only when the direct peer matches a trusted proxy, and the first untrusted address is taken from right to left: forged values a client puts on the left have no effect.
- If the forwarded header is malformed, fall back to the direct peer address (safety first); during that time the proxy IP's counter is shared.
- There is no "unlock" endpoint: wait for the window to expire (15 minutes by default).
- Password-reset answer failures use the same counters as login failures, so they also contribute to the lockout.
- Counting uses a fixed window, with no exponential backoff.

Tests: `go test ./internal/ratelimit/` (Redis is provided by in-process `miniredis`).
