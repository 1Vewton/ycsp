# TASKS.md —— ycsp 项目开发计划

> **这份文件是干什么的**
> 记录“项目要做哪些事、按什么顺序做、每件事拆到哪一步”，方便逐项打卡推进。
> 这里**只写计划**；每个模块具体怎么实现、怎么用，写在各自目录的 `README.md` 里。
> 项目定位与技术栈见根 [`README.md`](./README.md)。

## 怎么用这份计划

- **阶段**按依赖先后排列：上一个阶段的产出，是下一个阶段的输入。
- 每个阶段先写 **目标** 与 **验收标准**，再拆成一组 **子任务**，子任务逐条可勾选。
- 子任务遵守“三自”：**自己能独立做完、自己能独立验证、做完别人能看懂**。
- 勾选标记：

  | 标记 | 含义 |
  | --- | --- |
  | `[ ]` | 未开始 |
  | `[~]` | 进行中 |
  | `[x]` | 已完成 |

- 完成一个子任务就把它勾上，并在文末《进度记录》追加一行（日期 + 做了什么）。

## 里程碑总览

| 阶段 | 名称 | 主要产出 | 工作量（粗估） | 依赖 |
| --- | --- | --- | --- | --- |
| 0 | 项目骨架与约定 | 三层可跑的空壳 + 层间接口约定 | 小 | 无 |
| 1 | 数据处理层 | 能算晶体结构的 Python 服务 | 大 | 阶段 0 |
| 2 | 数据管理层 | 统一对外接口 + 数据落库 | 中 | 阶段 0、1 |
| 3 | 前端 | 学生能操作的网页 + 结构可视化 | 中 | 阶段 0、2 |
| 4 | 端到端集成 | 从“输入”到“看到结果”的完整链路 | 中 | 阶段 1、2、3 |
| 5 | 教学化打磨与验收 | 中学生能独立走通的成品 | 中 | 阶段 4 |

> 阶段 1 与阶段 3 的界面设计可以部分并行；真正锁死顺序的是阶段 0 → 1 → 2 → 4。

## 全局约定（动手前先定，避免返工）

- [x] **G0.1 确认三层目录命名**：已按暂拟名 `frontend/`、`backend/`、`compute/` 落地；若要改名，需同步各 README 与根 README 导航
- [x] **G0.2 确认数据库**：已定——开发/测试用 SQLite（零安装），生产用 PostgreSQL；连接层可切换（`YCSP_DB_DRIVER` / `YCSP_DB_DSN`），生产环境拒绝 SQLite
- [x] **G0.3 确认 CSP 教学范围**：已定——离子晶体（半径比规则）+ 金属晶体（实验结构表）+ 金刚石型共价晶体，详见 `compute/README.md`《算法思路》
- [x] **G0.4 定义层间接口**：统一 JSON；应答为 `{ service, status, code, params? }`，后端只给语言无关的 `code`，文案由前端按语言翻译；端口 8000/8001/5173
- [x] **G0.5 定义命名与代码规范**：已按 `AGENTS.md` 执行（完整单词命名、面向用户文案全中文、术语首次出现给通俗解释）
- [x] **G0.6 定义三层启动方式**：已在各目录 `README.md` 与根 README《安装与启动》写明

> 下列阶段中子任务里出现的目录名（`frontend/`、`backend/`、`compute/`）均为**暂拟**，以 G0.1 确认结果为准。

## 阶段 0：项目骨架与约定

**目标**：仓库里出现三条清晰骨架，每条都能“跑起来”，并能互相打通一个空请求。

**验收标准**：在浏览器点一下，能收到来自 Python 层返回的一句中文（证明 Vue → Go → Python 链路通畅）。

- [x] **0.1** 前端：Vue 3 + TypeScript + Vite 最小项目，页面含“打招呼”按钮，已建 `frontend/README.md`
- [x] **0.2** 数据管理：Go + Gin 项目，`GET /api/health` 返回服务状态，已建 `backend/README.md`
- [x] **0.3** 数据处理：Python + FastAPI 项目，`GET /health` 返回服务状态，已建 `compute/README.md`
- [x] **0.4** 数据管理：`GET /api/ping-compute`，由 Go 转发请求到 Python 层
- [x] **0.5** 前端：按钮调用 `/api/ping-compute`，在页面上展示返回文字
- [x] **0.6** 回填根 `README.md` 的《安装与启动》《目录导航》两节
- [x] **0.7** 三层各有一条启动命令，写到各自 `README.md`

