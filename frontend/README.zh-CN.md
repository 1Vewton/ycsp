> [English](./README.md) | 简体中文

# 前端（frontend）

学生看到并操作的网页（TypeScript + Vue 3 + Vite），只与数据管理层通信。

## 职责

- 学生界面：输入化学式、看进度、看三维结构与解释、登录 / 注册。
- 只与数据管理层通信，不直接连数据处理层；计算与数据都留在后端。
- 不复制后端的元素与电荷数据：写法校验只做前端能独立判断的部分（见 `src/formulaValidation.ts`）。
- 面向学生的文字全部来自语言包，组件里不写死中文或英文。

## 常用命令

开发端口与后端地址都有默认值；要改的话，复制 `frontend/.env.example` 为 `frontend/.env`：

```powershell
Copy-Item .env.example .env
```

| 命令 | 作用 |
| --- | --- |
| `npm install` | 首次安装依赖 |
| `npm run dev` | 开发服务器（<http://127.0.0.1:5173>） |
| `npm run build` | 打包到 `dist/` |
| `npm run type-check` | TypeScript 类型检查 |
| `npm run check-codes` | 检查 backend / compute 的每个 code 都有中英文案（CI 也会跑） |
| `npm run check-csp` | `npm run build` 之后检查生产包中是否有 `eval` / `Function` 构造器（CI 也会跑） |
| `npm run test` | 跑一次自动化测试（纯函数 / 接口 / 组件，CI 也会跑） |
| `npm run test:watch` | 监听改动，持续跑测试 |

开发时 Vite 把 `/api` 代理到 `VITE_BACKEND_BASE_URL`（默认 `http://127.0.0.1:8000`，可在 `frontend/.env` 覆盖），所以用前请先启动 compute 与 backend。

自动化测试在 `tests/`，用 Vitest + jsdom，不需要启动后端，见 [tests/README.zh-CN.md](./tests/README.zh-CN.md)。

错误码文案检查脚本在 `scripts/`，见 [scripts/README.zh-CN.md](./scripts/README.zh-CN.md)。

## Docker（开发）

`Dockerfile` 用 Node 22 安装依赖，容器内运行 `npm run dev`；Vite 的 `/api` 代理指向 Compose 里的 `backend` 服务。启动方式与端口见根 README《Docker Compose（开发）》。

## 页面结构

七个页面用 vue-router 的 hash 模式（地址形如 `/#/predict`）：直接打开或刷新任何页面都不会 404，
部署时也不用给服务器加“回退到 index.html”的规则。地址登记在 `src/router.ts`。

| 路由 | 页面 | 内容 |
| --- | --- | --- |
| `/#/` | `HomeView` | 平台能做什么、三步上手、支持范围；后台服务连通检查 |
| `/#/predict` | `PredictView` | 化学式输入与即时写法反馈、示例快捷选择、开始预测 |
| `/#/result/:taskId` | `ResultView` | 轮询进度、取消排队任务、失败原因与耗时、三维结构与学生能懂的解释 |
| `/#/login` | `LoginView` | 登录 / 注册共用表单、可选安全问题、忘记密码入口；已登录时显示账号与退出 |
| `/#/reset-password` | `ResetPasswordView` | 按昵称取安全问题、回答后设置新密码；成功后直接登录 |
| `/#/account` | `AccountView` | 修改密码、设置或更新安全问题 |
| `/#/history` | `HistoryView` | 当前账号的预测记录，最新在前，可打开结果并“加载更多” |

每个页面现在做什么、下一步要接什么，见 [src/views/README.zh-CN.md](./src/views/README.zh-CN.md)。
登录不是预测的前提；不登录也能创建任务，只是任务不记归属，也不会出现在预测记录里。

## 与后端接口约定

| 约定 | 前端怎么做 |
| --- | --- |
| 唯一入口 | 只调用数据管理层的 `/api/...`；开发时由 Vite 代理，生产时走同源反向代理（见 [../backend/README.zh-CN.md](../backend/README.zh-CN.md)） |
| 统一信封 | 应答是 `{ service, status, code, params? }`，成功时另有 `task` / `result` / `user`；`status` 不是 `ok` 就按失败处理 |
| 错误文案 | 前端只拿语言无关的 `code` + `params`，用 `tCode()` 从 `src/locales/` 取当前语言的句子；后端不返回中文或英文 |
| 会话 | 登录 Cookie 由后端以 `HttpOnly` 下发，前端不读也不存；同源请求由浏览器自动携带 |
| 任务流程 | `POST /api/tasks` 立即返回 `pending` → 每秒轮询 `GET /api/tasks/:id` → 成功后取 `/result`；只有排队中的任务可取消 |
| 追踪 | 每次请求带 `X-Request-ID`，失败时 `ApiError.requestId` 保留同一个编号 |
| 内容安全 | 不把后端技术细节直接显示给学生；未知 code 显示 `unknown_code`，细节只进控制台 |

具体接口与字段见 [../backend/README.zh-CN.md](../backend/README.zh-CN.md)《对外接口》；前端封装见 [src/api/README.zh-CN.md](./src/api/README.zh-CN.md)。

## 日志追踪

每次 `/api` 请求都会带 `X-Request-ID`（`web-…`），控制台以
`[ycsp] request start / done / failed request_id=…` 记录；失败时 `ApiError.requestId` 保留同一个编号。
按这个编号可以查后端与数据处理层的日志，详见 [src/api/README.zh-CN.md](./src/api/README.zh-CN.md)
与 [../backend/README.zh-CN.md](../backend/README.zh-CN.md)《日志追踪》。

## 严格 CSP

三维视图用 [Three.js](https://threejs.org/) 渲染，不需要 `unsafe-eval`。
`npm run build` 之后执行 `npm run check-csp`，会扫描 `dist/assets/*.js` 中的
`eval` 与 `Function` 构造器；CI 在构建后也会跑这一步。

## 已知限制

- **支持范围有限**：离子晶体（1:1、1:2、2:1，含六方纤锌矿 ZnO / ZnS 与金红石型 MgF2 / TiO2）、金属单质与 C / Si / Ge / SiC 等共价晶体；多元素、2:3 计量、3 配位等会明确提示不支持（详见 [../compute/README.zh-CN.md](../compute/README.zh-CN.md)《已知限制》）。
- **三维视图需要 WebGL2**：不支持 WebGL2 的设备只显示文字解释；结构图是单个晶胞，不做对称扩展或周期性重复。
- **账号功能**：已有注册、登录 / 退出、修改密码、用安全问题重置密码和预测记录。自助重置要求事先设置安全问题（注册时或账号页）；匿名预测不会进入记录；修改 / 重置密码暂时不会让其他设备上的旧会话失效。
- **必须与数据管理层同源**：前端只请求相对路径 `/api`；如果静态站点与后端分域部署，需要反向代理或补 CORS，当前未做跨域支持。
- **只支持中 / 英两种语言**：其他语言会回退英文，不保证完整翻译。
- **不能离线使用**：预测、进度与结果都依赖后端；前端只负责展示。
- **前端测试不连真实后端**：`npm run test` 用 mock 的 `fetch` 覆盖纯函数、接口与组件；真实三层链路的回归仍靠手工或后续 E2E。

第三方图标与三维渲染库的许可见 [THIRD_PARTY_NOTICES.md](./THIRD_PARTY_NOTICES.md)。

界面文案按语言存放在 `src/locales/`，组件里只写 tag 并通过 `t()` 取词（见 [src/README.zh-CN.md](./src/README.zh-CN.md)）；默认中文，可切到英文，缺翻译时回退英文。
