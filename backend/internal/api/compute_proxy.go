package api

import (
	"encoding/json"
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

// handlePingCompute 请求数据处理层的 /health，用于验证“前端 → 数据管理 → 数据处理”是否通畅。
func (proxy *computeProxy) handlePingCompute(context *gin.Context) {
	targetURL := proxy.baseURL + "/health"

	response, err := proxy.client.Get(targetURL)
	if err != nil {
		writeComputeFailure(context, CodeComputeUnreachable, gin.H{"url": targetURL})
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		writeComputeFailure(context, CodeComputeBadStatus, gin.H{"status_code": response.StatusCode})
		return
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		writeComputeFailure(context, CodeComputeReadFailed, nil)
		return
	}

	var computePayload map[string]any
	if err := json.Unmarshal(responseBody, &computePayload); err != nil {
		writeComputeFailure(context, CodeComputeInvalidJSON, nil)
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"service": ServiceName,
		"status":  "ok",
		"code":    CodeComputeHealthy,
		// 上游完整应答一并返回，方便联调时看清细节。
		"compute": computePayload,
	})
}

// writeComputeFailure 写出“联系数据处理层失败”的应答：只有 code 与参数，没有语言相关文案。
func writeComputeFailure(context *gin.Context, code string, params gin.H) {
	payload := gin.H{
		"service": ServiceName,
		"status":  "error",
		"code":    code,
	}
	if len(params) > 0 {
		payload["params"] = params
	}
	context.JSON(http.StatusBadGateway, payload)
}
