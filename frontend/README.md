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
