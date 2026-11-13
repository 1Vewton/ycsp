> [English](./README.md) | 简体中文

# 前端 · 检查脚本（frontend/scripts）

给开发和 CI 用的独立脚本，不参与浏览器构建。

| 文件 | 作用 |
| --- | --- |
| `check-codes.mjs` | 检查 backend / compute 的每个 code 在 `src/locales/en.ts` 与 `zh.ts` 里都有文案；缺一条即失败 |
| `check-csp.mjs` | 构建后扫描 `dist/assets/*.js` 里的 `eval`、`new Function` 与 `Function()`；出现即失败 |

## 怎么用

```powershell
cd frontend
npm run check-codes

npm run build
npm run check-csp
```

## 为什么放在前端

错误码的文案由前端维护，但 code 的“源头”在另外两层。脚本直接读 `../backend/internal/codes/codes.go` 与 `../compute/app/constants/codes.py`，不手工维护第二份清单；CI 的 frontend 任务会运行它，新增 code 忘记补文案时会直接失败。

`check-csp.mjs` 读取的是生产构建产物，所以要先 `npm run build`。严格的内容安全策略（CSP）会拦截 `eval` 与 `Function` 构造器；三维视图不应依赖 `unsafe-eval`，CI 在构建后紧接着运行这项检查。
