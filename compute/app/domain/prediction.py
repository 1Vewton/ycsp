"""按化学式预测晶体结构：先判断晶体类型，再按课本规则挑出结构原型。

三类判据：
- 只含一种金属元素 → 金属晶体，结构型取实验数据（METAL_STRUCTURE_TYPES）。
- 只含一种能形成金刚石型的元素（C / Si / Ge）→ 共价晶体（金刚石型）。
- 一种金属阳离子 + 一种阴离子 → 离子晶体，按 Pauling 半径比规则定配位数；
  课本经典化合物改用实验确定的结构型（KNOWN_IONIC_STRUCTURES），并说明差异。
其余情况暂不支持，返回 composition_not_supported。
"""

from dataclasses import dataclass
from math import gcd

from app.constants.codes import (
    CAVEAT_CONFIRMED_BY_EXPERIMENT,
    CAVEAT_LOW_IONIC_CHARACTER,
    CAVEAT_METAL_STRUCTURE_EXPERIMENTAL,
    CAVEAT_NEAR_BOUNDARY,
    CAVEAT_RADIUS_RATIO_EMPIRICAL,
    CAVEAT_STOICHIOMETRY_OUTSIDE_BAND,
    ERROR_COMPOSITION_NOT_SUPPORTED,
    ERROR_ELEMENT_DATA_MISSING,
    ERROR_FORMULA_CHARGE_UNBALANCED,
    ERROR_RADIUS_RATIO_TOO_SMALL,
    RULE_ANTIFLUORITE_STOICHIOMETRY,
    RULE_DIAMOND_COVALENT,
    RULE_EXPERIMENTAL_OVERRIDE,
    RULE_FLUORITE_STOICHIOMETRY,
    RULE_METAL_EXPERIMENTAL,
    RULE_RADIUS_RATIO,
    RULE_ZINC_BLENDE_COVALENT,
)
from app.constants.elements import ANION_CHARGES, CATION_CHARGES, DIAMOND_FORMING_ELEMENTS
from app.constants.prediction_rules import (
    IONIC_ELECTRONEGATIVITY_THRESHOLD,
    NEAR_BOUNDARY_TOLERANCE,
    RADIUS_RATIO_BOUNDARIES,
    RADIUS_RATIO_CUBIC_LOWER,
    RADIUS_RATIO_MAXIMUM,
    RADIUS_RATIO_OCTAHEDRAL_LOWER,
    RADIUS_RATIO_TETRAHEDRAL_LOWER,
)
from app.constants.structures import (
    KNOWN_IONIC_STRUCTURES,
    METAL_STRUCTURE_TYPES,
    PROTOTYPE_ANTIFLUORITE,
    PROTOTYPE_BODY_CENTERED_CUBIC,
    PROTOTYPE_CESIUM_CHLORIDE,
    PROTOTYPE_COORDINATION_NUMBERS,
    PROTOTYPE_DIAMOND,
    PROTOTYPE_FACE_CENTERED_CUBIC,
    PROTOTYPE_FLUORITE,
    PROTOTYPE_HEXAGONAL_CLOSE_PACKED,
    PROTOTYPE_ROCK_SALT,
    PROTOTYPE_ZINC_BLENDE,
)
from app.domain import prototypes
from app.domain.composition import Composition, parse_formula
from app.domain.elements import (
    COMMON_ELEMENTS,
    find_covalent_radius,
    find_electronegativity,
    find_ionic_radius,
    find_metallic_radius,
)
from app.domain.notice import DomainError, Notice
from app.domain.structure import UnitCell

FAMILY_IONIC = "ionic"
FAMILY_COVALENT = "covalent"
FAMILY_METALLIC = "metallic"

# 结构型 → 生成晶胞的函数（阳离子在前，半径为皮米）。
_IONIC_CELL_BUILDERS = {
    PROTOTYPE_ROCK_SALT: prototypes.build_rock_salt,
    PROTOTYPE_CESIUM_CHLORIDE: prototypes.build_cesium_chloride,
    PROTOTYPE_ZINC_BLENDE: prototypes.build_zinc_blende,
    PROTOTYPE_FLUORITE: prototypes.build_fluorite,
    PROTOTYPE_ANTIFLUORITE: prototypes.build_antifluorite,
}

