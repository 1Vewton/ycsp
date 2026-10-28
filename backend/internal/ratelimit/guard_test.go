package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/1Vewton/ycsp/backend/internal/ratelimit"
)

func newGuard(t *testing.T, policy ratelimit.Policy) *ratelimit.Guard {
	t.Helper()

	guard, err := ratelimit.NewGuard(ratelimit.NewMemoryCounter(), policy)
	if err != nil {
		t.Fatalf("NewGuard: %v", err)
	}
	return guard
}

func TestNewGuardRejectsInvalidPolicies(t *testing.T) {
	counter := ratelimit.NewMemoryCounter()
	validWindow := time.Minute

	invalidPolicies := map[string]ratelimit.Policy{
		"zero account limit": {MaxAttemptsPerAccount: 0, MaxAttemptsPerAddress: 1, Window: validWindow},
		"zero address limit": {MaxAttemptsPerAccount: 1, MaxAttemptsPerAddress: 0, Window: validWindow},
		"zero window":        {MaxAttemptsPerAccount: 1, MaxAttemptsPerAddress: 1, Window: 0},
	}
	for description, policy := range invalidPolicies {
		if _, err := ratelimit.NewGuard(counter, policy); err == nil {
			t.Errorf("%s: expected an error", description)
		}
	}

	if _, err := ratelimit.NewGuard(nil, ratelimit.Policy{
		MaxAttemptsPerAccount: 1, MaxAttemptsPerAddress: 1, Window: validWindow,
	}); err == nil {
		t.Error("nil counter: expected an error")
	}
}

func TestGuardBlocksOnlyAfterTheAccountLimit(t *testing.T) {
	guard := newGuard(t, ratelimit.Policy{
		MaxAttemptsPerAccount: 3,
		MaxAttemptsPerAddress: 100,
		Window:                time.Minute,
	})
	ctx := context.Background()

	if retryAfter, err := guard.Check(ctx, "小明", "10.0.0.1"); err != nil || retryAfter != 0 {
		t.Fatalf("fresh account: retryAfter = %v err = %v", retryAfter, err)
	}

	for attempt := 1; attempt <= 2; attempt++ {
		retryAfter, err := guard.RecordFailure(ctx, "小明", "10.0.0.1")
		if err != nil {
			t.Fatalf("RecordFailure: %v", err)
		}
		if retryAfter != 0 {
			t.Errorf("attempt %d blocked too early", attempt)
		}
	}

	retryAfter, err := guard.RecordFailure(ctx, "小明", "10.0.0.1")
	if err != nil {
		t.Fatalf("RecordFailure: %v", err)
	}
	if retryAfter <= 0 || retryAfter > time.Minute {
		t.Fatalf("third failure retryAfter = %v, want (0, 1m]", retryAfter)
	}

	blocked, err := guard.Check(ctx, "小明", "10.0.0.1")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if blocked <= 0 {
		t.Error("Check should report the account as blocked")
	}
}

func TestGuardResetAccountKeepsTheAddressCounter(t *testing.T) {
	guard := newGuard(t, ratelimit.Policy{
		MaxAttemptsPerAccount: 3,
		MaxAttemptsPerAddress: 3,
		Window:                time.Minute,
	})
	ctx := context.Background()

	guard.RecordFailure(ctx, "小明", "10.0.0.1")
	guard.RecordFailure(ctx, "小明", "10.0.0.1")

	if err := guard.ResetAccount(ctx, "小明"); err != nil {
		t.Fatalf("ResetAccount: %v", err)
	}
	if retryAfter, err := guard.Check(ctx, "小明", "10.0.0.1"); err != nil || retryAfter != 0 {
		t.Fatalf("after reset: retryAfter = %v err = %v", retryAfter, err)
	}

	// 地址计数仍是 2：再失败一次就达到地址上限 3。
	retryAfter, err := guard.RecordFailure(ctx, "小明", "10.0.0.1")
	if err != nil {
		t.Fatalf("RecordFailure: %v", err)
	}
	if retryAfter <= 0 {
		t.Error("the address counter should still block after a successful reset")
	}
}

func TestGuardLimitsPerAddressAcrossAccounts(t *testing.T) {
	guard := newGuard(t, ratelimit.Policy{
		MaxAttemptsPerAccount: 5,
		MaxAttemptsPerAddress: 2,
		Window:                time.Minute,
	})
	ctx := context.Background()

	if _, err := guard.RecordFailure(ctx, "甲", "10.0.0.1"); err != nil {
		t.Fatalf("first RecordFailure: %v", err)
	}
	retryAfter, err := guard.RecordFailure(ctx, "乙", "10.0.0.1")
	if err != nil {
		t.Fatalf("second RecordFailure: %v", err)
	}
	if retryAfter <= 0 {
		t.Error("the second failure from one address should block the address")
	}

	// 别的地址不受影响。
	if blocked, err := guard.Check(ctx, "丙", "10.0.0.2"); err != nil || blocked != 0 {
		t.Errorf("other address: retryAfter = %v err = %v", blocked, err)
	}
}

func TestGuardTreatsAccountNamesCaseInsensitively(t *testing.T) {
	guard := newGuard(t, ratelimit.Policy{
		MaxAttemptsPerAccount: 1,
		MaxAttemptsPerAddress: 100,
		Window:                time.Minute,
	})
	ctx := context.Background()

	if _, err := guard.RecordFailure(ctx, "Alice", "10.0.0.1"); err != nil {
		t.Fatalf("RecordFailure: %v", err)
	}

	blocked, err := guard.Check(ctx, "alice", "10.0.0.2")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if blocked <= 0 {
		t.Error("case variants of one name should share the counter")
	}
}
