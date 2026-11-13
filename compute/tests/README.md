> English | [简体中文](./README.zh-CN.md)

# Data processing layer · Tests (compute/tests)

Run with `uv run pytest` from the `compute/` root; the command matches CI.

| File | Coverage |
| --- | --- |
| `test_composition.py` | Formula parsing: subscripts, brackets, spaces, and the code returned for each malformed input |
| `test_structure.py` | Element / atom / unit-cell validation: non-positive edges, out-of-range angles, non-finite coordinates |
| `test_prototypes.py` | Edge-length formulas, atom counts, and fractional-coordinate ranges for the 11 structure prototypes |
| `test_prediction.py` | Structure type and unit-cell edges for classic textbook examples (NaCl / CsCl / CaF₂ / Na₂O / ZnO / ZnS / MgF₂ / TiO₂ / diamond / SiC / Cu / Fe / Mg) |
| `test_element_data.py` | Consistency of constants data: unregistered elements, metals missing radii, ions missing electronegativity |
| `test_constants.py` | Architecture constraint (no `app.domain` / `app.api` imports) and consistency of wurtzite / rutile cell parameters |
| `test_serialization.py` | Response field names and display precision (all floats rounded to 3 decimals) |
| `test_predict_api.py` | `POST /predict` response contract: ionic / rutile / metallic results, plus translatable codes |
| `test_request_context.py` | Request IDs: reuse / generate / replace invalid values, and prediction logs carrying request_id |
| `test_api_errors.py` | `GET /health`, unsupported substances, and unified responses for unexpected exceptions |
| `test_env.py` | Environment variable reading: defaults, empty values, invalid integers |

## Coverage

```powershell
uv run pytest --cov=app --cov-report=term-missing
```

Currently about 98%. The main uncovered parts are the startup entry (`app/main.py` really binds the port),
defensive branches in `app/domain/prediction.py` that require deliberately bad data (missing element data / radius ratio too small),
and the "root logger already has a handler" branch in logging setup.

CI enforces `--cov-fail-under=90`: new code with no tests at all fails the pipeline.

> The allowed error for unit-cell edges is 0.01 Å. The hard-sphere model itself has a systematic deviation of a few percent,
> so assertions compare "same order of magnitude as the textbook and same structure type", not refined experimental values.
