> English | [简体中文](./README.zh-CN.md)

# Frontend · Components (frontend/src/components)

Small widgets shared by pages; the pages themselves live in [`../views/`](../views/README.md).

| File | Purpose |
| --- | --- |
| `FormulaInput.vue` | Formula input: label, input, instant feedback (how to write / syntax looks good / which character is not understood) |
| `StructureExplanation.vue` | Result explanation: crystal family, structure type, coordination number, unit-cell parameters, rationale, and caveats |
| `StructureViewer.vue` | 3D crystal structure view: lazy-loaded Three.js renderer, unit-cell wireframe, reset view; shows a hint when WebGL is unavailable |

## Conventions

- Components only display and handle interaction; they do not decide whether a formula is correct: syntax validation lives in [`../formulaValidation.ts`](../formulaValidation.ts), and the page passes the result in as `problem`.
- All text is fetched through `t()` from [`../locales`](../locales/README.md); colors and other styles use variables from [`../styles.css`](../styles.css).
- `StructureViewer` mounts only after a task succeeds and a structure result exists; the Three.js renderer is dynamically imported inside the component, and the WebGL context, `ResizeObserver`, geometries, and materials are released when leaving the page.
