# 数据处理层（compute）

按化学组成预测晶体结构，并以 HTTP 接口提供给数据管理层调用（Python + FastAPI）。

## 对外接口

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| GET | `/health` | 健康检查 |
| POST | `/predict` | 按化学式预测晶体结构 |

实现见 [app/api/README.md](./app/api/README.md)。应答统一为 `{ service, status, code, params? }`：
成功时另有 `prediction`；出错时只给语言无关的 `code` 与 `params`，文案由前端翻译（见根 README「多语言」）。

### 预测请求与应答

请求只需一个化学式，例如 `NaCl`、`CaF2`、`C`、`Cu`（也接受 `Ca(OH)2` 这类带括号的写法）：

```text
POST /predict
{ "formula": "NaCl" }
```

应答（节选）：

```json
{
  "service": "ycsp-compute",
  "status": "ok",
  "code": "prediction_ready",
  "prediction": {
    "formula": "NaCl",
    "family": "ionic",
    "candidates": [{
      "prototype": "rock_salt",
      "coordination_number": 6,
      "rule": { "code": "rule_radius_ratio", "params": { "ratio": 0.564, "coordination_number": 6 } },
      "caveats": [{ "code": "caveat_confirmed_by_experiment", "params": {} }],
      "cell": { "edge_a_angstrom": 5.66, "angle_alpha_degree": 90.0 },
      "atoms": [{ "symbol": "Na", "fractional_coordinates": [0.0, 0.0, 0.0] }]
    }]
  }
}
```

- `cell` 是晶胞参数（棱长 Å、夹角 °），`atoms` 是原子与分数坐标，前端可直接拿去画图。
- `rule` 说明“为什么给出这个结构”，`caveats` 说明结果的局限；两者都只含 code 与数值，
  前端按 code 翻成一句话（文案见 `frontend/src/locales/`）。

## 算法思路

先按组成判断晶体类型，再套课本规则（代码见 `app/domain/prediction.py`）：

| 类型 | 判据 | 结构型 |
| --- | --- | --- |
| 离子晶体 | 一种金属阳离子 + 一种阴离子，且电荷守恒 | 1:1 用 Pauling 半径比规则；1:2 / 2:1 用萤石型 / 反萤石型 |
| 共价晶体 | C / Si / Ge 单质；两种非金属 1:1（如 SiC） | 金刚石型；SiC 用闪锌矿型 |
| 金属晶体 | 只含一种金属元素 | 面心立方 / 体心立方 / 六方最密堆积（查实验数据表） |

半径比规则的分界（无量纲）：小于 0.225 需 3 配位、0.225~0.414 为 4 配位、0.414~0.732 为 6 配位、
大于 0.732 为 8 配位。晶胞边长由硬球相切模型算出，例如 NaCl 沿棱相切 a = 2(r₊+r₋) = 5.66 Å（课本 5.64 Å）。

元素数据、结构原型数据与这些判据阈值都集中在 `app/constants/`；`app/domain/` 只放类型与规则。

本阶段只用标准库 `math` 完成这些计算；NumPy 留到需要批量坐标或矩阵运算时再引入，避免为用而用。

## 已知限制

- **半径比只是经验规律**：比值贴近分界时两种配位数都可能，结果用 `caveat_near_boundary` 标出。
  KCl 的比值约 0.76，规则判成 8 配位，实测却是 NaCl 型；因此课本经典化合物直接采用实验数据
  （`app/constants/structures.py` 的 `KNOWN_IONIC_STRUCTURES`），并用 `rule_experimental_override` 说明差异。
- **金属的堆积方式不能只由成分推导**（同种金属在不同温度下结构可能不同），结构型取自实验数据，
  界面必须注明（`caveat_metal_structure_experimental`）。
- **六方纤锌矿型暂不覆盖**：ZnO 会被规则判成 NaCl 型，属已知偏差。
- **1:2 / 2:1 的结构型按离子个数比判定**（萤石型 / 反萤石型，如 CaF₂、Na₂O）；半径比偏离 8 配位区间时
  （如 MgF₂ 实际是金红石型）结果会带 `caveat_stoichiometry_outside_band` 附注，提示以实验为准。
- 只覆盖上面三种晶体类型；多元素（如 CaCO₃）、2:3 计量（如 Al₂O₃）、3 配位结构会返回
  `composition_not_supported` / `radius_ratio_too_small`，不会硬给一个结果。
- 计算是纯算术、毫秒级完成，因此不需要进度条与取消入口。

## 目录

| 路径 | 作用 |
| --- | --- |
| `app/` | 服务代码，见 [app/README.md](./app/README.md) |
| `tests/` | 单元测试（课本经典例子），见 [tests/README.md](./tests/README.md) |

## 如何运行

依赖用 [uv](https://docs.astral.sh/uv/) 管理（版本锁在 `uv.lock`，首次运行自动建好 `.venv`）。

```powershell
uv run python -m app.main   # 服务：默认 127.0.0.1:8001
uv run pytest               # 单元测试
uv run pytest --cov=app --cov-report=term-missing   # 单元测试 + 覆盖率
```

服务启动后：健康检查 <http://127.0.0.1:8001/health>，接口文档 <http://127.0.0.1:8001/docs>。
要改监听地址或端口，复制 `.env.example` 为 `.env` 再编辑。
