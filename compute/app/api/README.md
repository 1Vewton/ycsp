# 数据处理层 · 接口层（compute/app/api）

数据管理层调用的 HTTP 接口。

| 文件 | 作用 |
| --- | --- |
| `health.py` | `GET /health` 路由 |
| `predict.py` | `POST /predict` 路由：把领域层的预测结果包成统一应答 |
| `serialization.py` | 结果转字典：晶胞参数、原子坐标，浮点数按展示精度舍入 |
| `display.py` | 展示精度（保留几位小数）与舍入函数 |
| `errors.py` | 框架层错误（请求体不合法、未预期异常）也转成统一应答，避免界面出现英文报错 |

状态码统一放在 [../constants/codes.py](../constants/codes.py)（语言无关，文案由前端翻译）。

