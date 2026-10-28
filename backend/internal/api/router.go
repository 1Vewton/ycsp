package api

import (
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
}

// NewRouter 注册数据管理层的全部路由。
func NewRouter(options Options) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), recoveryMiddleware())

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
