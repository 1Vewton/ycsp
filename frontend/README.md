> English | [简体中文](./README.zh-CN.md)

# Frontend (frontend)

The web interface students see and use (TypeScript + Vue 3 + Vite). It talks only to the data management layer.

## Responsibilities

- Student interface: enter a chemical formula, follow progress, inspect the 3D structure and its explanation, sign in / register.
- Talks only to the data management layer, never directly to the data processing layer; computation and data stay on the backend.
- Does not duplicate the backend's element and charge data: validation covers only what the frontend can decide on its own (see `src/formulaValidation.ts`).
- All student-facing text comes from language packs; no Chinese or English is hard-coded in components.

## Common commands

The dev port and backend address have defaults. To change them, copy `frontend/.env.example` to `frontend/.env`:

```powershell
Copy-Item .env.example .env
```

| Command | Purpose |
| --- | --- |
| `npm install` | Install dependencies for the first time |
| `npm run dev` | Dev server (<http://127.0.0.1:5173>) |
| `npm run build` | Bundle into `dist/` |
| `npm run type-check` | TypeScript type checking |
| `npm run check-codes` | Check that every backend / compute code has Chinese and English copy (CI also runs this) |
| `npm run check-csp` | After `npm run build`, check the production bundle for `eval` / `Function` constructors (CI also runs this) |
| `npm run test` | Run the automated tests once (pure functions / API / components; CI also runs this) |
| `npm run test:watch` | Watch changes and keep running tests |

In development, Vite proxies `/api` to `VITE_BACKEND_BASE_URL` (default `http://127.0.0.1:8000`, override it in `frontend/.env`), so start compute and backend first.

Automated tests live in `tests/` and use Vitest + jsdom; they do not need a running backend. See [tests/README.md](./tests/README.md).

The error-code copy check script lives in `scripts/`; see [scripts/README.md](./scripts/README.md).

## Pages

The seven pages use vue-router's hash mode (addresses look like `/#/predict`): opening or refreshing any page never returns 404, and deployment needs no "fall back to index.html" rule. Routes are registered in `src/router.ts`.

| Route | View | Content |
| --- | --- | --- |
| `/#/` | `HomeView` | What the platform does, three steps to start, supported range; background service connectivity check |
| `/#/predict` | `PredictView` | Formula input with instant validation feedback, example shortcuts, start prediction |
| `/#/result/:taskId` | `ResultView` | Poll progress, cancel a queued task, failure reason and duration, 3D structure and a student-friendly explanation |
| `/#/login` | `LoginView` | Shared login / registration form, optional security question, forgot-password link; when signed in, shows the account and sign-out |
| `/#/reset-password` | `ResetPasswordView` | Look up the security question by nickname, answer it and set a new password; signs in on success |
| `/#/account` | `AccountView` | Change password, set or update the security question |
| `/#/history` | `HistoryView` | The signed-in account's prediction history, newest first, with a link to each result and "load more" paging |

Each page's current job and next integration step: [src/views/README.md](./src/views/README.md).
Login is not required for prediction; tasks can be created without signing in, they simply have no owner and do not appear in history.

## API conventions

| Convention | What the frontend does |
| --- | --- |
| Single entry point | Calls only `/api/...` on the data management layer; Vite proxies it in development and a same-origin reverse proxy handles it in production (see [../backend/README.md](../backend/README.md)) |
| Envelope | Responses are `{ service, status, code, params? }`, with `task` / `result` / `user` on success; anything other than `status: ok` is treated as a failure |
| Error copy | The frontend receives only language-neutral `code` + `params` and uses `tCode()` to read the current language from `src/locales/`; the backend never returns Chinese or English |
| Session | The backend sets the login cookie as `HttpOnly`; the frontend neither reads nor stores it; same-origin requests carry it automatically |
| Task flow | `POST /api/tasks` returns `pending` immediately → poll `GET /api/tasks/:id` every second → fetch `/result` on success; only queued tasks can be cancelled |
| Tracing | Every request carries `X-Request-ID`; on failure `ApiError.requestId` keeps the same value |
| Content safety | Backend technical details are never shown to students; unknown codes display `unknown_code`, and details go only to the console |

See [../backend/README.md](../backend/README.md) "Public API" for endpoints and fields; see [src/api/README.md](./src/api/README.md) for the frontend wrapper.

## Log tracing

Every `/api` request carries `X-Request-ID` (`web-…`). The console logs
`[ycsp] request start / done / failed request_id=…`; on failure `ApiError.requestId` keeps the same value.
Search that ID in the backend and data processing logs. Details are in [src/api/README.md](./src/api/README.md)
and [../backend/README.md](../backend/README.md) "Log tracing".

## Strict CSP

The 3D view uses [Three.js](https://threejs.org/) and does not need `unsafe-eval`.
After `npm run build`, `npm run check-csp` scans `dist/assets/*.js` for `eval` and
`Function` constructors; CI runs it after the build.

## Known limitations

- **Narrow support**: ionic crystals (1:1, 1:2, 2:1; including wurtzite ZnO / ZnS and rutile MgF2 / TiO2), elemental metals, and covalent crystals such as C / Si / Ge / SiC; multi-element formulas, 2:3 stoichiometry, and 3-coordination are explicitly reported as unsupported (see [../compute/README.md](../compute/README.md) "Known limitations").
- **The 3D view requires WebGL2**: devices without WebGL2 only show the text explanation; the structure image is a single unit cell, with no symmetry expansion or periodic repetition.
- **Account features**: registration, sign in/out, password change, security-question reset and prediction history. Reset needs a security question, which must be set at registration or in account settings; anonymous predictions are not saved to history; changing or resetting a password does not invalidate sessions on other devices yet.
- **Must be same-origin with the data management layer**: the frontend requests only the relative path `/api`; if the static site and backend are deployed on different domains, a reverse proxy or CORS is needed, and CORS is not currently implemented.
- **Only Chinese / English**: other languages fall back to English and are not guaranteed to be fully translated.
- **Cannot work offline**: prediction, progress, and results all depend on the backend; the frontend only displays them.
- **Frontend tests do not use a real backend**: `npm run test` mocks `fetch` to cover pure functions, the API layer, and components; real three-layer regression still relies on manual checks or future E2E.

Third-party icon and 3D rendering licenses: [THIRD_PARTY_NOTICES.md](./THIRD_PARTY_NOTICES.md).

Interface copy lives by language in `src/locales/`; components use only tags and call `t()` (see [src/README.md](./src/README.md)). Chinese is the default and English is switchable; missing translations fall back to English.
