> English | [简体中文](./README.zh-CN.md)

# Frontend · Source (frontend/src)

Frontend source compiled by Vite; start it with `npm run dev` (see the parent [frontend/README.md](../README.md)).

| File | Purpose |
| --- | --- |
| `main.ts` | Entry: mount the Vue app, install routing and global styles |
| `App.vue` | Application shell: header (title / navigation / sign-in state / language switch) + router outlet |
| `router.ts` | Route table: home, predict, result, login, reset password, account settings, prediction history (hash mode) |
| `session.ts` | Site-wide sign-in state: current account, restore / sign in / register / sign out, password change, security question update and password reset, shared by the header and account pages |
| `styles.css` | Shared style variables and base classes (page container, buttons, cards, status messages) |
| `formulaValidation.ts` | Formula syntax validation (pure functions): characters, brackets, subscripts; elements and charges are left to the data processing layer |
| `structureScene.ts` | Pure geometry for the 3D view: unit-cell vectors, fractional → Cartesian coordinates, bonds, cell edges, camera radius |
| `structureRenderer.ts` | Three.js renderer (lazy-loaded): atoms, bonds, cell wireframe, lighting, reset view, resource disposal |
| `structureExplanation.ts` | Pure functions for result explanations: unit-cell shape detection, edge length and angle formatting |
| `api/` | Data management API wrappers (tasks, accounts, request IDs, unified errors), see [api/README.md](./api/README.md) |
| `components/` | Shared page components, see [components/README.md](./components/README.md) |
| `views/` | Page components, see [views/README.md](./views/README.md) |
| `locales/` | Language packs and the `t()` lookup entry point, see [locales/README.md](./locales/README.md) |

## Icons

The interface uses Font Awesome (`@fortawesome/vue-fontawesome` + `@fortawesome/free-solid-svg-icons`).

- Import icons one by one in the component that uses them (`import { faXxx } from '@fortawesome/free-solid-svg-icons/faXxx'`); only icons actually used end up in the bundle.
- Icons are decorative only: text always remains beside them, and they carry `aria-hidden="true"` so screen readers do not announce extra graphics.
- Icon color follows the text (`currentColor`); buttons and links use `gap` for spacing instead of extra margins.

## 3D structures

Structure results are rendered with [Three.js](https://threejs.org/) (MIT; copyright notice in [THIRD_PARTY_NOTICES.md](../THIRD_PARTY_NOTICES.md)).

- `structureScene.ts` is a pure module: it turns cell parameters and fractional coordinates into atom positions, bonds, cell edges, and the camera radius. It is covered by `tests/structureScene.test.ts`.
- `StructureViewer.vue` dynamically imports `structureRenderer.ts` only when the result page succeeds; Three.js is split into its own ~560 kB chunk, so the home and predict pages never download it.
- Bonds are the first coordination shell: pairs within 1.2× the shortest distance; ionic crystals only connect different ions, and metals have no bonds.
- Rendering needs WebGL; when unavailable, a translatable hint is shown and the text result remains readable.
- The renderer uses no `eval` / `Function` constructor, so a strict CSP does not need `unsafe-eval`; `npm run check-csp` verifies the built bundle after `npm run build`.

