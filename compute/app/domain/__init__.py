"""领域模型：晶体学数据结构、化学式解析与预测规则。

元素数据、结构原型数据、判据阈值与状态码都在 `app/constants/`；
本包只放类型与规则，不重复数据。
"""

from app.domain.composition import Composition, parse_formula
from app.domain.elements import COMMON_ELEMENTS, get_element, get_ionic_radius
from app.domain.notice import DomainError, Notice
from app.domain.prediction import Candidate, PredictionResult, predict_structure
from app.domain.structure import Atom, Element, UnitCell

__all__ = [
    "Atom",
    "Candidate",
    "COMMON_ELEMENTS",
    "Composition",
    "DomainError",
    "Element",
    "Notice",
    "PredictionResult",
    "UnitCell",
    "get_element",
    "get_ionic_radius",
    "parse_formula",
    "predict_structure",
]
