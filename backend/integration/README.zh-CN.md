> [English](./README.md) | 简体中文

# 数据管理层 · 跨层集成测试（backend/integration）

带构建标签的测试：把**真实数据处理层**和 Go 任务服务一起跑起来。
它对应任务 4.3：Python 拒绝一个化学式后，Go 任务必须保存同一个 `code` + `params`，前端语言包里要有这个 code 的中文解释。

## 检查什么

- `POST /api/tasks` 把 `NaCl2`、`Xx2`、`H2O` 真正发给 Python 的 `/predict`。
- 任务最终 `failed`，错误码分别是 `formula_charge_unbalanced`、`element_not_supported`、`composition_not_supported`。
- 失败信息保留上游的 `params`（formula / symbol / position）。
- `frontend/src/locales/zh.ts` 有这个 key 和对应的中文句子；`en.ts` 有这个 key。
- 结果页把这个 code 渲染成中文提示，由 `frontend/tests/ResultView.test.ts` 覆盖。

## 本地运行

一个终端启动数据处理层：

```powershell
cd compute
uv run uvicorn app.main:app --host 127.0.0.1 --port 8001
```

另一个终端跑带标签的测试：

```powershell
cd backend
$env:YCSP_TEST_COMPUTE_URL = "http://127.0.0.1:8001"
go test -tags=integration ./integration/...
```

不设 `YCSP_TEST_COMPUTE_URL` 时测试会跳过；普通 `go test ./...` 不会编译这个目录，因为它需要 `integration` 构建标签。

CI 用一个独立 job 先启动 compute，再跑这个带标签的测试。
