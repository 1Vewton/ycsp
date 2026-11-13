"""晶体结构原型的常数：结构型编号、配位数、分数坐标与几何比值。

坐标都是分数坐标（相对棱长的比例，无量纲）；几何比值来自硬球模型。
把“结构型 + 半径”真正算成晶胞的函数在 `app/domain/prototypes.py`。
"""

from math import sqrt
from typing import NamedTuple

PROTOTYPE_ROCK_SALT = "rock_salt"
PROTOTYPE_CESIUM_CHLORIDE = "cesium_chloride"
PROTOTYPE_ZINC_BLENDE = "zinc_blende"
PROTOTYPE_WURTZITE = "wurtzite"
PROTOTYPE_FLUORITE = "fluorite"
PROTOTYPE_ANTIFLUORITE = "antifluorite"
PROTOTYPE_RUTILE = "rutile"
PROTOTYPE_DIAMOND = "diamond"
PROTOTYPE_FACE_CENTERED_CUBIC = "face_centered_cubic"
PROTOTYPE_BODY_CENTERED_CUBIC = "body_centered_cubic"
PROTOTYPE_HEXAGONAL_CLOSE_PACKED = "hexagonal_close_packed"

# 每种原型里阳离子（或金属原子）的配位数：最近邻原子的个数。
PROTOTYPE_COORDINATION_NUMBERS: dict[str, int] = {
    PROTOTYPE_ROCK_SALT: 6,
    PROTOTYPE_CESIUM_CHLORIDE: 8,
    PROTOTYPE_ZINC_BLENDE: 4,
    PROTOTYPE_WURTZITE: 4,
    PROTOTYPE_FLUORITE: 8,
    PROTOTYPE_ANTIFLUORITE: 8,
    PROTOTYPE_RUTILE: 6,
    PROTOTYPE_DIAMOND: 4,
    PROTOTYPE_FACE_CENTERED_CUBIC: 12,
    PROTOTYPE_BODY_CENTERED_CUBIC: 8,
    PROTOTYPE_HEXAGONAL_CLOSE_PACKED: 12,
}

# 常见金属的实测结构型（A 型堆积）。金属的堆积方式不能只由成分推导，
# 所以这里放的是实验数据；结果里用 RULE_METAL_EXPERIMENTAL 注明来源。
METAL_STRUCTURE_TYPES: dict[str, str] = {
    # 面心立方最密堆积（A1）
    "Al": PROTOTYPE_FACE_CENTERED_CUBIC,
    "Ni": PROTOTYPE_FACE_CENTERED_CUBIC,
    "Cu": PROTOTYPE_FACE_CENTERED_CUBIC,
    "Ag": PROTOTYPE_FACE_CENTERED_CUBIC,
    "Au": PROTOTYPE_FACE_CENTERED_CUBIC,
    "Pb": PROTOTYPE_FACE_CENTERED_CUBIC,
    # 体心立方堆积（A2）
    "Na": PROTOTYPE_BODY_CENTERED_CUBIC,
    "K": PROTOTYPE_BODY_CENTERED_CUBIC,
    "Cr": PROTOTYPE_BODY_CENTERED_CUBIC,
    "Fe": PROTOTYPE_BODY_CENTERED_CUBIC,
    "Mo": PROTOTYPE_BODY_CENTERED_CUBIC,
    "W": PROTOTYPE_BODY_CENTERED_CUBIC,
    "Cs": PROTOTYPE_BODY_CENTERED_CUBIC,
    # 六方最密堆积（A3）
    "Mg": PROTOTYPE_HEXAGONAL_CLOSE_PACKED,
    "Ti": PROTOTYPE_HEXAGONAL_CLOSE_PACKED,
    "Co": PROTOTYPE_HEXAGONAL_CLOSE_PACKED,
    "Zn": PROTOTYPE_HEXAGONAL_CLOSE_PACKED,
}

# 课本经典离子化合物的实测结构型（含 1:1 与 1:2）。
# 半径比规则只是经验规律，边界附近会判错（例如 KCl 比值约 0.76，规则判成 8 配位，
# 实测却是 NaCl 型），所以这些化合物直接采用实验数据，并在结果里说明差异。
KNOWN_IONIC_STRUCTURES: dict[str, str] = {
    "NaF": PROTOTYPE_ROCK_SALT,
    "NaCl": PROTOTYPE_ROCK_SALT,
    "NaBr": PROTOTYPE_ROCK_SALT,
    "NaI": PROTOTYPE_ROCK_SALT,
    "KF": PROTOTYPE_ROCK_SALT,
    "KCl": PROTOTYPE_ROCK_SALT,
    "KBr": PROTOTYPE_ROCK_SALT,
    "KI": PROTOTYPE_ROCK_SALT,
    "CsF": PROTOTYPE_ROCK_SALT,
    "CsCl": PROTOTYPE_CESIUM_CHLORIDE,
    "CsBr": PROTOTYPE_CESIUM_CHLORIDE,
    "CsI": PROTOTYPE_CESIUM_CHLORIDE,
    "MgO": PROTOTYPE_ROCK_SALT,
    "MgS": PROTOTYPE_ROCK_SALT,
    "CaO": PROTOTYPE_ROCK_SALT,
    "CaS": PROTOTYPE_ROCK_SALT,
    "BaO": PROTOTYPE_ROCK_SALT,
    "BaS": PROTOTYPE_ROCK_SALT,
    "FeO": PROTOTYPE_ROCK_SALT,
    "ZnO": PROTOTYPE_WURTZITE,
    "ZnS": PROTOTYPE_WURTZITE,
    "CaF2": PROTOTYPE_FLUORITE,
    "MgF2": PROTOTYPE_RUTILE,
    "TiO2": PROTOTYPE_RUTILE,
}


