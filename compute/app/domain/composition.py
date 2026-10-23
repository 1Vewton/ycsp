"""化学式解析：把 "CaF2"、"Ca(OH)2" 这样的式子解析成“元素符号 → 原子个数”。"""

import re
from dataclasses import dataclass

from app.constants.codes import (
    ERROR_ELEMENT_NOT_SUPPORTED,
    ERROR_FORMULA_EMPTY,
    ERROR_FORMULA_INVALID_CHARACTER,
    ERROR_FORMULA_UNBALANCED_PARENTHESES,
    ERROR_FORMULA_ZERO_COUNT,
)
from app.domain.elements import find_element
from app.domain.notice import DomainError, Notice

# 元素符号：一个大写字母，可跟一个小写字母（如 "Na"、"Cl"）。
ELEMENT_SYMBOL_PATTERN = re.compile(r"[A-Z][a-z]?")
OPENING_BRACKET = "("
CLOSING_BRACKET = ")"


@dataclass(frozen=True)
class Composition:
    """解析后的化学式。"""

    formula: str  # 用户输入的式子（已去掉首尾空格）
    counts: dict[str, int]  # 元素符号 → 原子个数


def parse_formula(formula: str) -> Composition:
    """解析化学式；写法有问题时抛出 DomainError，由前端翻成学生看得懂的提示。"""
    cleaned = formula.strip()
    if not cleaned:
        raise DomainError(Notice(ERROR_FORMULA_EMPTY))

    # 栈顶是当前所在的括号组；遇到 ")" 就把组内原子按括号后的下标合并到上一层。
    group_stack: list[dict[str, int]] = [{}]
    position = 0
    while position < len(cleaned):
        character = cleaned[position]
        if character == OPENING_BRACKET:
            group_stack.append({})
            position += 1
        elif character == CLOSING_BRACKET:
            position = _close_group(cleaned, position, group_stack)
        else:
            position = _read_element(cleaned, position, group_stack[-1])

    if len(group_stack) > 1:
        raise DomainError(Notice(ERROR_FORMULA_UNBALANCED_PARENTHESES, {"formula": cleaned}))
    counts = group_stack[0]
    if not counts:
        raise DomainError(Notice(ERROR_FORMULA_EMPTY))
    return Composition(formula=cleaned, counts=counts)


def _read_element(formula: str, position: int, counts: dict[str, int]) -> int:
    """读一个元素符号及其下标并累加进 counts，返回读到的下一个位置。"""
    match = ELEMENT_SYMBOL_PATTERN.match(formula, position)
    if match is None:
        raise _invalid_character(formula, position)

    symbol = match.group()
    if find_element(symbol) is None:
        raise DomainError(
            Notice(ERROR_ELEMENT_NOT_SUPPORTED, {"symbol": symbol, "position": position + 1})
        )

    count, next_position = _read_count(formula, position + len(symbol))
    counts[symbol] = counts.get(symbol, 0) + count
    return next_position


def _close_group(formula: str, position: int, group_stack: list[dict[str, int]]) -> int:
    """结束一个括号组，返回读到的下一个位置。"""
    if len(group_stack) == 1:
        raise DomainError(Notice(ERROR_FORMULA_UNBALANCED_PARENTHESES, {"formula": formula}))

    group = group_stack.pop()
    multiplier, next_position = _read_count(formula, position + 1)
    parent_counts = group_stack[-1]
    for symbol, count in group.items():
        parent_counts[symbol] = parent_counts.get(symbol, 0) + count * multiplier
    return next_position


def _read_count(formula: str, position: int) -> tuple[int, int]:
    """读下标数字；没有写数字时按 1 个算。"""
    end = position
    while end < len(formula) and formula[end].isdigit():
        end += 1
    if end == position:
        return 1, position

    count = int(formula[position:end])
    if count == 0:
        raise DomainError(Notice(ERROR_FORMULA_ZERO_COUNT, {"formula": formula}))
    return count, end


def _invalid_character(formula: str, position: int) -> DomainError:
    """把当前位置的字符报成“化学式里不该出现的字符”（位置从 1 开始数）。"""
    return DomainError(
        Notice(
            ERROR_FORMULA_INVALID_CHARACTER,
            {"character": formula[position], "position": position + 1},
        )
    )
