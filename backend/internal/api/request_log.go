package api

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/1Vewton/ycsp/backend/internal/traceid"
)

// requestIDMiddleware 给每个请求确定一个编号：
// 优先使用前端传来的 X-Request-ID，没有或格式不合法就生成一个。
// 编号同时写回响应头，前端控制台与三层日志可以对上。
func requestIDMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		requestID := traceid.Normalize(context.GetHeader(traceid.HeaderName))
		if requestID == "" {
			requestID = traceid.New("api")
		}
		context.Writer.Header().Set(traceid.HeaderName, requestID)
		context.Request = context.Request.WithContext(
			traceid.With(context.Request.Context(), requestID),
		)
		context.Next()
	}
}

// requestLogger 记录每次请求的结果。
// 只记方法、路径、状态码、耗时与限流使用的客户端地址，不记请求体与查询参数：登录密码不能进日志。
func requestLogger(addresses clientAddressResolver) gin.HandlerFunc {
	return func(context *gin.Context) {
		startedAt := time.Now()
		context.Next()

		log.Printf(
			"http: request_id=%s client_ip=%s method=%s path=%s status=%d duration_ms=%d",
			traceid.From(context.Request.Context()),
			addresses.Address(context.Request),
			context.Request.Method,
			context.Request.URL.Path,
			context.Writer.Status(),
			time.Since(startedAt).Milliseconds(),
		)
	}
}
