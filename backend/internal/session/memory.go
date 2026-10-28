package session

import (
	"context"
	"sync"
	"time"
)

// MemoryStore 把会话放在进程内存里：只适合单机开发——
// 重启后端就全部掉登录，多实例之间也不共享。
type MemoryStore struct {
	mutex    sync.Mutex
	sessions map[string]memorySession
	now      func() time.Time // 测试里替换成可控时钟
}

type memorySession struct {
	userID    uint
	expiresAt time.Time
}

// NewMemoryStore 创建内存会话存储。
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{sessions: make(map[string]memorySession), now: time.Now}
}

func (store *MemoryStore) Create(
	_ context.Context, userID uint, lifetime time.Duration,
) (string, error) {
	sessionID, err := newSessionID()
	if err != nil {
		return "", err
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.sessions[sessionID] = memorySession{
		userID:    userID,
		expiresAt: store.now().Add(lifetime),
	}
	return sessionID, nil
}

func (store *MemoryStore) Lookup(_ context.Context, sessionID string) (uint, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	found, exists := store.sessions[sessionID]
	if !exists {
		return 0, ErrSessionNotFound
	}
	if store.now().After(found.expiresAt) {
		delete(store.sessions, sessionID)
		return 0, ErrSessionNotFound
	}
	return found.userID, nil
}

func (store *MemoryStore) Delete(_ context.Context, sessionID string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	delete(store.sessions, sessionID)
	return nil
}
