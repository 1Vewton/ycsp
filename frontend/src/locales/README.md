> English | [简体中文](./README.zh-CN.md)

# Frontend · Language packs (frontend/src/locales)

Interface copy is stored per language; components write only tags (e.g. `page_title`) and never hard-code text in any language (see the root README "Languages").

| File | Purpose |
| --- | --- |
| `en.ts` | English copy; also the baseline for "which tags exist" |
| `zh.ts` | Chinese copy (default language) |
| `index.ts` | Current language state, `t()` lookup and placeholder replacement, switching and persistence |

## How to use

```ts
import { currentLocale, setLocale, t, tCode } from './locales'

t('page_title')                        // copy in the current language
t('task_not_found', { task_id: 7 })    // replace {task_id} with 7
tCode('task_not_found', { task_id: 7 }) // code returned by the backend; unknown ones show unknown_code
setLocale('en')                        // switch and remember the choice
```

## Rules

- New copy must be added to both `en.ts` and `zh.ts`: `zh.ts` keys are checked against `en.ts`, so missing, misspelled, or extra keys fail `npm run type-check`.
- Codes returned by the backend / data processing layer use `tCode()`: unknown codes show `unknown_code` (with the original code), so internal identifiers such as `task_future_code` are never shown to students; `t()` is still for the interface's own copy.
- Lookup order is "current language → English → raw tag"; the third case also triggers `console.warn`, so missing translations are never hidden.
- Numbers that belong in a sentence use `{name}` placeholders passed as `t()`'s second argument; do not concatenate strings manually.
- `中文` / `English` are the languages' own names and do not change with the interface language, so they live in `LOCALE_NAMES` in `index.ts`, not in the language packs.
- After switching, the choice is written to `localStorage` (key `ycsp_locale`) and `<html lang>` is updated, so a refresh or the next visit keeps the same language.
