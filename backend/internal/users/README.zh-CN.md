> [English](./README.md) | 简体中文

# 数据管理层 · 账号（backend/internal/users）

注册与登录校验。密码哈希交给 [`../auth`](../auth/README.zh-CN.md)，会话交给 [`../session`](../session/README.zh-CN.md)。

| 文件 | 作用 |
| --- | --- |
| `service.go` | `Register` / `RegisterWithSecurity`、`Authenticate`（登录校验）、`ByID`、`ChangePassword`、`SetSecurityQuestion`、`QuestionFor`、`ResetPassword` |

## 规则

- **昵称**：2~32 个字符（按 Unicode 字符数算），自动去掉首尾空格，**唯一**——它同时是登录名。
- **密码**：至少 8 位，且含英文字母、数字、特殊符号各一个（见 `auth.CheckPasswordPolicy`）。
  不合规时一次返回**所有**未满足的规则，学生改一遍就能通过。
- **不留线索**：昵称不存在与密码错误都返回 `ErrInvalidCredentials`；账号不存在时也会做一次
  同代价的 bcrypt 校验，避免用响应时间探测账号是否存在。
- **安全问题**：注册时可选（问题 2~80 个字符，答案至少 2 个字符且不超过 72 字节）。
  答案会去首尾空格、转小写，只存 bcrypt 哈希。登录后可以设置或更新，但必须输入当前密码。
- **修改密码**：需要当前密码，沿用注册时的密码策略，并拒绝与旧密码相同的新密码。
- **重置密码**：先校验安全问题答案，再校验新密码策略。账号不存在、没有设置安全问题、
  答案不对都返回 `ErrPasswordResetFailed`，避免泄露账号是否存在。

## 已知限制

- 昵称唯一性与登录都是**区分大小写**的（`Alice` 与 `alice` 是两个账号）。
- 修改或重置密码不会让其他设备上的会话失效；目前还没有“退出所有设备”。
- 老库里 `password_hash` 为空的账号无法登录，也无法设置安全问题，需要管理员手工修复。

测试：`go test ./internal/users/`（用 `testutil` 的临时 SQLite）。
