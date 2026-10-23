"""全部状态码：应答状态、结果依据、结果附注、错误。

文案由前端按当前语言翻译（见根 README「多语言」），所以这里只放语言无关的 code；
需要填进文案的数值放在 `Notice.params` 里。
"""

# —— 应答状态 ——
CODE_COMPUTE_HEALTHY = "compute_healthy"
# 健康检查通过

CODE_PREDICTION_READY = "prediction_ready"
# 预测完成，结果在 prediction 字段

CODE_REQUEST_INVALID_BODY = "request_invalid_body"
# 请求体不符合约定（框架层校验失败）

CODE_INTERNAL_ERROR = "internal_error"
# 未预期的服务端异常（完整堆栈只写日志）

# —— 结果依据：告诉学生“为什么给出这个结构” ——
RULE_RADIUS_RATIO = "rule_radius_ratio"
# 半径比规则（参数 ratio、coordination_number）

RULE_EXPERIMENTAL_OVERRIDE = "rule_experimental_override"
# 规则与实验不符，采用实验结构（参数 ratio、rule_prototype、rule_coordination_number）

RULE_FLUORITE_STOICHIOMETRY = "rule_fluorite_stoichiometry"
# 1:2 化学计量 → 萤石型（参数 ratio、cation_count、anion_count）

RULE_ANTIFLUORITE_STOICHIOMETRY = "rule_antifluorite_stoichiometry"
# 2:1 化学计量 → 反萤石型（参数 ratio、cation_count、anion_count）

RULE_DIAMOND_COVALENT = "rule_diamond_covalent"
# 单质共价晶体 → 金刚石型

RULE_ZINC_BLENDE_COVALENT = "rule_zinc_blende_covalent"
# 两种非金属元素 1:1 → 闪锌矿型（参数 electronegativity_difference）

RULE_METAL_EXPERIMENTAL = "rule_metal_experimental"
# 金属的结构型来自实验数据

# —— 结果附注：说明结果的局限与例外 ——
CAVEAT_RADIUS_RATIO_EMPIRICAL = "caveat_radius_ratio_empirical"
# 半径比只是经验规律，实际结构以实验测定为准

CAVEAT_NEAR_BOUNDARY = "caveat_near_boundary"
# 比值贴近分界，另一种配位数也可能（参数 ratio、boundary、alternative_prototype）

CAVEAT_CONFIRMED_BY_EXPERIMENT = "caveat_confirmed_by_experiment"
# 实验数据与规则结论一致

CAVEAT_LOW_IONIC_CHARACTER = "caveat_low_ionic_character"
# 电负性差偏小，键的共价性明显（参数 difference、threshold）

CAVEAT_METAL_STRUCTURE_EXPERIMENTAL = "caveat_metal_structure_experimental"
# 金属的堆积方式不能只由成分推导，结构型来自实验数据

CAVEAT_STOICHIOMETRY_OUTSIDE_BAND = "caveat_stoichiometry_outside_band"
# 萤石型的结构型按离子个数比判定，但半径比不在 8 配位区间（参数 ratio）

# —— 错误：输入或数据无法处理 ——
ERROR_FORMULA_EMPTY = "formula_empty"
# 化学式为空

ERROR_FORMULA_INVALID_CHARACTER = "formula_invalid_character"
# 出现了化学式里不该有的字符（参数 character、position）

ERROR_FORMULA_UNBALANCED_PARENTHESES = "formula_unbalanced_parentheses"
# 括号没有配对（参数 formula）

ERROR_FORMULA_ZERO_COUNT = "formula_zero_count"
# 下标写成了 0（参数 formula）

ERROR_FORMULA_CHARGE_UNBALANCED = "formula_charge_unbalanced"
# 正负电荷加起来不为 0（参数 formula）

ERROR_ELEMENT_NOT_SUPPORTED = "element_not_supported"
# 未收录该元素（参数 symbol、position）

ERROR_ELEMENT_DATA_MISSING = "element_data_missing"
# 收录了元素但缺少推算所需的数据（参数 symbol）

ERROR_COMPOSITION_NOT_SUPPORTED = "composition_not_supported"
# 这类化学式不在本平台覆盖范围内（参数 formula）

ERROR_RADIUS_RATIO_TOO_SMALL = "radius_ratio_too_small"
# 半径比小于 0.225，需要 3 配位结构，暂不覆盖（参数 ratio）
