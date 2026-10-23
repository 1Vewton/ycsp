package compute

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// predictTimeout 是单次预测的超时上限：计算本身是毫秒级，10 秒足够兜住卡死的情况，
// 也避免后端为一次请求无限等待。
const predictTimeout = 10 * time.Second

// Client 调用数据处理层的 POST /predict。
type Client struct {
	predictURL string
	httpClient *http.Client
}

// NewClient 按数据处理层地址创建客户端。
func NewClient(baseURL string) *Client {
	return &Client{
		predictURL: strings.TrimRight(baseURL, "/") + "/predict",
		httpClient: &http.Client{Timeout: predictTimeout},
	}
}

// Predict 把化学式发给数据处理层；失败时返回可翻译的错误（见 Describe）。
func (client *Client) Predict(ctx context.Context, formula string) (Prediction, error) {
	requestBody, err := json.Marshal(map[string]string{"formula": formula})
	if err != nil {
		return Prediction{}, fmt.Errorf("compute: encode request: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx, http.MethodPost, client.predictURL, bytes.NewReader(requestBody),
	)
	if err != nil {
		return Prediction{}, fmt.Errorf("compute: build request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return Prediction{}, &TransportError{URL: client.predictURL, Err: err}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return Prediction{}, &ProtocolError{
			Problem:    ProblemReadBody,
			StatusCode: response.StatusCode,
			Err:        err,
		}
	}

	// 上游用 400 + {status:error, code, params} 表示“这个化学式算不了”：
	// 这种错误对学生有意义（要改化学式），把它的 code 原样透传。
	if response.StatusCode == http.StatusBadRequest {
		var rejection rejectionPayload
		if err := json.Unmarshal(responseBody, &rejection); err == nil &&
			rejection.Status == "error" && rejection.Code != "" {
			return Prediction{}, &PredictionError{Code: rejection.Code, Params: rejection.Params}
		}
		return Prediction{}, &ProtocolError{
			Problem:    ProblemInvalidJSON,
			StatusCode: response.StatusCode,
		}
	}

	// 状态码不对时先报状态码：即使应答体是空的也能说清问题出在哪一步。
	if response.StatusCode != http.StatusOK {
		return Prediction{}, &ProtocolError{
			Problem:    ProblemStatusCode,
			StatusCode: response.StatusCode,
		}
	}

	var envelope struct {
		Prediction Prediction `json:"prediction"`
	}
	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		return Prediction{}, &ProtocolError{
			Problem:    ProblemInvalidJSON,
			StatusCode: response.StatusCode,
			Err:        err,
		}
	}
	if len(envelope.Prediction.Candidates) == 0 {
		return Prediction{}, &ProtocolError{
			Problem:    ProblemEmptyResult,
			StatusCode: response.StatusCode,
		}
	}
	return envelope.Prediction, nil
}

// rejectionPayload 是上游拒绝输入时的应答形状。
type rejectionPayload struct {
	Status string         `json:"status"`
	Code   string         `json:"code"`
	Params map[string]any `json:"params"`
}
