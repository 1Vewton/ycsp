"""把预测结果转成接口应答里的普通字典（只有数值与语言无关的 code）。"""

from app.api.display import round_for_display
from app.domain.notice import Notice
from app.domain.prediction import Candidate, PredictionResult
from app.domain.structure import Atom, UnitCell


def prediction_to_dict(result: PredictionResult) -> dict[str, object]:
    """预测结果 → 应答里的 prediction 字段。"""
    return {
        "formula": result.formula,
        "family": result.family,
        "candidates": [candidate_to_dict(candidate) for candidate in result.candidates],
    }


def candidate_to_dict(candidate: Candidate) -> dict[str, object]:
    """一个候选结构 → 晶胞参数 + 原子坐标，前端可直接拿去画图。"""
    return {
        "prototype": candidate.prototype,
        "coordination_number": candidate.coordination_number,
        "rule": notice_to_dict(candidate.rule),
        "caveats": [notice_to_dict(caveat) for caveat in candidate.caveats],
        "cell": cell_parameters_to_dict(candidate.cell),
        "atoms": [atom_to_dict(atom) for atom in candidate.cell.atoms],
    }


def notice_to_dict(notice: Notice) -> dict[str, object]:
    """说明只传出 code 与数值参数；浮点数按展示精度舍入后再填进文案。"""
    return {
        "code": notice.code,
        "params": {
            name: round_for_display(value) if isinstance(value, float) else value
            for name, value in notice.params.items()
        },
    }


def cell_parameters_to_dict(cell: UnitCell) -> dict[str, float]:
    """晶胞的 6 个参数：三条棱长单位 Å，三个夹角单位 °。"""
    return {
        "edge_a_angstrom": round_for_display(cell.edge_a_angstrom),
        "edge_b_angstrom": round_for_display(cell.edge_b_angstrom),
        "edge_c_angstrom": round_for_display(cell.edge_c_angstrom),
        "angle_alpha_degree": round_for_display(cell.angle_alpha_degree),
        "angle_beta_degree": round_for_display(cell.angle_beta_degree),
        "angle_gamma_degree": round_for_display(cell.angle_gamma_degree),
    }


def atom_to_dict(atom: Atom) -> dict[str, object]:
    """一个原子：元素符号 + 分数坐标（相对棱长的比例，无量纲）。"""
    return {
        "symbol": atom.element.symbol,
        "fractional_coordinates": [
            round_for_display(value) for value in atom.fractional_coordinates
        ],
    }