## 阶段 1：数据处理层（Python + FastAPI + NumPy）

**目标**：把“晶体结构预测”的核心计算做出来，先用命令行跑通，再包成接口。

**验收标准**：给定课本里的经典例子（如 NaCl），程序输出的结构与课本一致，并有单元测试证明。

- [x] **1.1** 定义内部数据结构：原子、元素、晶胞（`app/domain/structure.py`，字段注明物理量与单位）
- [x] **1.2** 内置常见元素数据：离子半径、电负性等（`app/domain/elements.py`，注明数据来源与单位）
- [x] **1.3** 实现最小教学算法 v1：离子晶体用 Pauling 半径比规则（4/6/8 配位）与萤石型，金属按实验结构表，C/Si/Ge 与 SiC 用金刚石型 / 闪锌矿型（`app/domain/prediction.py`、`prototypes.py`、`composition.py`）
- [x] **1.4** 用课本经典例子写单元测试：NaCl / CsCl / CaF₂ / ZnS / 金刚石 / SiC / Cu / Fe / Mg（`compute/tests/`，`uv run pytest`）
- [x] **1.5** 结果后处理：`app/api/serialization.py` 把晶胞转成“元素符号 + 分数坐标”列表，前端可直接画图
- [x] **1.6** 包成 FastAPI 接口 `POST /predict`，返回结构数据与可直接画图的原子列表
- [x] **1.7** 长耗时处理：本阶段计算是纯算术（毫秒级），不需要进度与取消入口，已在 `compute/README.md`《已知限制》写明
- [x] **1.8** 错误处理：计算与框架异常都转成语言无关的 `code` + `params`，文案由前端翻译（`app/constants/codes.py`、`app/api/errors.py`）
- [x] **1.9** 写 `compute/README.md`：职责、接口、算法思路、使用示例、已知限制
- [x] **1.10** 教学对齐：半径比规则、硬球相切模型、萤石型等概念在注释与 README 中点明对应课本概念

## 阶段 2：数据管理层（Go + Gin + GORM）

**目标**：成为对外统一接口，负责数据落库，并转发计算请求。

**验收标准**：能创建一次“预测任务”、存进数据库、再查回来，且状态正确。

- [x] **2.1** 设计数据库表：用户、任务、结构结果（字段、约束、状态机与关系已落到 `backend/README.md`《数据表》）
- [x] **2.2** 用 GORM 写模型并做自动迁移（`internal/store/`：三张表模型、引擎可切换的连接层、可重复执行的 `Migrate`）
- [x] **2.3** 实现任务接口：创建任务、查询状态、获取结果、取消任务（`POST /api/tasks` 等四个；异步执行、前端轮询）
- [x] **2.4** 对接数据处理层：把任务参数发给 Python，接收结果并落库（`internal/compute` 客户端 + `internal/tasks` worker）
- [x] **2.5** 统一错误码表：后端只定义 `code` + `params`，文案翻译放在前端（`internal/codes/codes.go`，前端文案见 `frontend/src/locales/`）
- [x] **2.6** 输出接口文档，并登记到 `backend/README.md`（《对外接口》一节）
- [x] **2.7** 写 `backend/README.md`：职责、对外接口、数据表、示例、限制

## 阶段 3：前端（Vue + TypeScript）

**目标**：做出学生能看懂、能操作的界面。

**验收标准**：一个没学过化学、也没学过编程的中学生，不看说明书就能自己完成一次“预测”。

- [x] **3.1** 页面骨架：首页（平台能干什么）、预测页、结果页
- [x] **3.2** 输入组件：让学生输入 / 选择化学组成（带即时校验与中文提示）
- [x] **3.3** 调用数据管理层接口，实时展示任务进度
- [x] **3.4** 晶体结构三维可视化（现用 Three.js；结果页按需加载，WebGL 不可用时保留文字说明）
- [x] **3.5** 结果解释区：把结构类型、晶胞参数用中学生能懂的话讲清楚
- [x] **3.6** 错误提示：后端每个 `code` 都有对应文案，并按当前语言显示（`npm run check-codes` 在 CI 中守护；未知 code 用 `unknown_code` 兜底）
- [x] **3.7** 写 `frontend/README.md`：职责、页面结构、与后端接口约定、限制
- [x] **3.8** 语言切换：界面可切换中/英（`frontend/src/locales/` 分开存放 `zh.ts` / `en.ts`，`t(tag, params)` 取词、缺翻译回退英文、选择记在 `localStorage`）

