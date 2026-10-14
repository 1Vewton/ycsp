package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ServiceName 随应答一起返回，便于联调时分辨是哪一层。
const ServiceName = "ycsp-数据管理层"

func handleBackendHealth(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{
		"service": ServiceName,
		"status":  "ok",
		"message": "你好，我是数据管理层，运行正常。",
	})
}
