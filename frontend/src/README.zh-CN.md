> [English](./README.md) | 简体中文

# 前端 · 源码（frontend/src）

会被 Vite 编译的前端源码，用 `npm run dev` 启动（见上级 [frontend/README.zh-CN.md](../README.zh-CN.md)）。

| 文件 | 作用 |
| --- | --- |
| `main.ts` | 入口：挂载 Vue 应用，装上路由与全局样式 |
| `App.vue` | 应用外壳：顶栏（标题 / 导航 / 登录状态 / 语言切换）+ 路由出口 |
| `router.ts` | 路由表：首页、预测页、结果页、登录、找回密码、账号设置、预测记录（hash 模式） |
| `session.ts` | 全站登录状态：当前账号、恢复 / 登录 / 注册 / 退出、修改密码、更新安全问题、重置密码，供顶栏与账号相关页面共用 |
| `styles.css` | 全站共用样式变量与基础类（页面容器、按钮、卡片、状态提示） |
| `formulaValidation.ts` | 化学式写法校验（纯函数）：字符、括号、下标；元素与电荷留给数据处理层 |
| `structureScene.ts` | 三维视图用的纯几何：晶胞棱向量、分数坐标 → 直角坐标、化学键、晶胞棱、取景半径 |
| `structureRenderer.ts` | Three.js 渲染器（按需加载）：原子、化学键、晶胞棱框、灯光、重置视角、资源释放 |
| `structureExplanation.ts` | 结果解释用的纯函数：晶胞形状判断、棱长与夹角格式化 |
| `api/` | 数据管理层接口封装（任务、账号、请求编号与统一错误），见 [api/README.zh-CN.md](./api/README.zh-CN.md) |
| `components/` | 页面共用组件，见 [components/README.zh-CN.md](./components/README.zh-CN.md) |
| `views/` | 页面组件，见 [views/README.zh-CN.md](./views/README.zh-CN.md) |
| `locales/` | 语言包与取词入口 `t()`，见 [locales/README.zh-CN.md](./locales/README.zh-CN.md) |

## 图标

界面图标用 Font Awesome（`@fortawesome/vue-fontawesome` + `@fortawesome/free-solid-svg-icons`）。

- 在用到图标的组件里按“单个图标”引入（`import { faXxx } from '@fortawesome/free-solid-svg-icons/faXxx'`），只有实际用到的图标会进打包产物。
- 图标只做装饰：旁边始终保留文字，并写 `aria-hidden="true"`，屏幕阅读器不会念出多余的图形。
- 图标颜色跟随文字（`currentColor`），按钮与链接用 `gap` 留出图标与文字的距离，不额外写 margin。

## 三维结构

结构结果用 [Three.js](https://threejs.org/) 渲染（MIT，版权声明见 [THIRD_PARTY_NOTICES.md](../THIRD_PARTY_NOTICES.md)）。

- `structureScene.ts` 是纯函数模块：把晶胞参数与分数坐标换成原子位置、化学键、晶胞棱与取景半径，单测在 `tests/structureScene.test.ts`。
- `StructureViewer.vue` 只在结果页成功后才动态加载 `structureRenderer.ts`；Three.js 被单独打成约 560 kB 的 chunk，首页与预测页不会下载。
- 化学键取第一配位壳层：距离不超过最短原子间距的 1.2 倍；离子晶体只连接不同种离子，金属晶体不画键。
- 渲染需要 WebGL；不支持时显示可翻译的提示，文字版结果仍然可读。
- 渲染器不使用 `eval` / `Function` 构造器，严格 CSP 不需要 `unsafe-eval`；`npm run build` 后由 `npm run check-csp` 检查构建产物。
