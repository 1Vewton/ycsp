> English | [简体中文](./README.zh-CN.md)

# Data processing layer (compute)

Predicts crystal structures from chemical composition and exposes them over HTTP for the data management layer to call (Python + FastAPI).

## Public API

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/health` | Health check |
| POST | `/predict` | Predict a crystal structure from a chemical formula |

Implementation: [app/api/README.md](./app/api/README.md). Responses use the envelope `{ service, status, code, params? }`:
success also carries `prediction`; errors carry only language-neutral `code` and `params`, and the frontend translates the copy (see the root README "Languages").

### Prediction request and response

The request needs only a formula, for example `NaCl`, `CaF2`, `C`, `Cu` (bracketed forms such as `Ca(OH)2` are also accepted):

```text
POST /predict
{ "formula": "NaCl" }
```

Response (excerpt):

```json
{
  "service": "ycsp-compute",
  "status": "ok",
  "code": "prediction_ready",
  "prediction": {
    "formula": "NaCl",
    "family": "ionic",
    "candidates": [{
      "prototype": "rock_salt",
      "coordination_number": 6,
      "rule": { "code": "rule_radius_ratio", "params": { "ratio": 0.564, "coordination_number": 6 } },
      "caveats": [{ "code": "caveat_confirmed_by_experiment", "params": {} }],
      "cell": { "edge_a_angstrom": 5.66, "angle_alpha_degree": 90.0 },
      "atoms": [{ "symbol": "Na", "fractional_coordinates": [0.0, 0.0, 0.0] }]
    }]
  }
}
```

- `cell` holds the unit-cell parameters (edge lengths in Å, angles in °); `atoms` holds atoms and fractional coordinates, ready for the frontend to draw.
- `rule` explains "why this structure", and `caveats` lists the limits of the result; both contain only codes and numbers, and the frontend turns each code into a sentence (copy lives in `frontend/src/locales/`).

## Log tracing

Each request logs one line with an ID: `request_id=… method=… path=… status=… duration_ms=…`;
a successful `/predict` also logs `formula`, `family`, and `prototype`, and exception stack traces carry the same ID.

- The ID reuses the data management layer's `X-Request-ID` when present; when missing or invalid, a `compute-…` ID is generated and written back to the response header.
- Middleware lives in `app/api/request_context.py`; the log format and the `request_id` filter live in `app/logging_setup.py`.
- Only method, path, status, duration, and prediction summary are logged; sensitive information such as cookies is never logged.

## Algorithm

It first classifies the crystal type from the composition, then applies textbook rules (code in `app/domain/prediction.py`):

| Type | Criterion | Structure type |
| --- | --- | --- |
| Ionic crystal | One metal cation + one anion, with charge balance | 1:1 uses the Pauling radius ratio (ZnO / ZnS use the measured wurtzite cell); 1:2 / 2:1 use fluorite / antifluorite (MgF2 / TiO2 use the measured rutile cell) |
| Covalent crystal | C / Si / Ge element; two nonmetals in 1:1 (e.g. SiC) | Diamond type; SiC uses zinc blende |
| Metallic crystal | Only one metal element | Face-centered cubic / body-centered cubic / hexagonal close-packed (from an experimental table) |

Textbook compounds with measured cells use experimental data: ZnO / ZnS are wurtzite, MgF2 / TiO2 are rutile, and the 1:1 compounds in `KNOWN_IONIC_STRUCTURES` override the radius-ratio rule.

Radius-ratio boundaries (dimensionless): below 0.225 needs 3-coordination; 0.225–0.414 is 4-coordination; 0.414–0.732 is 6-coordination; above 0.732 is 8-coordination. Unit-cell edges come from the hard-sphere contact model, e.g. NaCl touches along the edge with a = 2(r₊+r₋) = 5.66 Å (textbook: 5.64 Å).

Element data, structure prototype data, and these thresholds all live in `app/constants/`; `app/domain/` contains only types and rules.

This stage uses only the standard library `math`; NumPy will be introduced when batch coordinates or matrix operations are actually needed, not just for the sake of using it.

## Known limitations

- **The radius ratio is only an empirical rule**: when the ratio is close to a boundary, either coordination number may occur; such results are marked with `caveat_near_boundary`.
  KCl has a ratio of about 0.76, so the rule predicts 8-coordination, but the measured structure is NaCl type. Classic textbook compounds therefore use experimental data (`KNOWN_IONIC_STRUCTURES` in `app/constants/structures.py`) and explain the difference with `rule_experimental_override`, `rule_wurtzite_experimental`, or `rule_rutile_experimental`.
- **Metal packing cannot be derived from composition alone** (the same metal may have different structures at different temperatures); structure types come from experimental data, and the UI must say so (`caveat_metal_structure_experimental`).
- **Wurtzite is covered for ZnO / ZnS with measured cells**: both use the hexagonal wurtzite type. ZnS also has a cubic zinc blende form, so its result carries `caveat_polymorph`.
- **1:2 / 2:1 structure types are chosen by ion count ratio** (fluorite / antifluorite, e.g. CaF₂, Na₂O); MgF₂ and TiO₂ are measured rutile types and use their measured cells. Other 1:2 pairs still use fluorite, and when the radius ratio is outside the 8-coordination band the result carries `caveat_stoichiometry_outside_band` and tells students to trust experiment.
- Only the three crystal types above are covered; multi-element formulas (e.g. CaCO₃), 2:3 stoichiometry (e.g. Al₂O₃), and 3-coordination structures return
  `composition_not_supported` / `radius_ratio_too_small` instead of guessing a result.
- The computation is pure arithmetic and finishes in milliseconds, so no progress bar or cancel control is needed.

## Directory

| Path | Purpose |
| --- | --- |
| `app/` | Service code, see [app/README.md](./app/README.md) |
| `tests/` | Unit tests (classic textbook examples), see [tests/README.md](./tests/README.md) |

## How to run

Dependencies are managed with [uv](https://docs.astral.sh/uv/) (versions are locked in `uv.lock`; the first run creates `.venv` automatically).

```powershell
uv run python -m app.main   # service: default 127.0.0.1:8001
uv run pytest               # unit tests
uv run pytest --cov=app --cov-report=term-missing   # unit tests + coverage
```

Once the service is running: health check <http://127.0.0.1:8001/health>, API docs <http://127.0.0.1:8001/docs>.
To change the listen address or port, copy `.env.example` to `.env` and edit it.
