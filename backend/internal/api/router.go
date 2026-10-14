package api

import "github.com/gin-gonic/gin"

// NewRouter 注册数据管理层的全部路由。
func NewRouter(computeBaseURL string) *gin.Engine {
	router := gin.Default()

	router.GET("/api/health", handleBackendHealth)

	computeProxy := newComputeProxy(computeBaseURL)
	router.GET("/api/ping-compute", computeProxy.handlePingCompute)

	return router
}
