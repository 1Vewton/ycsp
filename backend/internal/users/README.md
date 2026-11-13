> English | [简体中文](./README.zh-CN.md)

# Data management layer · Accounts (backend/internal/users)

Registration and login validation. Password hashing is delegated to [`../auth`](../auth/README.md), and sessions to [`../session`](../session/README.md).

| File | Purpose |
| --- | --- |
| `service.go` | `Register` / `RegisterWithSecurity`, `Authenticate`, `ByID`, `ChangePassword`, `SetSecurityQuestion`, `QuestionFor`, `ResetPassword` |

## Rules

- **Display name**: 2–32 characters (counted in Unicode code points), leading/trailing spaces removed, and **unique** — it is also the login name.
- **Password**: at least 8 characters with one English letter, one digit, and one symbol (see `auth.CheckPasswordPolicy`). When invalid, **all** unmet rules are returned at once, so one edit can satisfy them all.
- **No hints**: an unknown display name and a wrong password both return `ErrInvalidCredentials`; when the account does not exist, a bcrypt check with the same cost is still performed, so response time cannot be used to probe whether an account exists.
- **Security question**: optional at registration (question 2–80 characters, answer at least 2 characters and at most 72 bytes). The answer is trimmed, lowercased and stored only as a bcrypt hash. Signed-in students can set or update it, but must enter the current password.
- **Change password**: needs the current password, applies the same policy as registration, and rejects a new password identical to the old one.
- **Reset password**: verifies the security answer, then applies the password policy. Unknown accounts, accounts without a security question and wrong answers all return `ErrPasswordResetFailed` to avoid leaking account existence.

## Known limitations

- Display-name uniqueness and login are **case-sensitive** (`Alice` and `alice` are two accounts).
- Changing or resetting a password does not invalidate sessions on other devices; there is no "sign out everywhere" yet.
- Accounts with an empty `password_hash` in an old database cannot sign in; they also cannot set a security question, so an administrator must repair them manually.

Tests: `go test ./internal/users/` (uses `testutil`'s temporary SQLite).
