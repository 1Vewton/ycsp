"""晶体结构原型的构建函数：把“结构型 + 半径”变成带原子坐标的晶胞。

结构型编号、分数坐标与几何比值都在 `app/constants/structures.py`；
这里只负责按硬球相切模型算晶胞边长。半径以皮米传入，晶胞边长以埃给出。
"""

from app.constants.structures import (
    ALL_TETRAHEDRAL_HOLE_SITES,
    BODY_CENTERED_SITES,
    CUBIC_BODY_DIAGONAL_RATIO,
    CUBIC_FACE_DIAGONAL_RATIO,
    FACE_CENTERED_SITES,
    HEXAGONAL_CLOSE_PACKED_SITES,
    HEXAGONAL_GAMMA_DEGREE,
    IDEAL_HEXAGONAL_AXIAL_RATIO,
    OCTAHEDRAL_HOLE_SITES,
    RIGHT_ANGLE_DEGREE,
    TETRAHEDRAL_HOLE_SITES_FIRST,
)
from app.domain.structure import Atom, Element, UnitCell, angstrom_from_picometers


def build_rock_salt(
    cation: Element,
    cation_radius_picometers: float,
    anion: Element,
    anion_radius_picometers: float,
) -> UnitCell:
    """氯化钠型：阴离子面心立方堆积、阳离子填满八面体空隙；沿棱相切 a = 2(r₊+r₋)。"""
    edge = 2 * _gap_angstrom(cation_radius_picometers, anion_radius_picometers)
    atoms = _atoms_at(FACE_CENTERED_SITES, cation) + _atoms_at(OCTAHEDRAL_HOLE_SITES, anion)
    return _cubic_cell(edge, atoms)


def build_cesium_chloride(
    cation: Element,
    cation_radius_picometers: float,
    anion: Element,
    anion_radius_picometers: float,
) -> UnitCell:
    """氯化铯型：阴离子在体心、阳离子在顶点；沿体对角线相切 √3·a = 2(r₊+r₋)。"""
    edge = 2 * _gap_angstrom(cation_radius_picometers, anion_radius_picometers)
    edge /= CUBIC_BODY_DIAGONAL_RATIO
    atoms = _atoms_at(((0.0, 0.0, 0.0),), cation) + _atoms_at(((0.5, 0.5, 0.5),), anion)
    return _cubic_cell(edge, atoms)


def build_zinc_blende(
    cation: Element,
    cation_radius_picometers: float,
    anion: Element,
    anion_radius_picometers: float,
) -> UnitCell:
    """闪锌矿型（立方 ZnS）：阴离子面心立方、阳离子占一半四面体空隙；键长 = √3·a/4。"""
    edge = 4 * _gap_angstrom(cation_radius_picometers, anion_radius_picometers)
    edge /= CUBIC_BODY_DIAGONAL_RATIO
    atoms = _atoms_at(FACE_CENTERED_SITES, anion) + _atoms_at(
        TETRAHEDRAL_HOLE_SITES_FIRST, cation
    )
    return _cubic_cell(edge, atoms)


def build_fluorite(
    cation: Element,
    cation_radius_picometers: float,
    anion: Element,
    anion_radius_picometers: float,
) -> UnitCell:
    """萤石型（CaF₂）：阳离子面心立方、阴离子填满 8 个四面体空隙；a = 4(r₊+r₋)/√3。"""
    edge = 4 * _gap_angstrom(cation_radius_picometers, anion_radius_picometers)
    edge /= CUBIC_BODY_DIAGONAL_RATIO
    atoms = _atoms_at(FACE_CENTERED_SITES, cation) + _atoms_at(
        ALL_TETRAHEDRAL_HOLE_SITES, anion
    )
    return _cubic_cell(edge, atoms)


def build_antifluorite(
    cation: Element,
    cation_radius_picometers: float,
    anion: Element,
    anion_radius_picometers: float,
) -> UnitCell:
    """反萤石型（Na₂O 等 2:1 化合物）：萤石型里的阴阳离子位置对调。"""
    edge = 4 * _gap_angstrom(cation_radius_picometers, anion_radius_picometers)
    edge /= CUBIC_BODY_DIAGONAL_RATIO
    atoms = _atoms_at(FACE_CENTERED_SITES, anion) + _atoms_at(
        ALL_TETRAHEDRAL_HOLE_SITES, cation
    )
    return _cubic_cell(edge, atoms)


def build_diamond(element: Element, radius_picometers: float) -> UnitCell:
    """金刚石型：每个原子与 4 个邻居相切，键长 = 2r，而键长又是体对角线的 1/4。"""
    bond_length = 2 * angstrom_from_picometers(radius_picometers)
    edge = 4 * bond_length / CUBIC_BODY_DIAGONAL_RATIO
    atoms = _atoms_at(FACE_CENTERED_SITES + TETRAHEDRAL_HOLE_SITES_FIRST, element)
    return _cubic_cell(edge, atoms)


def build_face_centered_cubic(element: Element, radius_picometers: float) -> UnitCell:
    """面心立方最密堆积：原子沿面对角线相切，√2·a = 4r。"""
    edge = 4 * angstrom_from_picometers(radius_picometers) / CUBIC_FACE_DIAGONAL_RATIO
    atoms = _atoms_at(FACE_CENTERED_SITES, element)
    return _cubic_cell(edge, atoms)


def build_body_centered_cubic(element: Element, radius_picometers: float) -> UnitCell:
    """体心立方堆积：原子沿体对角线相切，√3·a = 4r。"""
    edge = 4 * angstrom_from_picometers(radius_picometers) / CUBIC_BODY_DIAGONAL_RATIO
    atoms = _atoms_at(BODY_CENTERED_SITES, element)
    return _cubic_cell(edge, atoms)


def build_hexagonal_close_packed(element: Element, radius_picometers: float) -> UnitCell:
    """六方最密堆积：底面棱长 a = 2r，高 c = 1.633a。"""
    edge = 2 * angstrom_from_picometers(radius_picometers)
    height = IDEAL_HEXAGONAL_AXIAL_RATIO * edge
    atoms = _atoms_at(HEXAGONAL_CLOSE_PACKED_SITES, element)
    return UnitCell(
        edge_a_angstrom=edge,
        edge_b_angstrom=edge,
        edge_c_angstrom=height,
        angle_alpha_degree=RIGHT_ANGLE_DEGREE,
        angle_beta_degree=RIGHT_ANGLE_DEGREE,
        angle_gamma_degree=HEXAGONAL_GAMMA_DEGREE,
        atoms=atoms,
    )


def _gap_angstrom(first_radius_picometers: float, second_radius_picometers: float) -> float:
    """两个相切离子的半径之和，换算成埃。"""
    return angstrom_from_picometers(first_radius_picometers + second_radius_picometers)


def _atoms_at(sites: tuple[tuple[float, float, float], ...], element: Element) -> list[Atom]:
    """在这些分数坐标上各放一个同种原子。"""
    return [Atom(element, site) for site in sites]


def _cubic_cell(edge_angstrom: float, atoms: list[Atom]) -> UnitCell:
    """立方晶胞：三条棱等长、三个夹角都是 90°。"""
    return UnitCell(
        edge_a_angstrom=edge_angstrom,
        edge_b_angstrom=edge_angstrom,
        edge_c_angstrom=edge_angstrom,
        angle_alpha_degree=RIGHT_ANGLE_DEGREE,
        angle_beta_degree=RIGHT_ANGLE_DEGREE,
        angle_gamma_degree=RIGHT_ANGLE_DEGREE,
        atoms=atoms,
    )
