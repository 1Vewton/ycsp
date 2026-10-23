"""结构原型构建器：晶胞边长公式、原子个数与摆放位置。

期望值直接用整数半径手算，避免测试把生产代码的浮点数据再抄一遍。
"""

from collections.abc import Callable
from math import sqrt

import pytest

from app.constants.structures import (
    PROTOTYPE_COORDINATION_NUMBERS,
    PROTOTYPE_HEXAGONAL_CLOSE_PACKED,
)
from app.domain import prototypes
from app.domain.elements import COMMON_ELEMENTS
from app.domain.structure import UnitCell

# 用 100 pm 与 180 pm 两个整数半径：相切距离就是 2.80 Å，手算期望值不容易错。
CATION_RADIUS_PICOMETERS = 100.0
ANION_RADIUS_PICOMETERS = 180.0
CONTACT_ANGSTROM = (CATION_RADIUS_PICOMETERS + ANION_RADIUS_PICOMETERS) / 100.0

ELEMENT_KEY = "Na"
ANION_KEY = "Cl"


def atoms_per_symbol(cell: UnitCell) -> dict[str, int]:
    counts: dict[str, int] = {}
    for atom in cell.atoms:
        counts[atom.element.symbol] = counts.get(atom.element.symbol, 0) + 1
    return counts


def build_ionic_prototype(build_function: Callable[..., UnitCell]) -> UnitCell:
    return build_function(
        COMMON_ELEMENTS[ELEMENT_KEY],
        CATION_RADIUS_PICOMETERS,
        COMMON_ELEMENTS[ANION_KEY],
        ANION_RADIUS_PICOMETERS,
    )


def test_rock_salt_edge_is_twice_the_contact_distance():
    cell = build_ionic_prototype(prototypes.build_rock_salt)

    assert cell.edge_a_angstrom == pytest.approx(2 * CONTACT_ANGSTROM)
    assert atoms_per_symbol(cell) == {"Na": 4, "Cl": 4}


def test_cesium_chloride_edge_uses_body_diagonal():
    cell = build_ionic_prototype(prototypes.build_cesium_chloride)

    assert cell.edge_a_angstrom == pytest.approx(2 * CONTACT_ANGSTROM / sqrt(3.0))
    assert atoms_per_symbol(cell) == {"Na": 1, "Cl": 1}


def test_zinc_blende_places_cation_in_tetrahedral_holes():
    cell = build_ionic_prototype(prototypes.build_zinc_blende)

    assert cell.edge_a_angstrom == pytest.approx(4 * CONTACT_ANGSTROM / sqrt(3.0))
    assert atoms_per_symbol(cell) == {"Na": 4, "Cl": 4}


def test_fluorite_puts_anion_in_all_eight_holes():
    cell = build_ionic_prototype(prototypes.build_fluorite)

    assert cell.edge_a_angstrom == pytest.approx(4 * CONTACT_ANGSTROM / sqrt(3.0))
    assert atoms_per_symbol(cell) == {"Na": 4, "Cl": 8}


def test_antifluorite_swaps_the_two_ions():
    cell = build_ionic_prototype(prototypes.build_antifluorite)

    assert cell.edge_a_angstrom == pytest.approx(4 * CONTACT_ANGSTROM / sqrt(3.0))
    assert atoms_per_symbol(cell) == {"Na": 8, "Cl": 4}


def test_diamond_edge_is_four_times_the_bond_length_over_sqrt_three():
    radius_picometers = 77.0
    cell = prototypes.build_diamond(COMMON_ELEMENTS["C"], radius_picometers)

    bond_length_angstrom = 2 * radius_picometers / 100.0
    assert cell.edge_a_angstrom == pytest.approx(4 * bond_length_angstrom / sqrt(3.0))
    assert atoms_per_symbol(cell) == {"C": 8}


def test_face_centered_cubic_touches_along_face_diagonal():
    radius_picometers = 128.0
    cell = prototypes.build_face_centered_cubic(COMMON_ELEMENTS["Cu"], radius_picometers)

    assert cell.edge_a_angstrom == pytest.approx(4 * radius_picometers / 100.0 / sqrt(2.0))
    assert atoms_per_symbol(cell) == {"Cu": 4}


def test_body_centered_cubic_touches_along_body_diagonal():
    radius_picometers = 126.0
    cell = prototypes.build_body_centered_cubic(COMMON_ELEMENTS["Fe"], radius_picometers)

    assert cell.edge_a_angstrom == pytest.approx(4 * radius_picometers / 100.0 / sqrt(3.0))
    assert atoms_per_symbol(cell) == {"Fe": 2}


def test_hexagonal_close_packed_uses_ideal_axial_ratio():
    radius_picometers = 160.0
    cell = prototypes.build_hexagonal_close_packed(COMMON_ELEMENTS["Mg"], radius_picometers)

    assert cell.edge_a_angstrom == pytest.approx(2 * radius_picometers / 100.0)
    assert cell.edge_c_angstrom == pytest.approx(sqrt(8.0 / 3.0) * cell.edge_a_angstrom)
    assert cell.angle_gamma_degree == pytest.approx(120.0)
    assert atoms_per_symbol(cell) == {"Mg": 2}


def test_every_fractional_coordinate_stays_inside_the_cell():
    cells = (
        build_ionic_prototype(prototypes.build_rock_salt),
        build_ionic_prototype(prototypes.build_cesium_chloride),
        build_ionic_prototype(prototypes.build_zinc_blende),
        build_ionic_prototype(prototypes.build_fluorite),
        build_ionic_prototype(prototypes.build_antifluorite),
        prototypes.build_diamond(COMMON_ELEMENTS["C"], 77.0),
        prototypes.build_face_centered_cubic(COMMON_ELEMENTS["Cu"], 128.0),
        prototypes.build_body_centered_cubic(COMMON_ELEMENTS["Fe"], 126.0),
        prototypes.build_hexagonal_close_packed(COMMON_ELEMENTS["Mg"], 160.0),
    )

    for cell in cells:
        for atom in cell.atoms:
            for value in atom.fractional_coordinates:
                assert 0.0 <= value < 1.0


def test_every_prototype_has_a_coordination_number():
    expected_prototypes = {
        "rock_salt",
        "cesium_chloride",
        "zinc_blende",
        "fluorite",
        "antifluorite",
        "diamond",
        "face_centered_cubic",
        "body_centered_cubic",
        "hexagonal_close_packed",
    }

    assert set(PROTOTYPE_COORDINATION_NUMBERS) == expected_prototypes
    assert PROTOTYPE_COORDINATION_NUMBERS[PROTOTYPE_HEXAGONAL_CLOSE_PACKED] == 12