_METAL_CELL_BUILDERS = {
    PROTOTYPE_FACE_CENTERED_CUBIC: prototypes.build_face_centered_cubic,
    PROTOTYPE_BODY_CENTERED_CUBIC: prototypes.build_body_centered_cubic,
    PROTOTYPE_HEXAGONAL_CLOSE_PACKED: prototypes.build_hexagonal_close_packed,
}


@dataclass(frozen=True)
class Candidate:
    """一个候选结构：结构型 + 判断依据 + 附注 + 具体晶胞。"""

    prototype: str
    coordination_number: int
    rule: Notice
    caveats: tuple[Notice, ...]
    cell: UnitCell


@dataclass(frozen=True)
class PredictionResult:
    """一次预测的结果。"""

    formula: str
    family: str
    candidates: tuple[Candidate, ...]


@dataclass(frozen=True)
class IonicPair:
    """一对阴阳离子，以及推算必需的中间量（半径为皮米）。"""

    composition: Composition
    cation_symbol: str
    cation_count: int
    cation_radius_picometers: float
    anion_symbol: str
    anion_count: int
    anion_radius_picometers: float
    electronegativity_difference: float
    formula_key: str

    @property
    def radius_ratio(self) -> float:
        """阳离子半径与阴离子半径之比（无量纲），半径比规则的输入。"""
        return self.cation_radius_picometers / self.anion_radius_picometers


def predict_structure(formula: str) -> PredictionResult:
    """按化学式给出候选晶体结构；无法处理时抛出 DomainError（只带 code 与 params）。"""
    composition = parse_formula(formula)
    element_count = len(composition.counts)
    if element_count == 1:
        return _predict_single_element(composition)
    if element_count == 2:
        return _predict_two_elements(composition)
    raise _unsupported(composition)


def _predict_single_element(composition: Composition) -> PredictionResult:
    (symbol,) = composition.counts
    if symbol in METAL_STRUCTURE_TYPES:
        return _predict_metal(composition, symbol)
    if symbol in DIAMOND_FORMING_ELEMENTS:
        return _predict_nonmetal_diamond(composition, symbol)
    raise _unsupported(composition)


def _predict_metal(composition: Composition, symbol: str) -> PredictionResult:
    """金属晶体：结构型直接取实验数据，并明确告诉学生这一点。"""
    radius_picometers = _require_metallic_radius(symbol)
    prototype = METAL_STRUCTURE_TYPES[symbol]
    cell = _METAL_CELL_BUILDERS[prototype](COMMON_ELEMENTS[symbol], radius_picometers)
    return _result(
        composition=composition,
        family=FAMILY_METALLIC,
        prototype=prototype,
        rule=Notice(RULE_METAL_EXPERIMENTAL),
        caveats=(Notice(CAVEAT_METAL_STRUCTURE_EXPERIMENTAL),),
        cell=cell,
    )


def _predict_nonmetal_diamond(composition: Composition, symbol: str) -> PredictionResult:
    """金刚石型共价晶体：相邻原子以共价键相连，键长取共价半径的 2 倍。"""
    radius_picometers = _require_covalent_radius(symbol)
    cell = prototypes.build_diamond(COMMON_ELEMENTS[symbol], radius_picometers)
    return _result(
        composition=composition,
        family=FAMILY_COVALENT,
        prototype=PROTOTYPE_DIAMOND,
        rule=Notice(RULE_DIAMOND_COVALENT),
        caveats=(),
        cell=cell,
    )


def _predict_two_elements(composition: Composition) -> PredictionResult:
    if _forms_diamond_pair(composition):
        return _predict_diamond_pair(composition)
    return _predict_ionic(composition)


def _forms_diamond_pair(composition: Composition) -> bool:
    """两种元素都能形成金刚石型、且个数相等（如 SiC）。"""
    counts = composition.counts
    if len(counts) != 2 or len(set(counts.values())) != 1:
        return False
    return all(symbol in DIAMOND_FORMING_ELEMENTS for symbol in counts)


