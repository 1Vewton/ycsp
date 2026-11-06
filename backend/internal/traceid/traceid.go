// Package traceid 生成与传递请求编号。
//
// 一次预测在前端、数据管理层、数据处理层使用同一个 X-Request-ID，
// 出问题时按编号就能把三层的日志串起来。
package traceid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const (
	// HeaderName 是三层统一的请求编号头。
	HeaderName = "X-Request-ID"
	// MaxLength 限制编号长度：过长的头会撑大日志与响应头。
	MaxLength = 64
)

// 请求编号放在 context 里，异步任务会把创建时的编号继续传给计算层。
type contextKey struct{}

// New 生成一个随机编号；prefix 用来说明是哪一层最先创建（web / api / compute）。
func New(prefix string) string {
	var randomBytes [16]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		// crypto/rand 失败极罕见；退化成时间戳也比没有编号好。
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(randomBytes[:])
}

// Normalize 校验外部传入的编号；不合法时返回空字符串，由调用方重新生成。
//
// 只允许字母、数字与 -_.:：换行等字符会污染日志（日志注入），必须拒绝。
func Normalize(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" || len(value) > MaxLength {
		return ""
	}
	for _, character := range value {
		if !isSafe(character) {
			return ""
		}
	}
	return value
}

func isSafe(character rune) bool {
	switch {
	case character >= 'a' && character <= 'z':
		return true
	case character >= 'A' && character <= 'Z':
		return true
	case character >= '0' && character <= '9':
		return true
	case character == '-', character == '_', character == '.', character == ':':
		return true
	default:
		return false
	}
}

// With 把编号放进 context，供同一次请求里的下游调用读取。
func With(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, contextKey{}, requestID)
}

// From 读取 context 里的编号；没有时返回空字符串。
func From(ctx context.Context) string {
	requestID, _ := ctx.Value(contextKey{}).(string)
	return requestID
}
