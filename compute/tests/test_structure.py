"""晶胞与原子数据结构的校验：非法值必须在构造时就报错。"""

import math

import pytest

from app.domain.structure import Atom, Element, UnitCell, angstrom_from_picometers

SODIUM = Element(symbol="Na", atomic_number=11)


def cubic_cell(**overrides: object) -> UnitCell:
    """构造一个合法的立方晶胞，测试里只覆盖需要变化的字段。"""
    parameters: dict[str, object] = {
        "edge_a_angstrom": 5.66,
        "edge_b_angstrom": 5.66,
        "edge_c_angstrom": 5.66,
        "angle_alpha_degree": 90.0,
        "angle_beta_degree": 90.0,
        "angle_gamma_degree": 90.0,
        "atoms": [Atom(SODIUM, (0.0, 0.0, 0.0))],
    }
    parameters.update(overrides)
    return UnitCell(**parameters)


def test_picometers_convert_to_angstrom():
    assert angstrom_from_picometers(283.0) == pytest.approx(2.83)


def test_valid_cell_keeps_its_values():
    cell = cubic_cell()

    assert cell.edge_a_angstrom == pytest.approx(5.66)
    assert cell.atoms[0].element.symbol == "Na"


def test_cell_atoms_default_to_empty_list():
    cell = UnitCell(5.0, 5.0, 5.0, 90.0, 90.0, 90.0)

    assert cell.atoms == []


def test_hexagonal_cell_accepts_120_degree_angle():
    cell = cubic_cell(angle_gamma_degree=120.0)

    assert cell.angle_gamma_degree == pytest.approx(120.0)


def test_element_symbol_cannot_be_empty():
    with pytest.raises(ValueError):
        Element(symbol="", atomic_number=11)


@pytest.mark.parametrize("atomic_number", [0, -1])
def test_atomic_number_must_be_positive(atomic_number: int):
    with pytest.raises(ValueError):
        Element(symbol="Na", atomic_number=atomic_number)


@pytest.mark.parametrize("coordinates", [(0.0, 0.0), (0.0, 0.0, 0.0, 0.0)])
def test_atom_needs_three_coordinates(coordinates: tuple[float, ...]):
    with pytest.raises(ValueError):
        Atom(SODIUM, coordinates)


@pytest.mark.parametrize("invalid_value", [math.inf, math.nan])
def test_atom_coordinates_must_be_finite(invalid_value: float):
    with pytest.raises(ValueError):
        Atom(SODIUM, (0.0, invalid_value, 0.0))


@pytest.mark.parametrize(
    "field_name", ["edge_a_angstrom", "edge_b_angstrom", "edge_c_angstrom"]
)
@pytest.mark.parametrize("invalid_length", [0.0, -1.0, math.nan])
def test_cell_edges_must_be_positive(field_name: str, invalid_length: float):
    with pytest.raises(ValueError):
        cubic_cell(**{field_name: invalid_length})


@pytest.mark.parametrize(
    "field_name", ["angle_alpha_degree", "angle_beta_degree", "angle_gamma_degree"]
)
@pytest.mark.parametrize("invalid_angle", [0.0, 180.0, -10.0, 200.0, math.nan])
def test_cell_angles_must_be_between_zero_and_180(field_name: str, invalid_angle: float):
    with pytest.raises(ValueError):
        cubic_cell(**{field_name: invalid_angle})
