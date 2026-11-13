> English | [简体中文](./README.zh-CN.md)

# Data management layer · Account security (backend/internal/auth)

Passwords are stored only as **salted hashes**; plaintext is never stored or logged.

| File | Purpose |
| --- | --- |
| `password.go` | `HashPassword` (bcrypt, new salt each time) and `VerifyPassword` (bcrypt handles constant-time comparison) |
| `policy.go` | `CheckPasswordPolicy`: checks length and character classes, returning every unmet rule code |

## Why bcrypt

- **The salt is random and embedded in the hash string**: hashing the same password twice yields different results, defeating rainbow tables;
  the database needs only one `password_hash` column and no separate salt column, so salt and hash cannot get out of sync.
- **Deliberately slow**: bcrypt's cost factor makes a single check take tens of milliseconds, imperceptible at login but much more expensive for offline cracking.
- **72-byte limit**: bcrypt silently ignores anything beyond 72 bytes, so we return an error instead of truncating silently.

## Password policy (`CheckPasswordPolicy`)

At least 8 characters, containing all of:

| Rule code | Meaning |
| --- | --- |
| `password_rule_too_short` | fewer than 8 characters |
| `password_rule_too_long` | more than bcrypt's 72 bytes |
| `password_rule_missing_letter` | no English letter (ASCII only; Chinese characters do not count) |
| `password_rule_missing_digit` | no digit |
| `password_rule_missing_symbol` | no symbol (punctuation or symbol; spaces do not count) |

Returns **all** unmet rules at once; the frontend translates and shows each one, so a single edit can satisfy them all.

## Stored shape

```text
$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy
 │  │  └── 22-character salt + 31-character hash
 │  └── cost factor 10
 └── bcrypt version
```

## Meaning of an empty hash

Accounts created before passwords existed have an empty `password_hash`: `VerifyPassword` always returns false for it,
so **such an account cannot sign in** until a password reset writes a hash back. There is no "empty password opens the door" hole.

## Security answers

Answers to security questions use the same bcrypt hashing: trimmed, lowercased, then hashed into `users.security_answer_hash`.
The plaintext answer is never stored. Because bcrypt only looks at the first 72 bytes, the account layer rejects longer answers.

Tests: `go test ./internal/auth/`
