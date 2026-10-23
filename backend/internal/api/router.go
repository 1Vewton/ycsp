package api

import (
	"github.com/gin-gonic/gin"

	"github.com/1Vewton/ycsp/backend/internal/tasks"
)

// Options 是路由需要的依赖。
type Options struct {
	// ComputeBaseURL 是数据处理层地址，用于 /api/ping-compute 与任务执行。
	ComputeBaseURL string
	// Tasks 为 nil 时不注册任务接口（只做健康检查与转发时用得上）。
	Tasks *tasks.Service
}

// NewRouter 注册数据管理层的全部路由。
func NewRouter(options Options) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), recoveryMiddleware())

	router.GET("/api/health", handleBackendHealth)

	computeProxy := newComputeProxy(options.ComputeBaseURL)
	router.GET("/api/ping-compute", computeProxy.handlePingCompute)

	if options.Tasks != nil {
		registerTaskRoutes(router, options.Tasks)
	}

	return router
}
