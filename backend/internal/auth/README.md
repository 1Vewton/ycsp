# 数据管理层 · 账号安全（backend/internal/auth）

密码只存**加盐哈希**，从不存明文、也不记录明文日志。

| 文件 | 作用 |
| --- | --- |
| `password.go` | `HashPassword`（bcrypt，每次新盐）与 `VerifyPassword`（常量时间比对由 bcrypt 负责） |
| `policy.go` | `CheckPasswordPolicy`：检查长度与字符种类，返回全部未满足的规则 code |

## 为什么用 bcrypt

- **盐是随机的、且写在哈希串里**：同一个密码两次哈希结果不同，彩虹表失效；
  数据库只要一列 `password_hash`，不用额外的 salt 列——盐和哈希不会配错。
- **故意慢**：bcrypt 的代价因子让单次校验约几十毫秒，登录感觉不到，离线爆破却贵得多。
- **上限 72 字节**：bcrypt 会忽略超长部分，所以我们直接报错，而不是静默截断。

## 密码策略（`CheckPasswordPolicy`）

至少 8 个字符，且同时包含：

| 规则 code | 含义 |
| --- | --- |
| `password_rule_too_short` | 少于 8 个字符 |
| `password_rule_too_long` | 超过 bcrypt 的 72 字节 |
| `password_rule_missing_letter` | 没有英文字母（只认 ASCII，中文不算） |
| `password_rule_missing_digit` | 没有数字 |
| `password_rule_missing_symbol` | 没有特殊符号（标点或符号，空格不算） |

一次返回**所有**未满足的规则，前端逐条翻译展示；改一次就能全部通过。

## 存的形状

```text
$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy
 │  │  └── 22 字符盐 + 31 字符哈希
 │  └── 代价因子 10
 └── bcrypt 版本
```

## 空哈希的含义

历史上没有密码的账号，`password_hash` 是空串：`VerifyPassword` 对它一律返回 false，
也就是**这种账号不能登录**，要等重置密码后再写回哈希——不会出现“空密码就能进”的漏洞。

测试：`go test ./internal/auth/`
