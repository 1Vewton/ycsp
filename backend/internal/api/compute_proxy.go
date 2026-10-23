package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/1Vewton/ycsp/backend/internal/codes"
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
		writeError(context, http.StatusBadGateway, codes.ComputeUnreachable, gin.H{"url": targetURL})
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		writeError(context, http.StatusBadGateway, codes.ComputeBadStatus, gin.H{
			"status_code": response.StatusCode,
		})
		return
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		writeError(context, http.StatusBadGateway, codes.ComputeReadFailed, nil)
		return
	}

	var computePayload map[string]any
	if err := json.Unmarshal(responseBody, &computePayload); err != nil {
		writeError(context, http.StatusBadGateway, codes.ComputeInvalidJSON, nil)
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"service": ServiceName,
		"status":  "ok",
		"code":    codes.ComputeHealthy,
		// 上游完整应答一并返回，方便联调时看清细节。
		"compute": computePayload,
	})
}
