package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ServiceName 随应答一起返回，便于联调时分辨是哪一层。
const ServiceName = "ycsp-backend"

func handleBackendHealth(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{
		"service": ServiceName,
		"status":  "ok",
		"code":    CodeBackendHealthy,
	})
}
