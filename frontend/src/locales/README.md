# 前端 · 语言包（frontend/src/locales）

界面文案按语言分开存放，组件里只写 tag（例如 `page_title`），不写死任何一种语言的文字（见根 README「多语言」）。

| 文件 | 作用 |
| --- | --- |
| `en.ts` | 英文文案；同时是“有哪些 tag”的基准 |
| `zh.ts` | 中文文案（默认语言） |
| `index.ts` | 当前语言状态、`t()` 取词与占位符替换、切换与记忆 |

## 怎么用

```ts
import { currentLocale, setLocale, t } from './locales'

t('page_title')                        // 取当前语言的文案
t('task_not_found', { task_id: 7 })    // 把 {task_id} 换成 7
setLocale('en')                        // 切换并记住选择
```

## 规则

- 新增文案必须同时写进 `en.ts` 与 `zh.ts`：`zh.ts` 的键按 `en.ts` 检查，少写、写错或多写都会让 `npm run type-check` 失败。
- 查找顺序是“当前语言 → 英文 → 原样显示 tag”，第三种情况还会 `console.warn`，缺翻译不会被悄悄藏起来。
- 要填进句子的数值用 `{name}` 占位符，通过 `t()` 的第二个参数传入，不要自己拼字符串。
- `中文` / `English` 是语言自己的称呼，不随界面语言变化，所以放在 `index.ts` 的 `LOCALE_NAMES` 里，不进语言包。
- 切换后写回 `localStorage`（键 `ycsp_locale`），并同步 `<html lang>`，刷新或下次打开仍是同一语言。