## 阶段 4：端到端集成

**目标**：把三条链路合成一个完整产品。

**验收标准**：从界面输入到看到结果，全流程无手工干预、无英文报错。

- [x] **4.1** 在真实数据上跑通“输入 → 计算 → 存储 → 展示”完整流程
- [x] **4.2** 处理并发与超时（多人同时用、计算卡住时给出中文提示）
- [ ] **4.3** 跨层错误传播测试：Python 报错，前端要显示中文解释
- [ ] **4.4** 用课本经典例子做端到端回归测试

## 阶段 5：教学化打磨与验收

**目标**：达到唯一验收标准——“中学生能否独立看懂并操作”。

**验收标准**：至少 1 名真实中学生独立完成一次完整操作，全程无需解释。

- [ ] **5.1** 术语审查：界面每个术语首次出现都有通俗解释
- [ ] **5.2** 文案审查：全中文、无黑箱描述、无未解释的等待
- [ ] **5.3** 请真实中学生试用，记录卡点并逐条修复
- [ ] **5.4** 文档收尾：各 README 与根 README 齐全，确认“文档与代码一致”
- [ ] **5.5** 部署上线（**部署方式待确认** Q5）

## 待确认事项（需要你拍板，直接影响上面的子任务）

| 编号 | 问题 | 为什么重要 | 影响任务 |
| --- | --- | --- | --- |
| Q1 | 三层目录到底叫什么名字？ | 定不下来没法建目录、没法写导航 | G0.1、阶段 0 |
| Q2 | 数据库用哪个？ | 影响 GORM 配置与部署方式 | G0.2、2.1 |
| Q3 | CSP 教学覆盖哪些内容、学生做什么？（已定：离子 + 金属 + 金刚石型共价，见《进度记录》2026-10-20） | 决定算法与界面，是最大变量 | G0.3、1.3、阶段 3 |
| Q4 | 晶体结构怎么可视化、用什么库？（现用 Three.js，结果页按需加载；严格 CSP 下不需要 unsafe-eval，见《进度记录》2026-11-09） | 直接影响前端工作量 | 3.4 |
| Q5 | 部署到哪里（本地 / 服务器 / 校园网）？ | 影响阶段 5 的做法 | 5.5 |

## 进度记录

