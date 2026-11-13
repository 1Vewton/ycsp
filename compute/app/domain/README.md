> English | [简体中文](./README.zh-CN.md)

# Data processing layer · Domain models (compute/app/domain)

Crystallography data types and prediction rules; element data, structure prototype data, thresholds, and status codes live in [../constants/README.md](../constants/README.md).

| File | Purpose |
| --- | --- |
| `structure.py` | Element / atom / unit cell (lengths in Å, angles in °), including picometer-to-ångström conversion |
| `elements.py` | Builds `Element` objects and provides table accessors (`find_*` return None; `get_*` raise with a hint about where to add data) |
| `composition.py` | Formula parsing (`CaF2`, `Ca(OH)2` → element → count) |
| `prototypes.py` | Computes unit-cell edges from the hard-sphere contact model and places atoms (NaCl / CsCl / ZnS / wurtzite / fluorite / antifluorite / rutile / diamond / fcc / bcc / hcp) |
| `prediction.py` | Classifies the type + selects the prototype + explains the rationale and caveats; only `predict_structure` is public |
| `notice.py` | `Notice` (code + params) and `DomainError`: language-neutral explanations and errors |

The "why" of each criterion is documented in comments in `prediction.py`, and data sources are at the top of each file in `constants/`;
this layer produces no student-facing text — the frontend translates every code.

