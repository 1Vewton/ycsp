# 前端 · 检查脚本（frontend/scripts）

给开发和 CI 用的独立脚本，不参与浏览器构建。

| 文件 | 作用 |
| --- | --- |
| `check-codes.mjs` | 检查 backend / compute 的每个 code 在 `src/locales/en.ts` 与 `zh.ts` 里都有文案；缺一条即失败 |

## 怎么用

```powershell
cd frontend
npm run check-codes
```

## 为什么放在前端

错误码的文案由前端维护，但 code 的“源头”在另外两层。脚本直接读 `../backend/internal/codes/codes.go` 与 `../compute/app/constants/codes.py`，不手工维护第二份清单；CI 的 frontend 任务会运行它，新增 code 忘记补文案时会直接失败。
