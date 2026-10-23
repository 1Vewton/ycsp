"""常数包的架构约束：只放数据，不反向依赖 domain / api。"""

import ast
from pathlib import Path

import app.constants


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
