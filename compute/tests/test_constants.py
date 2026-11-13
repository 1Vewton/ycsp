"""常数包的架构约束：只放数据，不反向依赖 domain / api。"""

import ast
from pathlib import Path

import app.constants
from app.constants.structures import (
    KNOWN_IONIC_STRUCTURES,
    POLYMORPHIC_WURTZITE_FORMULAS,
    PROTOTYPE_RUTILE,
    PROTOTYPE_WURTZITE,
    RUTILE_CELL_PARAMETERS,
    WURTZITE_CELL_PARAMETERS,
)


def test_constants_package_does_not_import_domain_or_api():
    """常数包若依赖 domain / api，就会和“数据 ← 逻辑”的方向相反，容易形成循环导入。"""
    package_directory = Path(app.constants.__file__).parent
    offenders: list[str] = []

    for module_path in sorted(package_directory.glob("*.py")):
        module_tree = ast.parse(module_path.read_text(encoding="utf-8"))
        for node in ast.walk(module_tree):
            if not isinstance(node, ast.ImportFrom) or not node.module:
                continue
            if node.module.startswith(("app.domain", "app.api")):
                offenders.append(f"{module_path.name} → {node.module}")

    assert not offenders, f"常数包不应依赖 domain / api：{offenders}"


def test_every_wurtzite_and_rutile_compound_has_cell_parameters():
    """经典表里标成纤锌矿 / 金红石的化合物，必须同时有实验晶胞参数。"""
    wurtzite_formulas = {
        formula
        for formula, prototype in KNOWN_IONIC_STRUCTURES.items()
        if prototype == PROTOTYPE_WURTZITE
    }
    rutile_formulas = {
        formula
        for formula, prototype in KNOWN_IONIC_STRUCTURES.items()
        if prototype == PROTOTYPE_RUTILE
    }

    assert wurtzite_formulas == set(WURTZITE_CELL_PARAMETERS)
    assert rutile_formulas == set(RUTILE_CELL_PARAMETERS)
    assert POLYMORPHIC_WURTZITE_FORMULAS <= wurtzite_formulas
    assert all(
        0.0 < parameters.anion_position_u < 0.5
        for parameters in WURTZITE_CELL_PARAMETERS.values()
    )
    assert all(
        0.0 < parameters.anion_position_u < 0.5
        for parameters in RUTILE_CELL_PARAMETERS.values()
    )
