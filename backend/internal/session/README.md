# 数据管理层 · 会话存储（backend/internal/session）

登录会话就是“随机编号 → 用户编号”的映射，浏览器只拿到编号（存在 Cookie 里）。

| 文件 | 作用 |
| --- | --- |
| `session.go` | `Store` 接口、`ErrSessionNotFound`、随机编号生成 |
| `redis.go` | `RedisStore`：会话存 Redis，多实例共享、重启不掉登录（默认） |
| `memory.go` | `MemoryStore`：存在进程内存，只适合单机开发 |

## 编号怎么来的

32 字节 `crypto/rand` → base64url（43 字符）。256 位随机，猜中概率可以忽略。
编号本身**不含**任何用户信息，也无法从编号推算出用户。

## Redis 里存什么

```text
ycsp:session:<随机编号>  →  用户编号（字符串），TTL = 会话有效期
```

只存用户编号而不是整个用户对象：用户改名后会话仍然有效，也不用同步多份数据。

## 两种实现怎么选

| 驱动 | 适用 | 代价 |
| --- | --- | --- |
| `redis`（默认） | 正式运行、多实例 | 需要起 Redis；启动时连不上会直接失败并提示 |
| `memory` | 本地没有 Redis 时临时开发 | 重启全部掉登录、多实例不共享 |

## 已知限制

- 会话有效期固定（不滑动续期）：默认 7 天，可在 `YCSP_SESSION_LIFETIME_SECONDS` 调整。
- 退出登录只删当前会话；同一账号在别的设备上的会话仍然有效。

测试：`go test ./internal/session/`（Redis 用进程内的 `miniredis`，CI 不需要起 Redis 服务）。
