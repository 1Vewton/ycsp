// Package session 管理登录会话：会话编号 → 用户。
//
// 会话存在 Redis（RedisStore）或进程内存（MemoryStore，仅单机开发）里，
// 浏览器只拿到一个随机编号的 Cookie，编号本身不含任何用户信息。
package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

// 支持的会话存储方式（YCSP_SESSION_DRIVER 的取值）。
const (
	DriverRedis  = "redis"
	DriverMemory = "memory"
)

// ErrSessionNotFound 表示会话不存在或已过期。
var ErrSessionNotFound = errors.New("session: not found")

// Store 保存会话；实现见 redis.go 与 memory.go。
type Store interface {
	Create(ctx context.Context, userID uint, lifetime time.Duration) (string, error)
	Lookup(ctx context.Context, sessionID string) (uint, error)
	Delete(ctx context.Context, sessionID string) error
}

// sessionIDBytes 是会话编号的随机字节数：32 字节 = 256 位，猜中的概率可以忽略。
const sessionIDBytes = 32

func newSessionID() (string, error) {
	buffer := make([]byte, sessionIDBytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("session: generate id: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