class WurtziteCellParameters(NamedTuple):
    """六方纤锌矿型的实验晶胞参数（Å）；u 是阴离子沿 c 轴的分数坐标。"""

    edge_a_angstrom: float
    edge_c_angstrom: float
    anion_position_u: float


class RutileCellParameters(NamedTuple):
    """四方金红石型的实验晶胞参数（Å）；u 是阴离子位置参数。"""

    edge_a_angstrom: float
    edge_c_angstrom: float
    anion_position_u: float


# 课本常见纤锌矿型化合物的实验晶胞参数（a、c 单位 Å）。
# 数据取自常用晶体学数据表；u 是阴离子沿 c 轴的分数坐标，理想值 3/8。
WURTZITE_CELL_PARAMETERS: dict[str, WurtziteCellParameters] = {
    "ZnO": WurtziteCellParameters(3.2498, 5.2066, 0.3819),
    "ZnS": WurtziteCellParameters(3.8230, 6.2610, 0.3750),
}

# 课本常见金红石型化合物的实验晶胞参数（a、c 单位 Å）。
# 数据取自常用晶体学数据表；u 是阴离子位置参数。
RUTILE_CELL_PARAMETERS: dict[str, RutileCellParameters] = {
    "MgF2": RutileCellParameters(4.6210, 3.0520, 0.3050),
    "TiO2": RutileCellParameters(4.5937, 2.9587, 0.3048),
}

# 这些物质除了六方纤锌矿型，还有别的常见晶型（例如 ZnS 的立方闪锌矿型），
# 结果里要用 caveat_polymorph 提醒学生。
POLYMORPHIC_WURTZITE_FORMULAS: frozenset[str] = frozenset({"ZnS"})

# 立方与六方晶胞的角度（°）。
RIGHT_ANGLE_DEGREE = 90.0
HEXAGONAL_GAMMA_DEGREE = 120.0
# 立方晶胞的体对角线 = √3·a、面对角线 = √2·a。
CUBIC_BODY_DIAGONAL_RATIO = sqrt(3.0)
CUBIC_FACE_DIAGONAL_RATIO = sqrt(2.0)
# 六方最密堆积的理想轴比 c/a = √(8/3) ≈ 1.633（硬球模型的几何结果）。
IDEAL_HEXAGONAL_AXIAL_RATIO = sqrt(8.0 / 3.0)

# 面心立方格点：8 个顶点 + 6 个面心，折算成晶胞内 4 个位置。
FACE_CENTERED_SITES = (
    (0.0, 0.0, 0.0),
    (0.0, 0.5, 0.5),
    (0.5, 0.0, 0.5),
    (0.5, 0.5, 0.0),
)
# 八面体空隙：12 条棱的中点折算成 3 个，加上体心 1 个，共 4 个。
OCTAHEDRAL_HOLE_SITES = (
    (0.5, 0.0, 0.0),
    (0.0, 0.5, 0.0),
    (0.0, 0.0, 0.5),
    (0.5, 0.5, 0.5),
)
# 四面体空隙共 8 个，分成两组；闪锌矿型与金刚石型各占其中一组。
TETRAHEDRAL_HOLE_SITES_FIRST = (
    (0.25, 0.25, 0.25),
    (0.25, 0.75, 0.75),
    (0.75, 0.25, 0.75),
    (0.75, 0.75, 0.25),
)
TETRAHEDRAL_HOLE_SITES_SECOND = (
    (0.75, 0.75, 0.75),
    (0.75, 0.25, 0.25),
    (0.25, 0.75, 0.25),
    (0.25, 0.25, 0.75),
)
ALL_TETRAHEDRAL_HOLE_SITES = TETRAHEDRAL_HOLE_SITES_FIRST + TETRAHEDRAL_HOLE_SITES_SECOND
# 体心立方：顶点 1 个 + 体心 1 个。
BODY_CENTERED_SITES = ((0.0, 0.0, 0.0), (0.5, 0.5, 0.5))
# 六方最密堆积：底面顶点 1 个 + 体内 1 个。
HEXAGONAL_CLOSE_PACKED_SITES = ((0.0, 0.0, 0.0), (1.0 / 3.0, 2.0 / 3.0, 0.5))

# 六方纤锌矿型：阳离子在 (1/3, 2/3, 0) 与 (2/3, 1/3, 1/2)；
# 阴离子在同一组 x-y 位置上，沿 c 轴偏移 u 与 u + 1/2。
WURTZITE_CATION_SITES = (
    (1.0 / 3.0, 2.0 / 3.0, 0.0),
    (2.0 / 3.0, 1.0 / 3.0, 0.5),
)
WURTZITE_ANION_XY = ((1.0 / 3.0, 2.0 / 3.0), (2.0 / 3.0, 1.0 / 3.0))
WURTZITE_ANION_Z_OFFSETS = (0.0, 0.5)
# 理想纤锌矿的阴离子位置参数 u = 3/8（硬球模型的几何结果）。
IDEAL_WURTZITE_ANION_U = 3.0 / 8.0

# 四方金红石型：阳离子在顶点与体心；阴离子的 x-y-z 坐标由位置参数 u 决定。
RUTILE_CATION_SITES = ((0.0, 0.0, 0.0), (0.5, 0.5, 0.5))