def _predict_diamond_pair(composition: Composition) -> PredictionResult:
    """两种非金属按闪锌矿（立方 ZnS）型交替排布，电负性小的一方占锌的位置。"""
    less_electronegative, more_electronegative = sorted(
        composition.counts, key=_require_electronegativity
    )
    cell = prototypes.build_zinc_blende(
        COMMON_ELEMENTS[less_electronegative],
        _require_covalent_radius(less_electronegative),
        COMMON_ELEMENTS[more_electronegative],
        _require_covalent_radius(more_electronegative),
    )
    difference = abs(
        _require_electronegativity(less_electronegative)
        - _require_electronegativity(more_electronegative)
    )
    return _result(
        composition=composition,
        family=FAMILY_COVALENT,
        prototype=PROTOTYPE_ZINC_BLENDE,
        rule=Notice(RULE_ZINC_BLENDE_COVALENT, {"electronegativity_difference": difference}),
        caveats=(),
        cell=cell,
    )


def _predict_ionic(composition: Composition) -> PredictionResult:
    """离子晶体：约分后的阴阳离子个数比决定用哪条规则。"""
    pair = _classify_ionic_pair(composition)
    stoichiometry = (pair.cation_count, pair.anion_count)
    if stoichiometry == (1, 1):
        return _predict_one_to_one(pair)
    if stoichiometry == (1, 2):
        return _predict_fluorite(pair, antifluorite=False)
    if stoichiometry == (2, 1):
        return _predict_fluorite(pair, antifluorite=True)
    raise _unsupported(composition)


def _classify_ionic_pair(composition: Composition) -> IonicPair:
    """找出“金属阳离子 + 阴离子”并检查正负电荷是否抵消。"""
    cation_symbol = _single_symbol_in(composition, CATION_CHARGES)
    anion_symbol = _single_symbol_in(composition, ANION_CHARGES)
    if cation_symbol is None or anion_symbol is None:
        raise _unsupported(composition)

    cation_charge = CATION_CHARGES[cation_symbol]
    anion_charge = ANION_CHARGES[anion_symbol]
    cation_count = composition.counts[cation_symbol]
    anion_count = composition.counts[anion_symbol]
    if cation_count * cation_charge != anion_count * anion_charge:
        raise DomainError(
            Notice(ERROR_FORMULA_CHARGE_UNBALANCED, {"formula": composition.formula})
        )

    divisor = gcd(cation_count, anion_count)
    reduced_cation_count = cation_count // divisor
    reduced_anion_count = anion_count // divisor
    return IonicPair(
        composition=composition,
        cation_symbol=cation_symbol,
        cation_count=reduced_cation_count,
        cation_radius_picometers=_require_ionic_radius(cation_symbol, cation_charge),
        anion_symbol=anion_symbol,
        anion_count=reduced_anion_count,
        anion_radius_picometers=_require_ionic_radius(anion_symbol, -anion_charge),
        electronegativity_difference=abs(
            _require_electronegativity(cation_symbol) - _require_electronegativity(anion_symbol)
        ),
        formula_key=(
            f"{cation_symbol}{_count_suffix(reduced_cation_count)}"
            f"{anion_symbol}{_count_suffix(reduced_anion_count)}"
        ),
    )


def _single_symbol_in(composition: Composition, charge_table: dict[str, int]) -> str | None:
    """在化学式里找出恰好一个属于 charge_table 的元素；没有或多个时返回 None。"""
    matches = [symbol for symbol in composition.counts if symbol in charge_table]
    return matches[0] if len(matches) == 1 else None


def _count_suffix(count: int) -> str:
    """化学式里下标为 1 时省略不写。"""
    return "" if count == 1 else str(count)


