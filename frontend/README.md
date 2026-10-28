# 前端（frontend）

学生看到并操作的网页（TypeScript + Vue 3 + Vite），只与数据管理层通信。

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

开发时 Vite 把 `/api` 代理到 `VITE_BACKEND_BASE_URL`（默认 `http://127.0.0.1:8000`，可在 `frontend/.env` 覆盖），所以用前请先启动 compute 与 backend。

## 页面与路由

三个页面用 vue-router 的 hash 模式（地址形如 `/#/predict`）：直接打开或刷新任何页面都不会 404，部署时也不用给服务器加“回退到 index.html”的规则。

地址登记在 `src/router.ts`；每个页面现在做什么、下一步要接什么，见 [src/views/README.md](./src/views/README.md)。首页（`/#/`）介绍平台能力与支持范围，并提供后台服务连通检查；预测页可以输入化学式（或点示例）并开始预测；结果页自动刷新任务进度（排队中 → 计算中 → 已完成 / 失败），排队中的任务可以取消；结构展示与解释由 3.4 / 3.5 接入。

界面文案按语言存放在 `src/locales/`，组件里只写 tag 并通过 `t()` 取词（见 [src/README.md](./src/README.md)）；默认中文，可切到英文，缺翻译时回退英文。
