"""元素数据的一致性检查：补数据时能立刻发现漏项。"""

import pytest

from app.constants.elements import (
    ANION_CHARGES,
    CATION_CHARGES,
    COVALENT_RADII_PICOMETERS,
    DIAMOND_FORMING_ELEMENTS,
    IONIC_RADII_PICOMETERS,
    METALLIC_RADII_PICOMETERS,
    PAULING_ELECTRONEGATIVITY,
)
from app.constants.structures import METAL_STRUCTURE_TYPES
from app.domain.elements import COMMON_ELEMENTS, get_element, get_ionic_radius


def test_every_referenced_symbol_is_registered():
    referenced = (
        set(PAULING_ELECTRONEGATIVITY)
        | set(COVALENT_RADII_PICOMETERS)
        | set(METALLIC_RADII_PICOMETERS)
        | {symbol for symbol, _charge in IONIC_RADII_PICOMETERS}
        | set(METAL_STRUCTURE_TYPES)
        | set(DIAMOND_FORMING_ELEMENTS)
    )

    assert referenced <= set(COMMON_ELEMENTS), (
        f"这些元素没有登记到 COMMON_ELEMENTS：{sorted(referenced - set(COMMON_ELEMENTS))}"
    )


def test_every_metal_has_a_metallic_radius():
    missing = set(METAL_STRUCTURE_TYPES) - set(METALLIC_RADII_PICOMETERS)

    assert not missing, f"这些金属缺少金属半径：{sorted(missing)}"


def test_every_ion_has_an_electronegativity():
    missing = {symbol for symbol, _charge in IONIC_RADII_PICOMETERS} - set(
        PAULING_ELECTRONEGATIVITY
    )

    assert not missing, f"这些离子缺少电负性：{sorted(missing)}"


def test_each_element_has_a_single_ionic_charge():
    """电价表由离子半径表反推，前提是每种元素只收录一种电价。"""
    ionic_symbols = {symbol for symbol, _charge in IONIC_RADII_PICOMETERS}

    assert len(CATION_CHARGES) + len(ANION_CHARGES) == len(ionic_symbols)
    assert not set(CATION_CHARGES) & set(ANION_CHARGES)


def test_unknown_element_and_ion_raise_with_guidance():
    with pytest.raises(KeyError):
        get_element("Xx")
    with pytest.raises(KeyError):
        get_ionic_radius("Na", 3)


def test_getters_return_the_registered_values():
    assert get_element("Na").atomic_number == 11
    assert get_ionic_radius("Na", 1) == pytest.approx(102.0)
