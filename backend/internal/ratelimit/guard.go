package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// 键前缀：账号按登录名（统一小写，避免大小写变体绕过），地址按 IP。
const (
	accountKeyPrefix = "ycsp:login:account:"
	addressKeyPrefix = "ycsp:login:address:"
)

// Guard 按 Policy 判断是否应拒绝登录、记录失败、成功后清零。
type Guard struct {
	counter Counter
	policy  Policy
}

// NewGuard 创建登录限流器；阈值不合法时直接报错，避免“配错了等于没限流”。
func NewGuard(counter Counter, policy Policy) (*Guard, error) {
	if counter == nil {
		return nil, errors.New("ratelimit: counter is nil")
	}
	if policy.MaxAttemptsPerAccount < 1 || policy.MaxAttemptsPerAddress < 1 {
		return nil, fmt.Errorf(
			"ratelimit: limits must be at least 1, got account=%d address=%d",
			policy.MaxAttemptsPerAccount, policy.MaxAttemptsPerAddress,
		)
	}
	if policy.Window <= 0 {
		return nil, fmt.Errorf("ratelimit: window must be positive, got %v", policy.Window)
	}
	return &Guard{counter: counter, policy: policy}, nil
}

// Check 判断这次登录尝试是否已被限制；返回大于 0 表示还要等多久。
func (guard *Guard) Check(ctx context.Context, account string, address string) (time.Duration, error) {
	accountWait, err := guard.wait(ctx, accountKey(account), guard.policy.MaxAttemptsPerAccount)
	if err != nil {
		return 0, err
	}
	addressWait, err := guard.wait(ctx, addressKey(address), guard.policy.MaxAttemptsPerAddress)
	if err != nil {
		return 0, err
	}
	return longer(accountWait, addressWait), nil
}

// RecordFailure 记一次失败；返回大于 0 表示这次失败已经触发限制。
func (guard *Guard) RecordFailure(
	ctx context.Context, account string, address string,
) (time.Duration, error) {
	accountWait, err := guard.record(ctx, accountKey(account), guard.policy.MaxAttemptsPerAccount)
	if err != nil {
		return 0, err
	}
	addressWait, err := guard.record(ctx, addressKey(address), guard.policy.MaxAttemptsPerAddress)
	if err != nil {
		return 0, err
	}
	return longer(accountWait, addressWait), nil
}

// ResetAccount 在登录成功后清掉该昵称的失败计数。
// 地址计数刻意保留：否则“攻破一个账号”就能把整台机器的限流清零。
func (guard *Guard) ResetAccount(ctx context.Context, account string) error {
	return guard.counter.Reset(ctx, accountKey(account))
}

func (guard *Guard) wait(ctx context.Context, key string, limit int) (time.Duration, error) {
	count, remaining, err := guard.counter.Current(ctx, key)
	if err != nil {
		return 0, err
	}
	if count >= limit {
		return remaining, nil
	}
	return 0, nil
}

func (guard *Guard) record(ctx context.Context, key string, limit int) (time.Duration, error) {
	count, remaining, err := guard.counter.Increment(ctx, key, guard.policy.Window)
	if err != nil {
		return 0, err
	}
	if count >= limit {
		return remaining, nil
	}
	return 0, nil
}

func accountKey(account string) string {
	return accountKeyPrefix + strings.ToLower(strings.TrimSpace(account))
}

func addressKey(address string) string {
	return addressKeyPrefix + address
}

func longer(first, second time.Duration) time.Duration {
	if second > first {
		return second
	}
	return first
}
