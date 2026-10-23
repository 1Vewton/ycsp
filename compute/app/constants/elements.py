"""元素属性数据（教学用）：只放数据，不放逻辑。

数据来源：
- 电负性：Pauling 标度（Wikipedia《Electronegativities of the elements (data page)》）。
- 离子半径：Shannon (1976) 有效离子半径，配位数 6，单位皮米（pm）。
- 金属半径：12 配位的金属半径，单位皮米（pm）。
- 共价半径：Cordero (2008)，单位皮米（pm）；碳取 sp³ 半径 77 pm，
  与金刚石中 C–C 键长 154 pm 一致。

只收录课本常见元素；需要更多元素时按同样格式补充即可
（查表与报错提示在 `app/domain/elements.py`）。
"""

# 常见元素的原子序数（符号 → 质子数）。
ATOMIC_NUMBERS: dict[str, int] = {
    "H": 1,
    "C": 6,
    "N": 7,
    "O": 8,
    "F": 9,
    "Na": 11,
    "Mg": 12,
    "Al": 13,
    "Si": 14,
    "P": 15,
    "S": 16,
    "Cl": 17,
    "K": 19,
    "Ca": 20,
    "Ti": 22,
    "Cr": 24,
    "Fe": 26,
    "Co": 27,
    "Ni": 28,
    "Cu": 29,
    "Zn": 30,
    "Ge": 32,
    "Br": 35,
    "Mo": 42,
    "Ag": 47,
    "I": 53,
    "Cs": 55,
    "Ba": 56,
    "W": 74,
    "Au": 79,
    "Pb": 82,
}

# Pauling 电负性（无量纲）。
PAULING_ELECTRONEGATIVITY: dict[str, float] = {
    "H": 2.20,
    "C": 2.55,
    "N": 3.04,
    "O": 3.44,
    "F": 3.98,
    "Na": 0.93,
    "Mg": 1.31,
    "Al": 1.61,
    "Si": 1.90,
    "P": 2.19,
    "S": 2.58,
    "Cl": 3.16,
    "K": 0.82,
    "Ca": 1.00,
    "Ba": 0.89,
    "Fe": 1.83,
    "Br": 2.96,
    "I": 2.66,
    "Cs": 0.79,
    "Ti": 1.54,
    "Cr": 1.66,
    "Co": 1.88,
    "Ni": 1.91,
    "Cu": 1.90,
    "Zn": 1.65,
    "Ge": 2.01,
    "Mo": 2.16,
    "Ag": 1.93,
    "W": 2.36,
    "Au": 2.54,
    "Pb": 2.33,
}

# 离子半径（pm）：键为（元素符号, 电荷数）。
# 同一元素在不同电荷、不同配位数下半径不同，所以不能只按元素存一个值。
IONIC_RADII_PICOMETERS: dict[tuple[str, int], float] = {
    ("Na", 1): 102.0,
    ("K", 1): 138.0,
    ("Cs", 1): 167.0,
    ("Mg", 2): 72.0,
    ("Ca", 2): 100.0,
    ("Ba", 2): 135.0,
    ("Al", 3): 53.5,
    ("Zn", 2): 74.0,
    ("Fe", 2): 78.0,
    ("F", -1): 133.0,
    ("Cl", -1): 181.0,
    ("Br", -1): 196.0,
    ("I", -1): 220.0,
    ("O", -2): 140.0,
    ("S", -2): 184.0,
}

# 共价半径（pm）：两个原子共用电子对时，核间距的一半。
COVALENT_RADII_PICOMETERS: dict[str, float] = {
    "H": 31.0,
    "C": 77.0,
    "N": 71.0,
    "O": 66.0,
    "F": 57.0,
    "Al": 121.0,
    "Si": 111.0,
    "P": 107.0,
    "S": 105.0,
    "Cl": 102.0,
    "Zn": 122.0,
    "Ge": 120.0,
    "Br": 120.0,
    "I": 139.0,
}

# 金属半径（pm）：金属晶体中把原子看成相切硬球时的半径。
METALLIC_RADII_PICOMETERS: dict[str, float] = {
    "Na": 186.0,
    "Mg": 160.0,
    "Al": 143.0,
    "K": 227.0,
    "Ca": 197.0,
    "Ti": 147.0,
    "Cr": 128.0,
    "Fe": 126.0,
    "Co": 125.0,
    "Ni": 124.0,
    "Cu": 128.0,
    "Zn": 134.0,
    "Mo": 139.0,
    "Ag": 144.0,
    "Cs": 265.0,
    "Ba": 222.0,
    "W": 139.0,
    "Au": 144.0,
    "Pb": 175.0,
}

# 单质以金刚石型共价晶体存在的元素（课本常见的三个）。
DIAMOND_FORMING_ELEMENTS: frozenset[str] = frozenset({"C", "Si", "Ge"})

# 常见电价（无量纲整数）：由离子半径表反推，避免两处数据写重、写岔。
# 前提是每种元素只收录一种电价（见 tests/test_element_data.py）。
CATION_CHARGES: dict[str, int] = {
    symbol: charge for (symbol, charge) in IONIC_RADII_PICOMETERS if charge > 0
}
ANION_CHARGES: dict[str, int] = {
    symbol: -charge for (symbol, charge) in IONIC_RADII_PICOMETERS if charge < 0
}
