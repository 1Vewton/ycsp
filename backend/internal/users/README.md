# 数据管理层 · 账号（backend/internal/users）

注册与登录校验。密码哈希交给 [`../auth`](../auth/README.md)，会话交给 [`../session`](../session/README.md)。

| 文件 | 作用 |
| --- | --- |
| `service.go` | `Register`（注册）、`Authenticate`（登录校验）、`ByID` |

## 规则

- **昵称**：2~32 个字符（按 Unicode 字符数算），自动去掉首尾空格，**唯一**——它同时是登录名。
- **密码**：至少 8 位，且含英文字母、数字、特殊符号各一个（见 `auth.CheckPasswordPolicy`）。
  不合规时一次返回**所有**未满足的规则，学生改一遍就能通过。
- **不留线索**：昵称不存在与密码错误都返回 `ErrInvalidCredentials`；账号不存在时也会做一次
  同代价的 bcrypt 校验，避免用响应时间探测账号是否存在。

## 已知限制

- 昵称唯一性与登录都是**区分大小写**的（`Alice` 与 `alice` 是两个账号）。
- 还没有“改密码 / 重置密码”的接口；老库里 `password_hash` 为空的账号无法登录，需要等重置功能补齐。

测试：`go test ./internal/users/`（用 `testutil` 的临时 SQLite）。
