package compute

import (
	"errors"
	"fmt"

	"github.com/1Vewton/ycsp/backend/internal/codes"
)

// PredictionError 表示数据处理层明确拒绝了这次输入（例如化学式电荷不平衡）。
// Code 与 Params 原样透传：前端已有对应文案，后端不再翻译。
type PredictionError struct {
	Code   string
	Params map[string]any
}

func (err *PredictionError) Error() string {
	return fmt.Sprintf("compute: upstream rejected the input, code=%s", err.Code)
}

// Problem 说明应答在哪一步不符合约定。
type Problem string

const (
	ProblemReadBody    Problem = "read_body"
	ProblemStatusCode  Problem = "status_code"
	ProblemInvalidJSON Problem = "invalid_json"
	ProblemEmptyResult Problem = "empty_result"
)

// ProtocolError 表示数据处理层的应答不符合约定（读不出、状态码不对、JSON 坏了、没有结果）。
type ProtocolError struct {
	Problem    Problem
	StatusCode int
	Err        error
}

func (err *ProtocolError) Error() string {
	if err.Err != nil {
		return fmt.Sprintf("compute: protocol problem %s: %v", err.Problem, err.Err)
	}
	return fmt.Sprintf("compute: protocol problem %s (status %d)", err.Problem, err.StatusCode)
}

// TransportError 表示连不上数据处理层：没启动、端口不对、超时。
type TransportError struct {
	URL string
	Err error
}

func (err *TransportError) Error() string {
	return fmt.Sprintf("compute: cannot reach %s: %v", err.URL, err.Err)
}

// Describe 把客户端错误翻成前端可用的 code 与 params。
func Describe(err error) (string, map[string]any) {
	var predictionError *PredictionError
	if errors.As(err, &predictionError) {
		return predictionError.Code, predictionError.Params
	}

	var transportError *TransportError
	if errors.As(err, &transportError) {
		return codes.ComputeUnreachable, map[string]any{"url": transportError.URL}
	}

	var protocolError *ProtocolError
	if errors.As(err, &protocolError) {
		switch protocolError.Problem {
		case ProblemReadBody:
			return codes.ComputeReadFailed, nil
		case ProblemStatusCode:
			return codes.ComputeBadStatus, map[string]any{"status_code": protocolError.StatusCode}
		case ProblemEmptyResult:
			return codes.ComputeNoResult, nil
		default:
			return codes.ComputeInvalidJSON, nil
		}
	}

	return codes.InternalError, nil
}
