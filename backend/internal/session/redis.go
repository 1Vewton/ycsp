package session

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// keyPrefix 与 Redis 里的其他键区分开，方便用 redis-cli 查看。
const keyPrefix = "ycsp:session:"

// RedisStore 把会话存在 Redis 里：多个后端实例共享，重启后端也不会掉登录。
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore 用一个已连通的 Redis 客户端创建会话存储。
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

func (store *RedisStore) Create(
	ctx context.Context, userID uint, lifetime time.Duration,
) (string, error) {
	sessionID, err := newSessionID()
	if err != nil {
		return "", err
	}

	// 存用户编号而不是整个用户对象：用户改名后会话仍然有效，也不用同步多份数据。
	if err := store.client.Set(ctx, keyPrefix+sessionID, userID, lifetime).Err(); err != nil {
		return "", fmt.Errorf("session: store in redis: %w", err)
	}
	return sessionID, nil
}

func (store *RedisStore) Lookup(ctx context.Context, sessionID string) (uint, error) {
	value, err := store.client.Get(ctx, keyPrefix+sessionID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, ErrSessionNotFound
		}
		return 0, fmt.Errorf("session: read from redis: %w", err)
	}

	userID, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("session: decode user id %q: %w", value, err)
	}
	return uint(userID), nil
}

func (store *RedisStore) Delete(ctx context.Context, sessionID string) error {
	if err := store.client.Del(ctx, keyPrefix+sessionID).Err(); err != nil {
		return fmt.Errorf("session: delete from redis: %w", err)
	}
	return nil
}
