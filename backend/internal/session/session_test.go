package session_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/1Vewton/ycsp/backend/internal/session"
)

// newRedisStore 用进程内的 miniredis 造一个真实的 Redis 客户端，CI 不需要起 Redis 服务。
func newRedisStore(t *testing.T) (*session.RedisStore, *miniredis.Miniredis) {
	t.Helper()

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	return session.NewRedisStore(client), server
}

func testStores(t *testing.T) map[string]session.Store {
	t.Helper()

	redisStore, _ := newRedisStore(t)
	return map[string]session.Store{
		"memory": session.NewMemoryStore(),
		"redis":  redisStore,
	}
}

func TestStoresRoundTripASession(t *testing.T) {
	for name, store := range testStores(t) {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			sessionID, err := store.Create(ctx, 42, time.Hour)
			if err != nil {
				t.Fatalf("Create: %v", err)
			}
			if sessionID == "" {
				t.Fatal("Create returned an empty session id")
			}

			userID, err := store.Lookup(ctx, sessionID)
			if err != nil {
				t.Fatalf("Lookup: %v", err)
			}
			if userID != 42 {
				t.Errorf("user id = %d, want 42", userID)
			}

			if err := store.Delete(ctx, sessionID); err != nil {
				t.Fatalf("Delete: %v", err)
			}
			if _, err := store.Lookup(ctx, sessionID); !errors.Is(err, session.ErrSessionNotFound) {
				t.Errorf("Lookup after Delete error = %v, want ErrSessionNotFound", err)
			}
		})
	}
}

func TestStoresReportUnknownSessions(t *testing.T) {
	for name, store := range testStores(t) {
		t.Run(name, func(t *testing.T) {
			if _, err := store.Lookup(context.Background(), "not-a-session"); !errors.Is(
				err, session.ErrSessionNotFound,
			) {
				t.Errorf("error = %v, want ErrSessionNotFound", err)
			}
		})
	}
}

func TestSessionIDsAreRandom(t *testing.T) {
	store := session.NewMemoryStore()

	firstID, err := store.Create(context.Background(), 1, time.Hour)
	if err != nil {
		t.Fatalf("first Create: %v", err)
	}
	secondID, err := store.Create(context.Background(), 1, time.Hour)
	if err != nil {
		t.Fatalf("second Create: %v", err)
	}

	if firstID == secondID {
		t.Fatal("two sessions share the same id: the id is not random")
	}
}

func TestRedisSessionsExpire(t *testing.T) {
	store, server := newRedisStore(t)
	ctx := context.Background()

	sessionID, err := store.Create(ctx, 7, time.Minute)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	server.FastForward(2 * time.Minute)

	if _, err := store.Lookup(ctx, sessionID); !errors.Is(err, session.ErrSessionNotFound) {
		t.Errorf("error = %v, want ErrSessionNotFound after expiry", err)
	}
}
