# 数据处理层 · 测试（compute/tests）

用 `uv run pytest` 运行（从 `compute/` 根目录），命令与 CI 一致。

| 文件 | 覆盖内容 |
| --- | --- |
| `test_composition.py` | 化学式解析：下标、括号、空格，以及各种写错时的 code |
| `test_structure.py` | 元素 / 原子 / 晶胞的校验：非正棱长、夹角越界、坐标不是有限数 |
| `test_prototypes.py` | 9 种结构原型的晶胞边长公式、原子个数与分数坐标范围 |
| `test_prediction.py` | 课本经典例子（NaCl / CsCl / CaF₂ / Na₂O / ZnS / 金刚石 / SiC / Cu / Fe / Mg）的结构型与晶胞边长 |
| `test_element_data.py` | 常数包数据的一致性：漏登记元素、金属缺半径、离子缺电负性 |
| `test_constants.py` | 架构约束：常数包不得 import `app.domain` / `app.api` |
| `test_serialization.py` | 应答字段名、展示精度（所有浮点都舍入到 3 位） |
| `test_predict_api.py` | `POST /predict` 的应答契约：离子 / 金属两类结果，以及可翻译的 code |
| `test_api_errors.py` | `GET /health`、不支持的物质、未预期异常的统一应答 |
| `test_env.py` | 环境变量读取：缺省值、空值、非法整数 |

## 覆盖率

```powershell
uv run pytest --cov=app --cov-report=term-missing
```

当前约 99%。未覆盖的只有两类：`app/main.py` 的启动入口（会真的占用端口），
以及 `prediction.py` 里“元素数据缺失 / 半径比过小”这类防御分支——它们要构造坏数据才能触发，
另一侧由 `test_element_data.py` 的一致性断言守住。

CI 用 `--cov-fail-under=90` 兜底：新增代码如果完全没测，流水线会直接失败。

> 晶胞边长的允许误差是 0.01 Å。硬球模型本身就有几个百分点的系统偏差，
> 所以断言比较的是“是否与课本同一数量级且结构型一致”，而不是实验精修值。
