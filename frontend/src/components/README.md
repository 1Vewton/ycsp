# 前端 · 组件（frontend/src/components）

页面共用的小部件；页面本身在 [`../views/`](../views/README.md)。

| 文件 | 作用 |
| --- | --- |
| `FormulaInput.vue` | 化学式输入框：标签、输入、即时反馈（怎么写 / 写法正确 / 第几个字符看不懂） |
| `StructureViewer.vue` | 晶体结构三维示意图：3Dmol 按需加载、晶胞棱框、重置视角；WebGL 不可用时给出提示 |

## 约定

- 组件只负责显示与交互，不判断化学式对不对：写法校验在 [`../formulaValidation.ts`](../formulaValidation.ts)，页面把结果作为 `problem` 传进来。
- 文字一律通过 [`../locales`](../locales/README.md) 的 `t()` 取；颜色等样式用 [`../styles.css`](../styles.css) 里的变量。
- `StructureViewer` 只在任务成功、拿到结构结果后挂载；3Dmol 在组件内部动态加载，离开页面时清空 viewer 与 `ResizeObserver`。
