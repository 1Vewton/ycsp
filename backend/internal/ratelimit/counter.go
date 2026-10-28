// Package ratelimit 限制登录失败次数，挡住暴力猜密码。
//
// 计数存 Redis（与登录会话同一个后端），内存模式下改存进程内存，便于本地开发。
package ratelimit

import (
	"context"
	"time"
)

// Counter 记录某个键在时间窗口内的次数。
type Counter interface {
	// Increment 增加一次计数，返回增加后的计数与剩余窗口时间。
	Increment(ctx context.Context, key string, window time.Duration) (int, time.Duration, error)
	// Current 读取当前计数与剩余窗口时间；键不存在时返回 0。
	Current(ctx context.Context, key string) (int, time.Duration, error)
	// Reset 清除计数（登录成功后调用）。
	Reset(ctx context.Context, key string) error
}

// Policy 是登录限流的阈值：同一昵称、同一地址各自最多失败几次，以及在多长的窗口内统计。
type Policy struct {
	MaxAttemptsPerAccount int
	MaxAttemptsPerAddress int
	Window                time.Duration
}
