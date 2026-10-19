"""数据处理层的内部数据结构：元素、原子、晶胞。

长度统一用埃（Ångström，记作 Å，1 Å = 10⁻¹⁰ m），角度用度（°）。
用 dataclass 而不是字典，是为了让字段名本身就说明“这是什么物理量”。
"""

from dataclasses import dataclass, field
from math import isfinite

# 三维空间中一个点由 3 个坐标分量描述（x, y, z）。
COORDINATE_COMPONENT_COUNT = 3
# 晶胞夹角必须严格介于 0° 与 180° 之间，否则晶胞退化或自交。
MIN_ANGLE_DEGREE = 0.0
MAX_ANGLE_DEGREE = 180.0


@dataclass(frozen=True)
class Element:
    """一种化学元素。

    只保留与语言无关的身份信息；元素名等展示文案由前端语言包提供
    （见根 README「多语言」），不写进数据模型。
    """

    symbol: str  # 元素符号，如 "Na"
    atomic_number: int  # 原子序数（质子数）

    def __post_init__(self) -> None:
        if not self.symbol:
            raise ValueError("元素符号不能为空。")
        if self.atomic_number <= 0:
            raise ValueError(f"原子序数必须是正整数，收到 {self.atomic_number}。")


@dataclass(frozen=True)
class Atom:
    """晶胞中的一个原子。"""

    element: Element
    # 分数坐标 (x, y, z)：每个分量是“相对所在晶胞棱长的比例”，无量纲，通常取 [0, 1)。
    fractional_coordinates: tuple[float, float, float]

    def __post_init__(self) -> None:
        if len(self.fractional_coordinates) != COORDINATE_COMPONENT_COUNT:
            raise ValueError(
                f"分数坐标必须由 {COORDINATE_COMPONENT_COUNT} 个数组成（x, y, z）。"
            )
        if not all(isfinite(value) for value in self.fractional_coordinates):
            raise ValueError(f"分数坐标必须是有限数，收到 {self.fractional_coordinates}。")


@dataclass
class UnitCell:
    """一个晶胞：三条棱长、三个夹角，以及位于其中的原子。"""

    edge_a_angstrom: float  # 棱长 a，单位 Å
    edge_b_angstrom: float  # 棱长 b，单位 Å
    edge_c_angstrom: float  # 棱长 c，单位 Å
    angle_alpha_degree: float  # 棱 b 与 c 的夹角 α，单位 °
    angle_beta_degree: float  # 棱 a 与 c 的夹角 β，单位 °
    angle_gamma_degree: float  # 棱 a 与 b 的夹角 γ，单位 °
    atoms: list[Atom] = field(default_factory=list)  # 晶胞内的原子

    def __post_init__(self) -> None:
        self._validate_edges()
        self._validate_angles()

    def _validate_edges(self) -> None:
        edge_lengths = (
            ("a", self.edge_a_angstrom),
            ("b", self.edge_b_angstrom),
            ("c", self.edge_c_angstrom),
        )
        for label, length in edge_lengths:
            if not isfinite(length) or length <= 0:
                raise ValueError(f"晶胞棱长 {label} 必须是正数，收到 {length}。")

    def _validate_angles(self) -> None:
        angles = (
            ("α", self.angle_alpha_degree),
            ("β", self.angle_beta_degree),
            ("γ", self.angle_gamma_degree),
        )
        for label, angle in angles:
            if not MIN_ANGLE_DEGREE < angle < MAX_ANGLE_DEGREE:
                raise ValueError(f"晶胞夹角 {label} 必须在 0° 与 180° 之间，收到 {angle}。")
