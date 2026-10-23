"""化学式解析的单元测试。"""

import pytest

from app.constants.codes import (
    ERROR_ELEMENT_NOT_SUPPORTED,
    ERROR_FORMULA_EMPTY,
    ERROR_FORMULA_INVALID_CHARACTER,
    ERROR_FORMULA_UNBALANCED_PARENTHESES,
    ERROR_FORMULA_ZERO_COUNT,
)
from app.domain.composition import parse_formula
from app.domain.notice import DomainError


def parse_counts(formula: str) -> dict[str, int]:
    """解析后只取“元素 → 个数”部分，方便断言。"""
    return parse_formula(formula).counts


def test_parse_one_to_one_formula():
    assert parse_counts("NaCl") == {"Na": 1, "Cl": 1}


def test_parse_formula_with_subscripts():
    assert parse_counts("CaF2") == {"Ca": 1, "F": 2}
    assert parse_counts("Al2O3") == {"Al": 2, "O": 3}


def test_parse_formula_with_brackets():
    assert parse_counts("Ca(OH)2") == {"Ca": 1, "O": 2, "H": 2}


def test_parse_single_element():
    assert parse_counts("C") == {"C": 1}


def test_parse_trims_surrounding_spaces():
    assert parse_formula(" NaCl ").formula == "NaCl"


@pytest.mark.parametrize(
    ("formula", "expected_code"),
    [
        ("", ERROR_FORMULA_EMPTY),
        ("   ", ERROR_FORMULA_EMPTY),
        ("NaCl)", ERROR_FORMULA_UNBALANCED_PARENTHESES),
        ("Ca(OH2", ERROR_FORMULA_UNBALANCED_PARENTHESES),
        ("2NaCl", ERROR_FORMULA_INVALID_CHARACTER),
        ("Na Cl", ERROR_FORMULA_INVALID_CHARACTER),
        ("Na0Cl", ERROR_FORMULA_ZERO_COUNT),
        ("()", ERROR_FORMULA_EMPTY),
        ("Xx", ERROR_ELEMENT_NOT_SUPPORTED),
    ],
)
def test_invalid_formula_reports_translatable_code(formula: str, expected_code: str):
    """写法有问题时只返回 code，让前端翻成学生看得懂的提示。"""
    with pytest.raises(DomainError) as error_info:
        parse_formula(formula)

    assert error_info.value.notice.code == expected_code
