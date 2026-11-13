> English | [简体中文](./README.zh-CN.md)

# Frontend · Check scripts (frontend/scripts)

Standalone scripts for development and CI; they are not part of the browser build.

| File | Purpose |
| --- | --- |
| `check-codes.mjs` | Checks that every backend / compute code has copy in both `src/locales/en.ts` and `zh.ts`; fails if one is missing |
| `check-csp.mjs` | Scans `dist/assets/*.js` after a build for `eval`, `new Function`, and `Function()`; fails if any appears |

## How to use

```powershell
cd frontend
npm run check-codes

npm run build
npm run check-csp
```

## Why it lives in the frontend

Error-code copy is maintained by the frontend, but the codes originate in the other two layers. The script reads `../backend/internal/codes/codes.go` and `../compute/app/constants/codes.py` directly instead of maintaining a second list; the CI frontend job runs it, so forgetting copy for a new code fails the build.

`check-csp.mjs` reads the production bundle, so run `npm run build` first. A strict Content-Security-Policy blocks `eval` and `Function` constructors; the 3D view must not need `unsafe-eval`, and CI runs this check right after the build.
