# ycsp

面向青少年的 CSP（晶体结构预测）教学平台：输入物质的化学成分，预测它最可能形成的晶体结构。学生不需要任何计算或编程基础。

## 项目定位

面向青少年，用户不预设任何计算或编程基础；以“中学生能否独立看懂并操作”为唯一验收标准。

## 整体架构

| 分层 | 职责 | 技术栈 |
| --- | --- | --- |
| 前端 | 学生界面与晶体结构可视化 | TypeScript + Vue |
| 数据管理 | 统一对外接口、数据存储、转发计算请求 | Go + Gin + GORM |
| 数据处理 | 晶体结构预测与数值计算 | Python + FastAPI + NumPy |

请求流向：`前端(Vue) → 数据管理(Go) → 数据处理(Python)`；前端只与数据管理通信，由它再调用数据处理。

## 多语言

产品要能切换界面语言，让不同语言的用户都能独立操作：

- 默认中文，并支持英文（中/英可切换）。
- 文案与代码分离：前端用语言包（如 vue-i18n），不在组件里写死文字。
- 后端只返回错误码与参数，由前端按当前语言翻译，避免把某种语言的提示写死在后端。

## 安装与启动

需要 Go 1.25+、Node.js 22+、[uv](https://docs.astral.sh/uv/)（数据处理层用它管理 Python 依赖与虚拟环境）。

端口与地址都有默认值，直接就能跑；需要改的话，把对应层的 `.env.example` 复制为 `.env` 再编辑：

```powershell
Copy-Item compute\.env.example compute\.env
Copy-Item backend\.env.example backend\.env
Copy-Item frontend\.env.example frontend\.env
```

然后在三个终端分别启动：

```powershell
# 数据处理层（默认 8001）
cd compute
uv run python -m app.main

# 数据管理层（8000）
cd backend
go mod tidy; go run .

# 前端（5173）
cd frontend
npm install; npm run dev
```

打开 <http://localhost:5173/>，点页面按钮；若显示“你好，我是数据处理层……”即三层已打通。

## 目录导航

| 目录 | 分层 | 文档 |
| --- | --- | --- |
| `frontend/` | 前端 | [README](./frontend/README.md) |
| `backend/` | 数据管理 | [README](./backend/README.md) |
| `compute/` | 数据处理 | [README](./compute/README.md) |

CI 配置见 [`.github/README.md`](./.github/README.md)，开发计划见 [TASKS.md](./TASKS.md)。
