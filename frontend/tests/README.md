> English | [简体中文](./README.zh-CN.md)

# Frontend · Automated tests (frontend/tests)

Runs pure functions, API wrappers, and component tests with [Vitest](https://vitest.dev/); components run in jsdom and **do not need a running backend**.

## How to run

```powershell
npm run test        # run once (used by CI)
npm run test:watch  # watch files and rerun on changes
```

Configuration: [`../vitest.config.ts`](../vitest.config.ts).

## What is covered

| File | Coverage |
| --- | --- |
| `formulaValidation.test.ts` | Syntax validation: empty input, length, invalid characters and positions, bracket pairing, zero subscript |
| `structureScene.test.ts` | 3D geometry: cubic / hexagonal cell vectors, fractional → Cartesian, bond selection (ionic / covalent / metallic), degenerate cell rejection |
| `structureExplanation.test.ts` | Unit-cell shape detection (cubic / tetragonal / hexagonal / general) and edge/angle formatting |
| `locales.test.ts` | Default Chinese, language persistence and switching, placeholder replacement, fallback for unknown codes |
| `apiTasks.test.ts` | Task API wrappers: request shape, error envelope → `ApiError`, non-JSON and offline, history paging |
| `apiAuth.test.ts` | Account API: register / sign in / sign out / current account, password change, security question, reset, null when signed out, password rules preserved |
| `requestId.test.ts` | Request IDs: `web-` prefix, length limit, unique per call, consistent header name |
| `session.test.ts` | Site-wide sign-in state: concurrent merge, restore, sign in / register / sign out, password change and reset, sign-out failure keeps the account |
| `LoginView.test.ts` | Login and registration form: mode switch, validation, per-rule password hints, optional security question, forgot-password link, in-site redirect, signed-in panel |
| `AccountView.test.ts` | Account settings: password change, mismatch handling, password rules, security question update |
| `ResetPasswordView.test.ts` | Reset flow: question lookup, answer and new password, unavailable accounts, local password mismatch |
| `HistoryView.test.ts` | Prediction history: signed-out prompt, list rendering, load more, error retry |
| `App.test.ts` | Header sign-in state: sign-in entry, display name and sign-out, sign-out failure message |
| `FormulaInput.test.ts` | Input component's three feedback levels, v-model, and accessibility attributes |
| `structureExplanationComponent.test.ts` | Result explanation component: rule / caveats from the backend, caveat hiding rule |
| `ResultView.test.ts` | Result page state machine: polling, failure, timeout copy, cancel, retry, structure result failing separately |

`fixtures.ts` holds shared samples (NaCl structure and task snapshot); it is not a test file.

## Conventions

- When asserting interface text, look it up with `t()` / `tCode()` instead of hard-coding Chinese or English: changing language or copy never requires test changes.
- The result page advances its 1-second polling with a fake clock (`vi.useFakeTimers`); tests never really wait.
- The 3D viewer is replaced with `stubs: { StructureViewer: true }` in component tests: Three.js needs WebGL, so unit tests only verify "whether the result was handed to the viewer".
- No real backend: `fetch` is replaced with `vi.stubGlobal`, and failure paths use forged responses.
- Do not rely on case to distinguish file names (Windows and macOS are case-insensitive by default): the component test is named `structureExplanationComponent.test.ts`.
