package ratelimit

import (
	"context"
	"sync"
	"time"
)

// MemoryCounter 把计数放在进程内存里，只适合单机开发（重启即清零、多实例不共享）。
type MemoryCounter struct {
	mutex   sync.Mutex
	entries map[string]memoryEntry
	now     func() time.Time // 测试里替换成可控时钟
}

type memoryEntry struct {
	count     int
	expiresAt time.Time
}

// NewMemoryCounter 创建内存计数器。
func NewMemoryCounter() *MemoryCounter {
	return &MemoryCounter{entries: make(map[string]memoryEntry), now: time.Now}
}

func (counter *MemoryCounter) Increment(
	_ context.Context, key string, window time.Duration,
) (int, time.Duration, error) {
	counter.mutex.Lock()
	defer counter.mutex.Unlock()

	currentTime := counter.now()
	entry, exists := counter.entries[key]
	if !exists || !currentTime.Before(entry.expiresAt) {
		// 新窗口：计数从 1 开始，并在 window 之后过期。
		entry = memoryEntry{count: 0, expiresAt: currentTime.Add(window)}
	}
	entry.count++
	counter.entries[key] = entry
	return entry.count, entry.expiresAt.Sub(currentTime), nil
}

func (counter *MemoryCounter) Current(
	_ context.Context, key string,
) (int, time.Duration, error) {
	counter.mutex.Lock()
	defer counter.mutex.Unlock()

	currentTime := counter.now()
	entry, exists := counter.entries[key]
	if !exists || !currentTime.Before(entry.expiresAt) {
		delete(counter.entries, key)
		return 0, 0, nil
	}
	return entry.count, entry.expiresAt.Sub(currentTime), nil
}

func (counter *MemoryCounter) Reset(_ context.Context, key string) error {
	counter.mutex.Lock()
	defer counter.mutex.Unlock()

	delete(counter.entries, key)
	return nil
}

var _ Counter = (*MemoryCounter)(nil)
