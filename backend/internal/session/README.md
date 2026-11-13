> English | [简体中文](./README.zh-CN.md)

# Data management layer · Session storage (backend/internal/session)

A login session is a "random ID → user ID" mapping; the browser only receives the ID (stored in a cookie).

| File | Purpose |
| --- | --- |
| `session.go` | `Store` interface, `ErrSessionNotFound`, random ID generation |
| `redis.go` | `RedisStore`: sessions in Redis, shared across instances and surviving restarts (default) |
| `memory.go` | `MemoryStore`: process memory, suitable only for single-machine development |

## Where the ID comes from

32 bytes from `crypto/rand` → base64url (43 characters). 256 bits of randomness, so the chance of guessing is negligible.
The ID itself contains no user information, and the user cannot be derived from it.

## What Redis stores

```text
ycsp:session:<random ID>  →  user ID (string), TTL = session lifetime
```

Only the user ID is stored, not the whole user object: a renamed user keeps a valid session, and there are no multiple copies to synchronize.

## Choosing an implementation

| Driver | Suitable for | Cost |
| --- | --- | --- |
| `redis` (default) | Production, multiple instances | Requires Redis; startup fails with a hint when it cannot connect |
| `memory` | Temporary development without Redis | A restart signs everyone out; not shared across instances |

## Known limitations

- The lifetime is fixed (no sliding renewal): 7 days by default, configurable via `YCSP_SESSION_LIFETIME_SECONDS`.
- Signing out deletes only the current session; sessions for the same account on other devices remain valid.
- Changing or resetting a password does not invalidate existing sessions either; there is no "sign out everywhere" yet.

Tests: `go test ./internal/session/` (Redis is provided by in-process `miniredis`; CI needs no Redis service).
