"""判断晶体类型的经验阈值（都来自课本口径，无量纲）。

改这些数字等于改教学口径，请同时更新 `compute/README.md`《算法思路》。
"""

from app.constants.structures import (
    PROTOTYPE_CESIUM_CHLORIDE,
    PROTOTYPE_ROCK_SALT,
    PROTOTYPE_ZINC_BLENDE,
)

# Pauling 半径比规则的分界：比值小于 0.225 需要 3 配位，
# 0.225~0.414 为 4 配位，0.414~0.732 为 6 配位，大于 0.732 为 8 配位。
RADIUS_RATIO_TETRAHEDRAL_LOWER = 0.225
RADIUS_RATIO_OCTAHEDRAL_LOWER = 0.414
RADIUS_RATIO_CUBIC_LOWER = 0.732
# 萤石型的理想区间上限：比值达到 1 时阳离子已不小于阴离子。
RADIUS_RATIO_MAXIMUM = 1.0

# 比值离分界小于这个量时，两种配位数都可能，结果要如实标注。
NEAR_BOUNDARY_TOLERANCE = 0.05
# 阴阳离子电负性差小于 1.7 时键的离子性明显减弱（Pauling 经验阈值）。
IONIC_ELECTRONEGATIVITY_THRESHOLD = 1.7

# 半径比的分界，以及分界两侧的结构型：用来提示“另一种可能”。
RADIUS_RATIO_BOUNDARIES = (
    (RADIUS_RATIO_OCTAHEDRAL_LOWER, PROTOTYPE_ZINC_BLENDE, PROTOTYPE_ROCK_SALT),
    (RADIUS_RATIO_CUBIC_LOWER, PROTOTYPE_ROCK_SALT, PROTOTYPE_CESIUM_CHLORIDE),
)
