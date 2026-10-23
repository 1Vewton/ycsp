# 数据处理层 · 领域模型（compute/app/domain）

晶体学数据类型与预测规则；元素数据、结构原型数据、阈值与状态码在 [../constants/README.md](../constants/README.md)。

| 文件 | 作用 |
| --- | --- |
| `structure.py` | 元素 / 原子 / 晶胞（长度用 Å，角度用 °），皮米换埃也在这里 |
| `elements.py` | 建 `Element` 对象并提供查表访问器（`find_*` 返回 None，`get_*` 报错并提示去哪里补） |
| `composition.py` | 化学式解析（`CaF2`、`Ca(OH)2` → 元素 → 个数） |
| `prototypes.py` | 按硬球相切模型算晶胞边长并放置原子（NaCl / CsCl / ZnS / 萤石 / 反萤石 / 金刚石 / fcc / bcc / hcp） |
| `prediction.py` | 判类型 + 选原型 + 说明依据与附注，对外只暴露 `predict_structure` |
| `notice.py` | `Notice`（code + params）与 `DomainError`：语言无关的说明与错误 |

判据的“为什么”写在 `prediction.py` 的注释里，数据来源写在 `constants/` 各文件顶部；
本层不产生任何面向学生的文字，文案全部由前端按 code 翻译。

