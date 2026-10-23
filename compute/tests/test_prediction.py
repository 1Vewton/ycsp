"""课本经典例子的预测结果（验收标准：结构与课本一致）。"""

from collections import Counter

import pytest

from app.constants.codes import (
    CAVEAT_CONFIRMED_BY_EXPERIMENT,
    CAVEAT_METAL_STRUCTURE_EXPERIMENTAL,
    CAVEAT_NEAR_BOUNDARY,
    CAVEAT_RADIUS_RATIO_EMPIRICAL,
    CAVEAT_STOICHIOMETRY_OUTSIDE_BAND,
    ERROR_COMPOSITION_NOT_SUPPORTED,
    ERROR_FORMULA_CHARGE_UNBALANCED,
    RULE_RADIUS_RATIO,
)
from app.domain.notice import DomainError
from app.domain.prediction import (
    FAMILY_COVALENT,
    FAMILY_IONIC,
    FAMILY_METALLIC,
    Candidate,
    predict_structure,
)
from app.constants.structures import (
    KNOWN_IONIC_STRUCTURES,
    PROTOTYPE_ANTIFLUORITE,
    PROTOTYPE_BODY_CENTERED_CUBIC,
    PROTOTYPE_CESIUM_CHLORIDE,
    PROTOTYPE_DIAMOND,
    PROTOTYPE_FACE_CENTERED_CUBIC,
    PROTOTYPE_FLUORITE,
    PROTOTYPE_HEXAGONAL_CLOSE_PACKED,
    PROTOTYPE_ROCK_SALT,
    PROTOTYPE_ZINC_BLENDE,
)

# 与课本晶胞边长比较时允许的误差（Å）：硬球模型本身就有几个百分点的偏差。
CELL_EDGE_TOLERANCE_ANGSTROM = 0.01


def only_candidate(formula: str) -> Candidate:
    """取唯一的候选结构；出现多个说明规则需要复核。"""
    result = predict_structure(formula)
    assert len(result.candidates) == 1, f"{formula} 不该给出多个候选结构"
    return result.candidates[0]


def atoms_per_symbol(candidate: Candidate) -> Counter[str]:
    return Counter(atom.element.symbol for atom in candidate.cell.atoms)


def caveat_codes(candidate: Candidate) -> set[str]:
    return {caveat.code for caveat in candidate.caveats}


def test_sodium_chloride_is_rock_salt():
    result = predict_structure("NaCl")
    candidate = result.candidates[0]

    assert result.family == FAMILY_IONIC
    assert candidate.prototype == PROTOTYPE_ROCK_SALT
    assert candidate.coordination_number == 6
    assert candidate.cell.edge_a_angstrom == pytest.approx(
        5.66, abs=CELL_EDGE_TOLERANCE_ANGSTROM
    )
    assert atoms_per_symbol(candidate) == {"Na": 4, "Cl": 4}
    assert candidate.rule.code == RULE_RADIUS_RATIO
    assert CAVEAT_CONFIRMED_BY_EXPERIMENT in caveat_codes(candidate)


def test_cesium_chloride_cell_edge():
    candidate = only_candidate("CsCl")

    assert candidate.prototype == PROTOTYPE_CESIUM_CHLORIDE
    assert candidate.coordination_number == 8
    assert candidate.cell.edge_a_angstrom == pytest.approx(
        4.02, abs=CELL_EDGE_TOLERANCE_ANGSTROM
    )
    assert atoms_per_symbol(candidate) == {"Cs": 1, "Cl": 1}


def test_calcium_fluoride_is_fluorite():
    candidate = only_candidate("CaF2")

    assert candidate.prototype == PROTOTYPE_FLUORITE
    assert candidate.cell.edge_a_angstrom == pytest.approx(
        5.38, abs=CELL_EDGE_TOLERANCE_ANGSTROM
    )
    assert atoms_per_symbol(candidate) == {"Ca": 4, "F": 8}


def test_sodium_oxide_is_antifluorite():
    """2:1 的离子晶体按个数比取反萤石型：阴离子做面心立方，阳离子填满四面体空隙。"""
    candidate = only_candidate("Na2O")

    assert candidate.prototype == PROTOTYPE_ANTIFLUORITE
    assert candidate.cell.edge_a_angstrom == pytest.approx(
        5.59, abs=CELL_EDGE_TOLERANCE_ANGSTROM
    )
    assert atoms_per_symbol(candidate) == {"Na": 8, "O": 4}
    assert CAVEAT_STOICHIOMETRY_OUTSIDE_BAND in caveat_codes(candidate)


