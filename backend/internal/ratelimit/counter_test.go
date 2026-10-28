package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/1Vewton/ycsp/backend/internal/ratelimit"
)

// 两种计数器共用同一组断言，保证内存模式与 Redis 模式行为一致。
func testCounters(t *testing.T) map[string]ratelimit.Counter {
	t.Helper()

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	return map[string]ratelimit.Counter{
		"memory": ratelimit.NewMemoryCounter(),
		"redis":  ratelimit.NewRedisCounter(client),
	}
}

func TestCountersCountWithinAWindow(t *testing.T) {
	for name, counter := range testCounters(t) {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			for attempt := 1; attempt <= 3; attempt++ {
				count, remaining, err := counter.Increment(ctx, "key", time.Minute)
				if err != nil {
					t.Fatalf("Increment: %v", err)
				}
				if count != attempt {
					t.Errorf("count = %d, want %d", count, attempt)
				}
				if remaining <= 0 {
					t.Errorf("remaining = %v, want a positive window", remaining)
				}
			}

			count, _, err := counter.Current(ctx, "key")
			if err != nil {
				t.Fatalf("Current: %v", err)
			}
			if count != 3 {
				t.Errorf("current count = %d, want 3", count)
			}
		})
	}
}

func TestCountersReportUnknownKeys(t *testing.T) {
	for name, counter := range testCounters(t) {
		t.Run(name, func(t *testing.T) {
			count, remaining, err := counter.Current(context.Background(), "missing")
			if err != nil {
				t.Fatalf("Current: %v", err)
			}
			if count != 0 || remaining != 0 {
				t.Errorf("count = %d remaining = %v, want zeros", count, remaining)
			}
		})
	}
}

func TestCountersReset(t *testing.T) {
	for name, counter := range testCounters(t) {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			if _, _, err := counter.Increment(ctx, "key", time.Minute); err != nil {
				t.Fatalf("Increment: %v", err)
			}
			if err := counter.Reset(ctx, "key"); err != nil {
				t.Fatalf("Reset: %v", err)
			}

			count, _, err := counter.Current(ctx, "key")
			if err != nil {
				t.Fatalf("Current: %v", err)
			}
			if count != 0 {
				t.Errorf("count after Reset = %d, want 0", count)
			}
		})
	}
}

func TestRedisCounterStartsANewWindowAfterExpiry(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	counter := ratelimit.NewRedisCounter(client)
	ctx := context.Background()

	if _, _, err := counter.Increment(ctx, "key", time.Minute); err != nil {
		t.Fatalf("Increment: %v", err)
	}
	server.FastForward(2 * time.Minute)

	count, _, err := counter.Current(ctx, "key")
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if count != 0 {
		t.Errorf("count after expiry = %d, want 0", count)
	}

	count, _, err = counter.Increment(ctx, "key", time.Minute)
	if err != nil {
		t.Fatalf("Increment after expiry: %v", err)
	}
	if count != 1 {
		t.Errorf("count in the new window = %d, want 1", count)
	}
}
