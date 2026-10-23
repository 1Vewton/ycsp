package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/1Vewton/ycsp/backend/internal/codes"
)

// writeError 输出统一的错误应答：只有 code 与 params，没有任何语言的文案。
func writeError(context *gin.Context, statusCode int, code string, params gin.H) {
	payload := gin.H{
		"service": ServiceName,
		"status":  "error",
		"code":    code,
	}
	if len(params) > 0 {
		payload["params"] = params
	}
	context.JSON(statusCode, payload)
}

// writeServerError 记录技术细节（日志里），给前端只留一句可翻译的提示。
func writeServerError(context *gin.Context, err error) {
	log.Printf("%s handler error: %v", ServiceName, err)
	writeError(context, http.StatusInternalServerError, codes.InternalError, nil)
}

// recoveryMiddleware 把 panic 变成统一应答；完整堆栈由 gin 的日志输出保留。
func recoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(context *gin.Context, recovered any) {
		log.Printf("%s panic recovered: %v", ServiceName, recovered)
		writeError(context, http.StatusInternalServerError, codes.InternalError, nil)
	})
}
