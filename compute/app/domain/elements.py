"""元素对象与查表访问器；元素数据本身在 `app/constants/elements.py`。

这里只做“把符号变成 Element / 半径数值”的转换，不负责校验化学式的写法。
"""

from app.constants.elements import (
    ATOMIC_NUMBERS,
    COVALENT_RADII_PICOMETERS,
    IONIC_RADII_PICOMETERS,
    METALLIC_RADII_PICOMETERS,
    PAULING_ELECTRONEGATIVITY,
)
from app.domain.structure import Element

# 常见元素（符号 → Element）。Element 自带校验，所以构造一次后缓存复用。
COMMON_ELEMENTS: dict[str, Element] = {
    symbol: Element(symbol=symbol, atomic_number=atomic_number)
    for symbol, atomic_number in ATOMIC_NUMBERS.items()
}


def find_element(symbol: str) -> Element | None:
    """按元素符号取元素；未收录时返回 None。"""
    return COMMON_ELEMENTS.get(symbol)


def get_element(symbol: str) -> Element:
    """按元素符号取元素；未收录时报错并提示去哪里补。"""
    element = find_element(symbol)
    if element is None:
        raise KeyError(f"未收录元素 {symbol}，可在 app/constants/elements.py 中补充。")
    return element


def find_ionic_radius(symbol: str, charge: int) -> float | None:
    """按元素符号与电荷数取离子半径（pm）；未收录时返回 None。"""
    return IONIC_RADII_PICOMETERS.get((symbol, charge))


def get_ionic_radius(symbol: str, charge: int) -> float:
    """按元素符号与电荷数取离子半径（pm）；只收录了配位数 6 的常见离子。"""
    radius = find_ionic_radius(symbol, charge)
    if radius is None:
        raise KeyError(
            f"未收录离子 {symbol}({charge:+d})（配位数 6）的半径，"
            "可在 app/constants/elements.py 中补充。"
        )
    return radius


def find_electronegativity(symbol: str) -> float | None:
    """取 Pauling 电负性（无量纲）；未收录时返回 None。"""
    return PAULING_ELECTRONEGATIVITY.get(symbol)


def find_covalent_radius(symbol: str) -> float | None:
    """取共价半径（pm）；未收录时返回 None。"""
    return COVALENT_RADII_PICOMETERS.get(symbol)


def find_metallic_radius(symbol: str) -> float | None:
    """取金属半径（pm）；未收录时返回 None。"""
    return METALLIC_RADII_PICOMETERS.get(symbol)