def test_zinc_sulfide_is_zinc_blende_with_boundary_caveat():
    candidate = only_candidate("ZnS")

    assert candidate.prototype == PROTOTYPE_ZINC_BLENDE
    assert candidate.coordination_number == 4
    assert candidate.cell.edge_a_angstrom == pytest.approx(
        5.96, abs=CELL_EDGE_TOLERANCE_ANGSTROM
    )
    assert CAVEAT_NEAR_BOUNDARY in caveat_codes(candidate)


def test_diamond_cell_edge():
    result = predict_structure("C")
    candidate = result.candidates[0]

    assert result.family == FAMILY_COVALENT
    assert candidate.prototype == PROTOTYPE_DIAMOND
    assert candidate.cell.edge_a_angstrom == pytest.approx(
        3.56, abs=CELL_EDGE_TOLERANCE_ANGSTROM
    )
    assert len(candidate.cell.atoms) == 8


def test_silicon_carbide_is_zinc_blende():
    candidate = only_candidate("SiC")

    assert candidate.prototype == PROTOTYPE_ZINC_BLENDE
    assert candidate.cell.edge_a_angstrom == pytest.approx(
        4.34, abs=CELL_EDGE_TOLERANCE_ANGSTROM
    )
    assert atoms_per_symbol(candidate) == {"Si": 4, "C": 4}


def test_copper_is_face_centered_cubic():
    result = predict_structure("Cu")
    candidate = result.candidates[0]

    assert result.family == FAMILY_METALLIC
    assert candidate.prototype == PROTOTYPE_FACE_CENTERED_CUBIC
    assert candidate.coordination_number == 12
    assert candidate.cell.edge_a_angstrom == pytest.approx(
        3.62, abs=CELL_EDGE_TOLERANCE_ANGSTROM
    )
    assert len(candidate.cell.atoms) == 4
    assert CAVEAT_METAL_STRUCTURE_EXPERIMENTAL in caveat_codes(candidate)


def test_iron_is_body_centered_cubic():
    candidate = only_candidate("Fe")

    assert candidate.prototype == PROTOTYPE_BODY_CENTERED_CUBIC
    assert candidate.coordination_number == 8
    assert candidate.cell.edge_a_angstrom == pytest.approx(
        2.91, abs=CELL_EDGE_TOLERANCE_ANGSTROM
    )


def test_magnesium_is_hexagonal_close_packed():
    candidate = only_candidate("Mg")

    assert candidate.prototype == PROTOTYPE_HEXAGONAL_CLOSE_PACKED
    assert candidate.cell.edge_a_angstrom == pytest.approx(
        3.20, abs=CELL_EDGE_TOLERANCE_ANGSTROM
    )
    assert candidate.cell.edge_c_angstrom == pytest.approx(
        5.23, abs=CELL_EDGE_TOLERANCE_ANGSTROM
    )
    assert candidate.cell.angle_gamma_degree == pytest.approx(120.0)


def test_fractional_coordinates_stay_inside_cell():
    """分数坐标必须落在 [0, 1)，否则前端会把原子画到晶胞外面。"""
    for formula in ("NaCl", "CaF2", "C", "Cu", "Mg"):
        candidate = only_candidate(formula)
        for atom in candidate.cell.atoms:
            for value in atom.fractional_coordinates:
                assert 0.0 <= value < 1.0


@pytest.mark.parametrize(
    ("formula", "expected_prototype"), sorted(KNOWN_IONIC_STRUCTURES.items())
)
def test_known_ionic_compounds_match_experiment(formula: str, expected_prototype: str):
    """课本经典化合物必须给出实验测定的结构型（半径比规则只做对照）。"""
    candidate = only_candidate(formula)

    assert candidate.prototype == expected_prototype


def test_unlisted_ionic_pair_falls_back_to_radius_ratio():
    """不在经典表里的 1:1 组合按半径比规则给结果，并注明这是经验规律。

    ZnO 的实际结构是六方纤锌矿型，属 `compute/README.md` 写明的已知偏差；
    这里断言的是“规则会给出什么”，不是实验值。
    """
    candidate = only_candidate("ZnO")

    assert candidate.prototype == PROTOTYPE_ROCK_SALT
    assert candidate.rule.code == RULE_RADIUS_RATIO
    assert CAVEAT_RADIUS_RATIO_EMPIRICAL in caveat_codes(candidate)


@pytest.mark.parametrize(
    ("formula", "expected_code"),
    [
        ("H2O", ERROR_COMPOSITION_NOT_SUPPORTED),
        ("O", ERROR_COMPOSITION_NOT_SUPPORTED),
        ("Al2O3", ERROR_COMPOSITION_NOT_SUPPORTED),
        ("NaCl2", ERROR_FORMULA_CHARGE_UNBALANCED),
    ],
)
def test_uncovered_composition_reports_translatable_code(formula: str, expected_code: str):
    with pytest.raises(DomainError) as error_info:
        predict_structure(formula)

    assert error_info.value.notice.code == expected_code
