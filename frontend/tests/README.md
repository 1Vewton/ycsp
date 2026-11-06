# 前端 · 自动化测试（frontend/tests）

用 [Vitest](https://vitest.dev/) 跑纯函数、接口封装与组件测试；组件跑在 jsdom 里，**不需要启动后端**。

## 怎么跑

```powershell
npm run test        # 跑一次（CI 用这个）
npm run test:watch  # 监听文件变化，边改边跑
```

配置在 [`../vitest.config.ts`](../vitest.config.ts)。

## 覆盖什么

| 文件 | 覆盖内容 |
| --- | --- |
| `formulaValidation.test.ts` | 写法校验：空值、长度、非法字符与位置、括号配对、0 下标 |
| `structureCif.test.ts` | 晶胞 → P1 CIF：参数格式、原子标签、坐标精度 |
| `structureExplanation.test.ts` | 晶胞形状判断（立方 / 六方 / 普通）与棱长、夹角格式化 |
| `locales.test.ts` | 默认中文、语言记忆与切换、占位符替换、未收录 code 的兜底 |
| `apiTasks.test.ts` | 任务接口封装：请求形状、错误信封 → `ApiError`、非 JSON 与断网 |
| `apiAuth.test.ts` | 账号接口：注册 / 登录 / 退出 / 当前账号，未登录返回 null、密码规则原样保留 |
| `requestId.test.ts` | 请求编号：`web-` 前缀、长度上限、每次不同、头名统一 |
| `session.test.ts` | 全站登录状态：并发合并、恢复、登录 / 注册 / 退出，退出失败保留账号 |
| `LoginView.test.ts` | 登录与注册表单：模式切换、校验、密码规则逐条提示、站内 redirect、已登录面板 |
| `App.test.ts` | 顶栏登录状态：登录入口、昵称与退出、退出失败提示 |
| `FormulaInput.test.ts` | 输入组件的三档反馈、v-model 与无障碍属性 |
| `structureExplanationComponent.test.ts` | 结果解释组件：来自后端的 rule / caveats、注意事项隐藏规则 |
| `ResultView.test.ts` | 结果页状态机：轮询、失败、取消、重试、结构结果单独失败 |

`fixtures.ts` 是共用样例（NaCl 结构与任务快照），不是测试文件。

## 约定

- 断言界面文字时用 `t()` / `tCode()` 取语言包，不把中文或英文写死：换语言或改文案都不用改测试。
- 结果页用假时钟（`vi.useFakeTimers`）推进 1 秒轮询，测试不真的等待。
- 三维查看器在组件测试里用 `stubs: { StructureViewer: true }` 替掉：3Dmol 需要 WebGL，单元测试只验证“结果有没有交给查看器”。
- 不连真实后端：`fetch` 用 `vi.stubGlobal` 替换，失败路径靠伪造应答覆盖。
- 文件名不要只靠大小写区分（Windows 与 macOS 默认不区分）：组件测试用 `structureExplanationComponent.test.ts`。
