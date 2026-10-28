package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCounterDropsExpiredEntries(t *testing.T) {
	counter := NewMemoryCounter()

	// 换成可控时钟，避免测试真的等一分钟。
	currentTime := time.Now()
	counter.now = func() time.Time { return currentTime }

	if _, _, err := counter.Increment(context.Background(), "key", time.Minute); err != nil {
		t.Fatalf("Increment: %v", err)
	}

	currentTime = currentTime.Add(2 * time.Minute)

	count, remaining, err := counter.Current(context.Background(), "key")
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if count != 0 || remaining != 0 {
		t.Errorf("count = %d remaining = %v, want zeros after expiry", count, remaining)
	}
}