def _predict_one_to_one(pair: IonicPair) -> PredictionResult:
    """1:1 离子晶体：半径比规则给配位数，课本经典化合物按实验数据校正。"""
    rule_prototype = _select_by_radius_ratio(pair.radius_ratio)
    experimental_prototype = KNOWN_IONIC_STRUCTURES.get(pair.formula_key)
    prototype = experimental_prototype or rule_prototype
    cell = _IONIC_CELL_BUILDERS[prototype](
        COMMON_ELEMENTS[pair.cation_symbol],
        pair.cation_radius_picometers,
        COMMON_ELEMENTS[pair.anion_symbol],
        pair.anion_radius_picometers,
    )
    return _result(
        composition=pair.composition,
        family=FAMILY_IONIC,
        prototype=prototype,
        rule=_one_to_one_rule_notice(pair, rule_prototype, experimental_prototype),
        caveats=_one_to_one_caveats(pair, rule_prototype, experimental_prototype),
        cell=cell,
    )


def _one_to_one_rule_notice(
    pair: IonicPair, rule_prototype: str, experimental_prototype: str | None
) -> Notice:
    """说明“这个结构是怎么来的”：规则直接给出，还是按实验数据覆盖了规则。"""
    if experimental_prototype is not None and experimental_prototype != rule_prototype:
        return Notice(
            RULE_EXPERIMENTAL_OVERRIDE,
            {
                "ratio": pair.radius_ratio,
                "rule_prototype": rule_prototype,
                "rule_coordination_number": PROTOTYPE_COORDINATION_NUMBERS[rule_prototype],
            },
        )
    return Notice(
        RULE_RADIUS_RATIO,
        {
            "ratio": pair.radius_ratio,
            "coordination_number": PROTOTYPE_COORDINATION_NUMBERS[rule_prototype],
        },
    )


def _one_to_one_caveats(
    pair: IonicPair, rule_prototype: str, experimental_prototype: str | None
) -> tuple[Notice, ...]:
    """列出 1:1 离子晶体结果的局限：经验规律、边界、共价性。"""
    caveats: list[Notice] = []
    if experimental_prototype is None:
        caveats.append(Notice(CAVEAT_RADIUS_RATIO_EMPIRICAL))
    elif experimental_prototype == rule_prototype:
        caveats.append(Notice(CAVEAT_CONFIRMED_BY_EXPERIMENT))

    near_boundary = _near_boundary_notice(pair.radius_ratio)
    if near_boundary is not None:
        caveats.append(near_boundary)

    low_ionic_character = _low_ionic_character_notice(pair.electronegativity_difference)
    if low_ionic_character is not None:
        caveats.append(low_ionic_character)
    return tuple(caveats)


def _select_by_radius_ratio(radius_ratio: float) -> str:
    """Pauling 半径比规则：比值决定阳离子的配位数，进而决定结构型。"""
    if radius_ratio < RADIUS_RATIO_TETRAHEDRAL_LOWER:
        raise DomainError(Notice(ERROR_RADIUS_RATIO_TOO_SMALL, {"ratio": radius_ratio}))
    if radius_ratio < RADIUS_RATIO_OCTAHEDRAL_LOWER:
        return PROTOTYPE_ZINC_BLENDE
    if radius_ratio < RADIUS_RATIO_CUBIC_LOWER:
        return PROTOTYPE_ROCK_SALT
    return PROTOTYPE_CESIUM_CHLORIDE


def _near_boundary_notice(radius_ratio: float) -> Notice | None:
    """比值贴近分界时，另一种配位数也可能，返回一条提示。"""
    for boundary, prototype_below, prototype_above in RADIUS_RATIO_BOUNDARIES:
        if abs(radius_ratio - boundary) <= NEAR_BOUNDARY_TOLERANCE:
            alternative = prototype_above if radius_ratio < boundary else prototype_below
            return Notice(
                CAVEAT_NEAR_BOUNDARY,
                {
                    "ratio": radius_ratio,
                    "boundary": boundary,
                    "alternative_prototype": alternative,
                },
            )
    return None


def _low_ionic_character_notice(electronegativity_difference: float) -> Notice | None:
    """电负性差偏小时键的共价性明显，提醒学生这已接近共价晶体。"""
    if electronegativity_difference >= IONIC_ELECTRONEGATIVITY_THRESHOLD:
        return None
    return Notice(
        CAVEAT_LOW_IONIC_CHARACTER,
        {
            "difference": electronegativity_difference,
            "threshold": IONIC_ELECTRONEGATIVITY_THRESHOLD,
        },
    )


