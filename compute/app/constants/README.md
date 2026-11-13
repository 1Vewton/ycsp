> English | [简体中文](./README.zh-CN.md)

# Data processing layer · Constants package (compute/app/constants)

Data and pure constants are centralized here: `domain` contains only types and rules, and looks up data here.

| File | Purpose |
| --- | --- |
| `elements.py` | Element properties: atomic number, Pauling electronegativity, ionic / covalent / metallic radii, diamond-forming elements, common ionic charges |
| `structures.py` | Structure prototypes: IDs, coordination numbers, fractional coordinates, geometric ratios, measured structure tables for metals and classic textbook ionic compounds, and measured cell parameters for wurtzite / rutile |
| `prediction_rules.py` | Thresholds: Pauling radius-ratio boundaries, near-boundary tolerance, electronegativity-difference threshold |
| `codes.py` | All status codes: response status, result rationale, result caveats, errors |

## Conventions

- Only data and pure constants, no business logic; **do not** import `app.domain` / `app.api`
  (that would reverse the "data ← logic" direction and risk circular imports; `tests/test_constants.py` checks this).
- Every number states its unit or physical meaning: lengths in picometers (pm), angles in degrees (°), ratios dimensionless.
- When adding data, follow the same format and add a consistency assertion in `tests/test_element_data.py`,
  so problems such as "element added but radius forgotten" do not slip through.
