> English | [简体中文](./README.zh-CN.md)

# Data management layer · Status codes (backend/internal/codes)

Every code the data management layer returns to the frontend. **Only codes and params, no language-specific copy**:
the frontend translates the copy into the current language (see the root README "Languages").

| Group | Examples | Meaning |
| --- | --- | --- |
| Success | `backend_healthy`, `task_created`, `task_result`, `task_history` | The request succeeded |
| Account security | `password_changed`, `password_reset`, `security_question_updated`, `current_password_incorrect`, `password_reset_failed` | Password change / reset and security question results |
| Forwarding failure | `compute_unreachable`, `compute_timeout`, `compute_bad_status` | The data processing layer did not respond, did not answer in time, or returned an invalid response |
| Task errors | `task_not_found`, `task_not_ready`, `task_cancel_conflict` | Task ID not found / result not ready / state does not allow cancellation |
| Input errors | `task_formula_empty`, `task_formula_too_long` | The formula is empty or too long |
| General | `request_invalid_body`, `internal_error` | Invalid request body, unexpected exception |

Conventions:

- A new code must also add Chinese and English copy in the frontend language packs `frontend/src/locales/` (see that directory's README).
- Put numbers that belong in a sentence into params (e.g. `{"max_length": 64}`); never hard-code numbers in copy.
- Codes returned by the data processing layer itself (e.g. `formula_charge_unbalanced`) pass through unchanged; the frontend already has copy for them.
