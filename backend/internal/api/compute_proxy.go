package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 调用数据处理层的超时上限，避免上游卡住时网页一直转圈。
const computeRequestTimeout = 5 * time.Second

type computeProxy struct {
	baseURL string
	client  *http.Client
}

func newComputeProxy(baseURL string) *computeProxy {
	return &computeProxy{
		baseURL: baseURL,
		client:  &http.Client{Timeout: computeRequestTimeout},
	}
}

// handlePingCompute 请求数据处理层的 /health，把它的中文问候转达给前端。
func (proxy *computeProxy) handlePingCompute(context *gin.Context) {
	targetURL := proxy.baseURL + "/health"

	response, err := proxy.client.Get(targetURL)
	if err != nil {
		writeComputeFailure(context, fmt.Sprintf("联系不上数据处理层（%s），请确认它已经启动。", targetURL))
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		writeComputeFailure(context, fmt.Sprintf("数据处理层返回了异常状态码 %d，请查看它的运行日志。", response.StatusCode))
		return
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		writeComputeFailure(context, "读取数据处理层的应答时出错，请稍后重试。")
		return
	}

	var computePayload map[string]any
	if err := json.Unmarshal(responseBody, &computePayload); err != nil {
		writeComputeFailure(context, "数据处理层的应答不是合法 JSON，请检查它的服务日志。")
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"service": ServiceName,
		"status":  "ok",
		"message": computePayload["message"],
		"compute": computePayload,
	})
}

// writeComputeFailure 统一写出“联系数据处理层失败”的中文应答。
func writeComputeFailure(context *gin.Context, message string) {
	context.JSON(http.StatusBadGateway, gin.H{
		"service": ServiceName,
		"status":  "error",
		"message": message,
	})
}