def _predict_fluorite(pair: IonicPair, antifluorite: bool) -> PredictionResult:
    """1:2 / 2:1 离子晶体：大离子面心立方、小离子填满四面体空隙（萤石型或反萤石型）。

    这类结构由离子个数比决定，所以不用半径比筛选（Na₂O 的比值 0.729 就略低于 0.732），
    比值偏离理想区间时用附注提醒学生。
    """
    prototype = PROTOTYPE_ANTIFLUORITE if antifluorite else PROTOTYPE_FLUORITE
    rule_code = (
        RULE_ANTIFLUORITE_STOICHIOMETRY if antifluorite else RULE_FLUORITE_STOICHIOMETRY
    )
    cell = _IONIC_CELL_BUILDERS[prototype](
        COMMON_ELEMENTS[pair.cation_symbol],
        pair.cation_radius_picometers,
        COMMON_ELEMENTS[pair.anion_symbol],
        pair.anion_radius_picometers,
    )
    return _result(
        composition=pair.composition,
        family=FAMILY_IONIC,
        prototype=prototype,
        rule=Notice(
            rule_code,
            {
                "ratio": pair.radius_ratio,
                "cation_count": pair.cation_count,
                "anion_count": pair.anion_count,
            },
        ),
        caveats=_fluorite_caveats(pair, prototype),
        cell=cell,
    )


def _fluorite_caveats(pair: IonicPair, prototype: str) -> tuple[Notice, ...]:
    """萤石型结果的附注：实验是否吻合、比值是否偏离、键的共价性。"""
    caveats: list[Notice] = []
    if KNOWN_IONIC_STRUCTURES.get(pair.formula_key) == prototype:
        caveats.append(Notice(CAVEAT_CONFIRMED_BY_EXPERIMENT))
    if not RADIUS_RATIO_CUBIC_LOWER <= pair.radius_ratio <= RADIUS_RATIO_MAXIMUM:
        caveats.append(Notice(CAVEAT_STOICHIOMETRY_OUTSIDE_BAND, {"ratio": pair.radius_ratio}))

    low_ionic_character = _low_ionic_character_notice(pair.electronegativity_difference)
    if low_ionic_character is not None:
        caveats.append(low_ionic_character)
    return tuple(caveats)


def _result(
    composition: Composition,
    family: str,
    prototype: str,
    rule: Notice,
    caveats: tuple[Notice, ...],
    cell: UnitCell,
) -> PredictionResult:
    """把“选好的结构型 + 晶胞”包装成一次预测结果。"""
    candidate = Candidate(
        prototype=prototype,
        coordination_number=PROTOTYPE_COORDINATION_NUMBERS[prototype],
        rule=rule,
        caveats=caveats,
        cell=cell,
    )
    return PredictionResult(
        formula=composition.formula,
        family=family,
        candidates=(candidate,),
    )


def _require_ionic_radius(symbol: str, charge: int) -> float:
    radius = find_ionic_radius(symbol, charge)
    if radius is None:
        raise DomainError(Notice(ERROR_ELEMENT_DATA_MISSING, {"symbol": symbol}))
    return radius


def _require_covalent_radius(symbol: str) -> float:
    radius = find_covalent_radius(symbol)
    if radius is None:
        raise DomainError(Notice(ERROR_ELEMENT_DATA_MISSING, {"symbol": symbol}))
    return radius


def _require_metallic_radius(symbol: str) -> float:
    radius = find_metallic_radius(symbol)
    if radius is None:
        raise DomainError(Notice(ERROR_ELEMENT_DATA_MISSING, {"symbol": symbol}))
    return radius


def _require_electronegativity(symbol: str) -> float:
    value = find_electronegativity(symbol)
    if value is None:
        raise DomainError(Notice(ERROR_ELEMENT_DATA_MISSING, {"symbol": symbol}))
    return value


def _unsupported(composition: Composition) -> DomainError:
    """本平台不覆盖这类化学式：给出可翻译的错误，而不是返回瞎猜的结构。"""
    return DomainError(Notice(ERROR_COMPOSITION_NOT_SUPPORTED, {"formula": composition.formula}))

