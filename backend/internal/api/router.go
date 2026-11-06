package api

import (
	"net/netip"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/1Vewton/ycsp/backend/internal/ratelimit"
	"github.com/1Vewton/ycsp/backend/internal/session"
	"github.com/1Vewton/ycsp/backend/internal/tasks"
	"github.com/1Vewton/ycsp/backend/internal/users"
)

// Options 是路由需要的依赖。
type Options struct {
	// ComputeBaseURL 是数据处理层地址，用于 /api/ping-compute 与任务执行。
	ComputeBaseURL string
	// Tasks 为 nil 时不注册任务接口（只做健康检查与转发时用得上）。
	Tasks *tasks.Service
	// Auth 为空（Users 或 Sessions 为 nil）时不注册账号接口。
	Auth AuthOptions
}

// AuthOptions 是账号体系需要的依赖。
type AuthOptions struct {
	Users           *users.Service
	Sessions        session.Store
	SessionLifetime time.Duration
	SecureCookies   bool
	// LoginGuard 为 nil 时不做登录限流（只在校验或测试里这样用）。
	LoginGuard *ratelimit.Guard
	// TrustedProxies 是允许提供 X-Forwarded-For 的反向代理地址 / CIDR。
	// 为空表示不信任任何代理，只按直连地址限流（安全默认）。
	TrustedProxies []netip.Prefix
}

// NewRouter 注册数据管理层的全部路由。
func NewRouter(options Options) *gin.Engine {
	router := gin.New()
	// 先确定编号，再记录请求：日志与 panic 兜底里都能拿到同一个 request_id。
	// 客户端地址与登录限流共用同一套可信代理规则，日志里看到的 IP 就是限流用的 IP。
	addresses := newClientAddressResolver(options.Auth.TrustedProxies)
	router.Use(requestIDMiddleware(), requestLogger(addresses), recoveryMiddleware())

	router.GET("/api/health", handleBackendHealth)

	computeProxy := newComputeProxy(options.ComputeBaseURL)
	router.GET("/api/ping-compute", computeProxy.handlePingCompute)

	auth := newAuthHandlers(options.Auth)
	if auth.enabled() {
		registerAuthRoutes(router, auth)
	}
	if options.Tasks != nil {
		registerTaskRoutes(router, options.Tasks, auth)
	}

	return router
}
