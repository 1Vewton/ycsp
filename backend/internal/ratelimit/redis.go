package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCounter 把计数存在 Redis 里：多实例共享，重启也不清零。
type RedisCounter struct {
	client *redis.Client
}

// NewRedisCounter 用一个已连通的 Redis 客户端创建计数器。
func NewRedisCounter(client *redis.Client) *RedisCounter {
	return &RedisCounter{client: client}
}

func (counter *RedisCounter) Increment(
	ctx context.Context, key string, window time.Duration,
) (int, time.Duration, error) {
	rawCount, err := counter.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("ratelimit: incr %s: %w", key, err)
	}
	count := int(rawCount)

	if count == 1 {
		// 第一次计数时开窗口。
		if err := counter.client.Expire(ctx, key, window).Err(); err != nil {
			return count, 0, fmt.Errorf("ratelimit: expire %s: %w", key, err)
		}
	}

	remaining, err := counter.client.TTL(ctx, key).Result()
	if err != nil {
		return count, 0, fmt.Errorf("ratelimit: ttl %s: %w", key, err)
	}
	if remaining < 0 {
		// 兜底：万一上次进程在 INCR 与 EXPIRE 之间挂掉，别把账号永久锁死。
		if err := counter.client.Expire(ctx, key, window).Err(); err != nil {
			return count, 0, fmt.Errorf("ratelimit: repair ttl %s: %w", key, err)
		}
		remaining = window
	}
	return count, remaining, nil
}

func (counter *RedisCounter) Current(
	ctx context.Context, key string,
) (int, time.Duration, error) {
	rawCount, err := counter.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, 0, nil
		}
		return 0, 0, fmt.Errorf("ratelimit: get %s: %w", key, err)
	}

	count, err := strconv.Atoi(rawCount)
	if err != nil {
		return 0, 0, fmt.Errorf("ratelimit: decode count %q: %w", rawCount, err)
	}

	remaining, err := counter.client.TTL(ctx, key).Result()
	if err != nil {
		return count, 0, fmt.Errorf("ratelimit: ttl %s: %w", key, err)
	}
	if remaining < 0 {
		return count, 0, nil
	}
	return count, remaining, nil
}

func (counter *RedisCounter) Reset(ctx context.Context, key string) error {
	if err := counter.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("ratelimit: del %s: %w", key, err)
	}
	return nil
}

var _ Counter = (*RedisCounter)(nil)
