> English | [简体中文](./README.zh-CN.md)

# Frontend · API (frontend/src/api)

HTTP request wrappers for the data management layer; pages receive structured data and never build URLs themselves.

| File | Purpose |
| --- | --- |
| `client.ts` | Unified response envelope and failure conversion (`ApiError` / `problemFromError`); shared by task and account APIs |
| `requestId.ts` | Generates `X-Request-ID` (`web-…`) and centralizes the header name and length limit |
| `tasks.ts` | Create / query / cancel prediction tasks, read structure results, and page through the signed-in account's history |
| `auth.ts` | Register, sign in, sign out, get the current account, change password, set the security question, look up the question, reset the password; `fetchCurrentUser()` returns `null` when signed out |

## Conventions

- Call only `/api/...` on the data management layer, never the data processing layer directly; Vite proxies it in development (see [../../README.md](../../README.md)).
- Every request carries `X-Request-ID`: the data management layer and data processing layer reuse it, the console logs
  `[ycsp] request ... request_id=...`, and on failure `ApiError.requestId` keeps the same value.
- Failures always throw `ApiError` (`code` + `params`); pages look up copy by code in [`../locales`](../locales/README.md), and the request layer never hard-codes sentences.
- Network failures and non-JSON responses use `request_failed` / `internal_error`; technical details go only to the console, never to the UI.
- The backend sets the session cookie as `HttpOnly`; the frontend neither reads nor stores it; same-origin requests carry it automatically.
