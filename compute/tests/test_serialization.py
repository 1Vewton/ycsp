"""接口序列化：字段名、展示精度，以及应答里不出现任何写死的文案。"""

from typing import Any

import pytest

from app.api.serialization import (
    cell_parameters_to_dict,
    notice_to_dict,
    prediction_to_dict,
)
from app.constants.codes import RULE_RADIUS_RATIO
from app.domain.notice import Notice
from app.domain.prediction import predict_structure


def collect_floats(node: Any) -> list[float]:
    """把嵌套字典 / 列表里的浮点数全部收集出来。"""
    if isinstance(node, float):
        return [node]
    if isinstance(node, dict):
        return [value for child in node.values() for value in collect_floats(child)]
    if isinstance(node, list):
        return [value for child in node for value in collect_floats(child)]
    return []


def test_notice_keeps_code_and_rounds_only_floats():
    notice = Notice(
        RULE_RADIUS_RATIO,
        {"ratio": 0.5635359116022099, "coordination_number": 6, "rule_prototype": "rock_salt"},
    )

    payload = notice_to_dict(notice)

    assert payload["code"] == RULE_RADIUS_RATIO
    assert payload["params"]["ratio"] == 0.564
    assert payload["params"]["coordination_number"] == 6
    assert payload["params"]["rule_prototype"] == "rock_salt"


def test_cell_parameters_state_their_units_in_key_names():
    candidate = predict_structure("NaCl").candidates[0]

    payload = cell_parameters_to_dict(candidate.cell)

    assert set(payload) == {
        "edge_a_angstrom",
        "edge_b_angstrom",
        "edge_c_angstrom",
        "angle_alpha_degree",
        "angle_beta_degree",
        "angle_gamma_degree",
    }
    assert payload["edge_a_angstrom"] == pytest.approx(5.66)
    assert payload["angle_alpha_degree"] == pytest.approx(90.0)


def test_prediction_payload_matches_the_frontend_contract():
    payload = prediction_to_dict(predict_structure("NaCl"))

    assert payload["formula"] == "NaCl"
    assert payload["family"] == "ionic"

    candidate = payload["candidates"][0]
    assert set(candidate) == {
        "prototype",
        "coordination_number",
        "rule",
        "caveats",
        "cell",
        "atoms",
    }
    first_atom = candidate["atoms"][0]
    assert set(first_atom) == {"symbol", "fractional_coordinates"}
    assert len(first_atom["fractional_coordinates"]) == 3


def test_candidate_without_caveats_serializes_an_empty_list():
    payload = prediction_to_dict(predict_structure("C"))

    assert payload["candidates"][0]["caveats"] == []


def test_every_float_in_the_payload_is_rounded_for_display():
    """漏掉舍入的浮点会直接出现在学生的说明文字里（如 0.5635359116…）。"""
    for formula in ("NaCl", "ZnS", "CaF2", "SiC", "Mg"):
        payload = prediction_to_dict(predict_structure(formula))
        for value in collect_floats(payload):
            assert round(value, 3) == value, f"{formula} 里出现了未舍入的浮点数 {value}"
