# 前端 · 源码（frontend/src）

会被 Vite 编译的前端源码，用 `npm run dev` 启动（见上级 [frontend/README.md](../README.md)）。

| 文件 | 作用 |
| --- | --- |
| `main.ts` | 入口：挂载 Vue 应用，装上路由与全局样式 |
| `App.vue` | 应用外壳：顶栏（标题 / 导航 / 语言切换）+ 路由出口 |
| `router.ts` | 路由表：首页、预测页、结果页（hash 模式） |
| `styles.css` | 全站共用样式变量与基础类（页面容器、按钮、卡片、状态提示） |
| `formulaValidation.ts` | 化学式写法校验（纯函数）：字符、括号、下标；元素与电荷留给数据处理层 |
| `structureCif.ts` | 把“晶胞参数 + 分数坐标”拼成 3Dmol 能读的最小 CIF（P1） |
| `structureExplanation.ts` | 结果解释用的纯函数：晶胞形状判断、棱长与夹角格式化 |
| `3dmol-es6.d.ts` | 声明 3Dmol 的 ES 模块构建，复用包自带类型 |
| `api/` | 数据管理层接口封装（任务创建 / 查询 / 取消 / 结构结果），见 [api/README.md](./api/README.md) |
| `components/` | 页面共用组件，见 [components/README.md](./components/README.md) |
| `views/` | 页面组件，见 [views/README.md](./views/README.md) |
| `locales/` | 语言包与取词入口 `t()`，见 [locales/README.md](./locales/README.md) |

## 图标

界面图标用 Font Awesome（`@fortawesome/vue-fontawesome` + `@fortawesome/free-solid-svg-icons`）。

- 在用到图标的组件里按“单个图标”引入（`import { faXxx } from '@fortawesome/free-solid-svg-icons/faXxx'`），只有实际用到的图标会进打包产物。
- 图标只做装饰：旁边始终保留文字，并写 `aria-hidden="true"`，屏幕阅读器不会念出多余的图形。
- 图标颜色跟随文字（`currentColor`），按钮与链接用 `gap` 留出图标与文字的距离，不额外写 margin。

## 三维结构

结构结果用 [3Dmol.js](https://3dmol.org/) 渲染（BSD-3-Clause，版权声明见 [THIRD_PARTY_NOTICES.md](../THIRD_PARTY_NOTICES.md)）。

- 只在结果页成功、且确实要画图时才 `import('3dmol/build/3Dmol.es6-min.js')`：它被单独打成约 530 kB 的 chunk，首页与预测页不会下载。
- 用 ES 模块构建而不是 npm 入口的 UMD 构建，命名导出更稳定；类型通过 `3dmol-es6.d.ts` 转接。
- 渲染需要 WebGL；不支持时显示可翻译的提示，文字版结果（3.5）仍然可读。
- 3Dmol 内部使用 `eval`；如果以后部署时启用严格 CSP，需要允许 `unsafe-eval` 或换用 Three.js。