| 日期 | 阶段 | 完成内容 | 备注 |
| --- | --- | --- | --- |
| 2026-10-13 | — | 建立本计划文件 | 待确认 Q1~Q5 |
| 2026-10-13 | 0 | 搭建三层骨架（Vue / Go / Python）并打通 `/api/ping-compute` | 已实测：`compute /health` ✓、`backend /api/health` ✓、经 Vite 代理的 `/api/ping-compute` ✓（HTTP 200）。目录名沿用暂拟名 |
| 2026-10-13 | — | 补 4 个代码子目录的 README；`AGENTS.md` 新增“父 README 不复制子模块文档、只登记一行导航”的规则 | 应作者要求：减少重复、节省 token |
| 2026-10-14 | — | 精简文档与代码注释（按新 `AGENTS.md` 标准）；数据处理层改用 uv 管理依赖（`pyproject.toml` + `uv.lock` 取代 `requirements.txt`） | 顺带修掉 `vue-tsc` 与 TypeScript 7 不兼容（`typescript` 固定 `^5.9.3`） |
| 2026-10-14 | 1 | 完成 1.1：数据处理层内部数据结构 `compute/app/domain/structure.py`（Element / Atom / UnitCell） | 长度 Å、角度 °；含基本合法性校验；按多语言规则不放展示文案 |
| 2026-10-16 | — | 去掉后端中文硬编码：应答改为 `{ service, status, code, params? }`（`backend/internal/api/codes.go`）；前端按 `code` 显示，文案抽到 `frontend/src/messages.ts` | 中/英文案已备齐；语言切换留到阶段 3（文案现已拆到 `frontend/src/locales/`） |
| 2026-10-17 | — | 三层网络配置外置到 `.env`（各层 `/.env.example` + `.env`），并各有一个带默认值的 env 管理器（`compute/app/env`、`backend/internal/env`、`frontend/env.ts`） | 服务名、状态码等协议常量仍留在代码里 |
| 2026-10-17 | — | 新增 GitHub Actions CI（`.github/workflows/ci.yml`）：三层并行检查（uv 依赖与导入 / Go 格式与构建 / 前端类型检查与构建） | 说明见 `.github/README-workflow.md`，并登记到根 README 导航 |
| 2026-10-19 | 1 | 完成 1.2：`app/domain/elements.py` 内置 19 个常见元素（Pauling 电负性）与 13 个常见离子半径 | 离子半径取 Shannon 1976 有效值、配位数 6；含 `KeyError` 提示 |
| 2026-10-20 | 1 | 确定教学范围（Q3/G0.3）：离子晶体 + 金属晶体 + 金刚石型共价晶体一起做；金属结构型按经验表（用户选 A） | 理由：三类共用一套“判类型 → 选原型 → 算晶胞”管线，比各做一套省成本 |
| 2026-10-20 | 1 | 完成 1.3 / 1.5 / 1.6：`composition.py`、`prototypes.py`、`prediction.py` + `POST /predict` | 覆盖 NaCl/CsCl/ZnS/萤石（含反萤石）、fcc/bcc/hcp、金刚石/SiC；元素数据扩到 31 个，补共价半径与金属半径 |
| 2026-10-20 | 1 | 完成 1.4：`compute/tests/` 四个测试文件、56 个用例（课本经典例子 + 数据一致性 + 接口契约） | `uv run pytest` 全绿；dev 依赖 pytest/httpx，`uv.lock` 已更新；CI 增加测试步骤 |
| 2026-10-20 | 1 | 完成 1.8 / 1.9 / 1.10：错误全部走 `code` + `params`；README 写明接口、算法与已知限制；前端补齐中/英文案 | KCl 等规则判错的经典化合物改用实验数据，并在结果里说明差异 |
| 2026-10-20 | 1 | 重构：数据与常数拆到 `app/constants/`（`elements` / `structures` / `prediction_rules` / `codes`），`app/domain/` 只留类型与规则 | 常数包不依赖 domain/api，`tests/test_constants.py` 守住这条线；57 个用例全绿 |
| 2026-10-20 | 1 | 补测：新增 `test_structure` / `test_prototypes` / `test_serialization` / `test_env` / `test_api_errors`，用例 57 → 122，覆盖率 99% | 引入 pytest-cov；CI 以 `--cov-fail-under=90` 兜底 |
| 2026-10-20 | 2 | 完成 2.1：设计 `users` / `prediction_tasks` / `structure_results` 三张表，含状态机与接口对应关系，写入 `backend/README.md`《数据表》 | 待定：数据库引擎（G0.2）、是否引入账号体系 |
| 2026-10-21 | 2 | 完成 2.2：`internal/store/` 三张表模型 + `Open`（sqlite / postgres / mysql 可切换）+ `Migrate`，启动时自动建表 | 9 个 Go 测试覆盖迁移、存取、级联删除与配置校验；默认 SQLite 文件 `ycsp.db` |
| 2026-10-21 | 2 | 补 Go 测试：`api`（7，含假数据处理层的五种应答）、`env`（7）、`main` 配置解析（4）、`store`（13） | 共 31 个测试函数；覆盖率 api/env 100%、store 92.9%、main 29.6%（仅启动入口未测）；CI 改为 `go test -cover ./...` |
| 2026-10-21 | 2 | 完成 2.3 + 2.4：任务接口（创建 / 查询 / 结果 / 取消）+ 异步 worker + 数据处理层客户端，结果与错误按 code 落库 | 端到端实测：`CaF2` → pending → succeeded（133 ms）→ fluorite 5.381 Å / 12 原子；`H2O` → failed（`composition_not_supported`） |
| 2026-10-23 | 2 | 按用户要求给 `users` 加 `password_hash`：bcrypt 加盐哈希（`internal/auth`），不存明文；旧库自动加列（默认空串 = 未设密码，不能登录） | 7 个 auth 测试 + 2 个迁移/落库测试；`golang.org/x/crypto` 变为直接依赖 |
| 2026-10-23 | 2 | 新增注册/登录：`POST /api/users`、`POST /api/sessions`、`DELETE /api/sessions`、`GET /api/users/me`；会话存 Redis（可切 `memory`），Cookie `HttpOnly`+`SameSite=Lax`；密码策略 8 位 + 字母 + 数字 + 特殊符号；任务记录登录用户 | 新增 `auth`（策略）、`session`、`users`、`testutil` 包；测试用 `miniredis`，CI 无需 Redis 服务 |
| 2026-10-23 | 2 | 新增登录失败限流：按昵称 5 次 / 按地址 20 次 / 15 分钟窗口（可配），超限返回 429 `too_many_attempts` + `Retry-After`；成功登录只清昵称计数 | 新增 `internal/ratelimit`（Redis + 内存两种计数器）；按直连 IP 统计，不信任 `X-Forwarded-For` |
| 2026-10-24 | 2 | 补齐 2.5 / 2.6 / 2.7 的勾选：统一 code 表（`internal/codes`）、对外接口文档与 `backend/README.md` 已随账号体系落地 | 勾选前逐项对照代码与 README 确认 |
| 2026-10-24 | — | 修文档与代码不一致：`.github` 的 CI 文档改名为 `README-workflow.md`（GitHub 会把 `README.md` 当成目录说明渲染，不该顶替）；`api/README` 去掉不存在的 `display.go` 并补 `auth.go`；`auth/README` 补 `policy.go`；后端覆盖率说明改为实测值；根 README 的连通提示改为实际文案 | 实测：compute 122 通过 / 99%，backend 全包通过，frontend 类型检查通过 |
| 2026-10-24 | 3 | 完成 3.8 的多语言底座：语言包拆到 `frontend/src/locales/`（`zh.ts` / `en.ts`），组件只写 tag、用 `t(tag, params)` 取词；当前语言缺文案回退英文，两边都缺则显示 tag 并 `console.warn`；切换写入 `localStorage` 并同步 `<html lang>` 与标签页标题 | `zh.ts` 的键按 `en.ts` 校验，中英缺一即 `npm run type-check` 失败；原 `messages.ts` 删除 |
| 2026-10-25 | 3 | 完成 3.1：页面骨架——引入 vue-router（hash 模式）与 `src/views/` 的三个页面（首页 / 预测页 / 结果页），`App.vue` 改为顶栏（标题 / 导航 / 语言切换）+ 路由出口；首页写清“能做什么、三步上手、支持范围”，阶段 0 的三层连通检查移到首页，并把“数据处理层”等架构词改成学生能看懂的说法 | 预测页输入框（3.2）、任务进度（3.3）、结构展示（3.4 / 3.5）留待后续；中英文案同步补齐，`src/styles.css` 收敛共用样式 |
| 2026-10-25 | 3 | 完成 3.2：化学式输入组件——`components/FormulaInput.vue`（输入 + 三档即时反馈：怎么写 / 写法正确 / 第几个字符看不懂）+ `formulaValidation.ts` 纯函数校验（字符、括号、下标、长度，规则对齐数据处理层的解析器），预测页加三种晶体类型的示例按钮，点一下直接填入 | 元素是否收录、电荷是否守恒仍由数据处理层判断（避免前端复制元素数据造成漂移）；“开始预测”与进度留到 3.3 |
| 2026-10-25 | 3 | 完成 3.3：预测页提交任务（`src/api/tasks.ts` 封装创建 / 查询 / 取消，统一抛 `ApiError`），结果页每秒轮询并用“排队中 → 计算中 → 已完成”三步展示进度；支持取消排队中的任务、显示失败原因与耗时、断网后重试 | 结构展示（3.4 / 3.5）仍是占位；任务快照字段保持后端 JSON 命名，联调确认创建 → 轮询 → 成功 / 失败全流程 |
| 2026-10-25 | 3 | 界面美化：接入 Font Awesome（官方 Vue 组件 + 按需引入 SVG 图标）——顶栏品牌与导航、首页三步卡片与按钮、预测页按钮与返回、结果页状态徽标与取消 / 重试按钮；图标一律 `aria-hidden` 且旁边保留文字，系统开启“减少动态效果”时停止旋转 | 只打包实际用到的图标（检查产物：用到的图标在、未用的不在）；JS 体积 114.8 → 232.0 kB（gzip 44.0 → 75.1 kB），主要来自 Font Awesome 的 SVG 渲染内核 |
| 2026-10-25 | 3 | 完成 Q4 + 3.4：可视化选 3Dmol.js（BSD-3、自带晶胞与元素配色）；结果页在任务成功后取 `/api/tasks/:id/result`，由 `structureCif.ts` 把晶胞参数 + 分数坐标拼成 P1 最小 CIF，3Dmol 再按需加载（独立 chunk ≈ 533 kB）画球棍与晶胞棱框，提供重置视角、拖动/滚轮提示、WebGL 不可用与加载失败提示 | CIF 链路用 jsdom + 3Dmol 实测通过（NaCl 2 原子、a = 5.66 Å）；首页与预测页主包只增 6.6 kB；3Dmol 内部使用 `eval`，严格 CSP 部署需注意 |
| 2026-10-25 | 3 | 完成 3.5：结果解释区——`structureExplanation.ts` 判断晶胞形状（立方 / 六方 / 普通）并格式化棱长与夹角，`StructureExplanation.vue` 展示晶体类型、结构类型、配位数（含术语解释）、晶胞参数、判断依据与注意事项；顺带把“面心立方 / 四面体空隙 / 电负性”等术语补了通俗解释 | 解释内容全部来自后端 rule / caveats，不在前端重写结论；无注意事项时整块隐藏 |
| 2026-10-25 | 3 | 完成 3.6：错误提示——审计 backend 29 + compute 26 个 code，全部有中英文案；新增 `tCode()`（未收录 code 显示 `unknown_code`，不把内部编号丢给学生）并接入首页 / 预测页 / 结果页；新增 `frontend/scripts/check-codes.mjs` 与 `npm run check-codes`，CI 在新增 code 漏翻时直接失败；真实跑通 `formula_charge_unbalanced`、`element_not_supported`、`composition_not_supported`、`task_not_found`、`task_formula_empty` / `task_formula_too_long`、`compute_unreachable` | 学生可见提示里去掉“数据管理层 / 数据处理层”等架构词；三维错误补 `role="alert"`、加载中补 `role="status"`；检查脚本在“故意漏翻”的镜像里实测会退出 1 |
| 2026-11-02 | 3 | 补前端自动化测试：接入 Vitest + @vue/test-utils + jsdom，覆盖写法校验、CIF 拼接、结果解释、语言包、任务接口，以及输入框与结果页（轮询 / 取消 / 重试） | `npm run test` 共 62 个用例；不依赖真实后端；CI 增加测试步骤；说明见 `frontend/tests/README.md` |
| 2026-11-02 | 3 | 补前端登录界面：登录与注册共用 `LoginView`，顶栏显示登录入口或昵称 + 退出；新增 `src/api/client.ts`（与任务共用信封 / 错误）、`src/api/auth.ts`、`src/session.ts`；登录后回跳站内 redirect，站外地址忽略 | 表单错误全部走 code + 中英文案；密码规则逐条显示；新增 30 个测试（账号接口 8、会话 7、登录页 11、顶栏 3、密码规则文案 1），前端共 92 个用例 |
| 2026-11-02 | 2 | 修复 nginx 反向代理下限流把所有人当成同一 IP：新增 `YCSP_TRUSTED_PROXIES`（默认空，只信任配置的代理），从 `X-Forwarded-For` 右往左取第一个不受信地址；拒绝 `0.0.0.0/0`，头部异常退回直连地址 | 新增 16 个测试（地址解析 11、限流集成 2、配置解析 3）；backend README 增加 nginx 部署示例 |
| 2026-11-02 | — | 三层日志追踪：统一 `X-Request-ID`；前端每次请求生成并记录，数据管理层中间件沿用/生成并写回，数据处理层沿用/生成并记录；异步任务把编号存进 `prediction_tasks.trace_id`，worker 调计算层时继续使用 | 新增 18 个测试（后端 10、计算层 5、前端 3）；日志不记请求体与密码；按 request_id 可串起三层 |
| 2026-11-02 | 3 | 完成 3.7：`frontend/README.md` 补齐职责、页面结构（4 个路由表）、与后端接口约定（统一信封 / code 文案 / Cookie / 异步任务 / X-Request-ID / 同源部署）与已知限制；同步 `src/README.md` 的 api 目录说明 | 逐项对照 3.7 原文检查；文档链接到 backend README、src/api、src/views、tests |
| 2026-11-05 | 4 | 完成 4.1：真实三层跑通“界面输入 → 计算 → 存储 → 展示”。临时 SQLite + 内存会话启动 compute 8001、backend 8000、Vite 5173；Playwright 驱动预测页输入 CaF2 并点击预测，结果页正常显示 | 任务 succeeded（20 ms）；fluorite、配位数 8、a=5.381 Å、12 原子；DB trace_id=web-e759… 与后端/计算层日志一致；界面反馈“写法没问题，可以继续”，结果页含“已完成 / 萤石型 / 结果解释 / CaF2 / 5.38”与三维晶胞；控制台仅未登录 401 与 favicon 404，无页面报错；本次为一次性验收脚本，4.4 再固化为可重复回归测试 |
| 2026-11-07 | 1 | 扩展覆盖范围：新增六方纤锌矿（ZnO / ZnS）与金红石型（MgF2 / TiO2）原型，补 Ti⁴⁺ 离子半径、实测晶胞参数与前端四方晶胞识别；新增 `rule_wurtzite_experimental`、`rule_rutile_experimental`、`caveat_polymorph` 中英文案 | compute 138 通过 / 覆盖率 98.08%；frontend 97 通过（type-check、check-codes、build 通过）；ZnO a=3.2498 Å、c=5.2066 Å，ZnS a=3.8230 Å、c=6.2610 Å，MgF2 a=4.6210 Å、c=3.0520 Å，TiO2 a=4.5937 Å、c=2.9587 Å |
| 2026-11-07 | 2 | 账号与记录：登录后修改密码、安全问题自助重置、按账号分页的预测记录；新增 `users.security_question` / `security_answer_hash` 字段与 `GET /api/users/me/tasks` 接口；前端新增账号设置、找回密码、我的记录页面与顶栏入口 | backend `go test ./...` 全通过；frontend 120 通过（type-check、check-codes、build 通过）；check-codes 67 code / 254 文案；compute 138 通过 / 覆盖率 98.08% |
| 2026-11-09 | — | 修文档与代码不一致：`frontend/src/README.md` 与中文版的路由表补齐登录、找回密码、账号设置、预测记录；`frontend/src/router.ts` 顶部注释同步；`TASKS.md` 进度表去掉断表空行 | 只改文档与注释，未改行为；路由以 `frontend/src/router.ts` 为准 |
| 2026-11-09 | 2 | 修复队列满：任务服务每秒补扫 pending 任务，并用内存集合避免同一任务重复入队；队列满时任务不再等到重启，队列腾出后自动重算 | 新增 `tasks.QueueCapacity` 导出常量与回归测试（阻塞 worker 填满队列，溢出任务自动完成）；`go test ./...` 全通过 |
| 2026-11-09 | 3 | 三维渲染从 3Dmol 换成 Three.js：新增 `structureScene.ts`（纯几何）与 `structureRenderer.ts`（按需加载的球棍 / 晶胞 / 取景 / 资源释放），删除 CIF 转换与 3Dmol 依赖；新增 `check-csp` 脚本扫描构建产物中的 eval / Function 构造器 | frontend 123 个用例、type-check、build、check-csp 通过；主包约 292 kB（gzip 90.85 kB），Three.js chunk 约 563 kB（gzip 139.39 kB）；产物 0 处 eval / new Function |
| 2026-11-09 | 4 | 完成 4.2：并发与超时——新增 compute_timeout，超时不再误报为连不上；compute 客户端支持可配置超时（默认 10 秒）；补超时（含响应体卡住）与 80 条并发创建回归测试；前端补 compute_timeout 中英文案与结果页提示测试；CI 增加 go test -race | backend go test ./... 与 go test -race ./... 全通过；frontend 124 个用例、type-check、check-codes 通过（68 code / 255 文案）；超时任务约 0.07 秒落成 failed / compute_timeout，80 条并发任务全部成功 |
| 2026-11-12 | 2 | 把 SQLite 降为开发/测试驱动，生产改为一等公民 PostgreSQL：`YCSP_ENV=production` 拒绝 SQLite；连接池按驱动配置并启动 Ping；新增 `YCSP_DB_AUTO_MIGRATE`、`go run . -migrate-only` 与启动前 `EnsureSchema` 表校验；任务 worker 可配置（默认 SQLite 1、服务端库 4，范围 1~16）；测试辅助支持 Postgres 独立 schema，CI backend job 增加 Postgres service 与整套测试 | 新增连接池/配置/worker 并发/表校验测试；SQLite `go test ./...` 与 `go test -race ./...` 全通过；本地临时 Postgres 16 容器上 `go test ./...` 全通过（`store` / `tasks` / `api` / `users` 均跑真实 Postgres），并修掉并行测试 schema 重名问题（schema 名加入进程号） |
