package session

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryStoreDropsExpiredSessions(t *testing.T) {
	store := NewMemoryStore()

	// 把时钟换成可控的，避免测试真的等一分钟。
	currentTime := time.Now()
	store.now = func() time.Time { return currentTime }

	sessionID, err := store.Create(context.Background(), 9, time.Minute)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	currentTime = currentTime.Add(2 * time.Minute)

	if _, err := store.Lookup(context.Background(), sessionID); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("error = %v, want ErrSessionNotFound after expiry", err)
	}
}
