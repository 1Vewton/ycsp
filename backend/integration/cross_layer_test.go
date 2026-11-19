//go:build integration

// Package integration_test 验证“Python 报错 → Go 任务失败 → 前端有中文文案”的完整链路。
//
// 需要先启动数据处理层，并通过 YCSP_TEST_COMPUTE_URL 指定地址：
//
//	cd compute
//	uv run uvicorn app.main:app --host 127.0.0.1 --port 8001
//
//	cd backend
//	$env:YCSP_TEST_COMPUTE_URL = "http://127.0.0.1:8001"
//	go test -tags=integration ./integration/...
package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/1Vewton/ycsp/backend/internal/api"
	"github.com/1Vewton/ycsp/backend/internal/compute"
	"github.com/1Vewton/ycsp/backend/internal/store"
	"github.com/1Vewton/ycsp/backend/internal/tasks"
	"github.com/1Vewton/ycsp/backend/internal/testutil"
)

const (
	waitTimeout  = 15 * time.Second
	pollInterval = 50 * time.Millisecond
)

// testHTTPClient 让每次请求都有上限，避免跨层测试卡死。
var testHTTPClient = &http.Client{Timeout: 5 * time.Second}

type taskErrorPayload struct {
	Code   string         `json:"code"`
	Params map[string]any `json:"params"`
}

type taskSnapshot struct {
	ID     uint              `json:"id"`
	Status string            `json:"status"`
	Error  *taskErrorPayload `json:"error"`
}

type taskEnvelope struct {
	Task *taskSnapshot `json:"task"`
}

type errorCase struct {
	name           string
	formula        string
	code           string
	chineseSnippet string
	params         map[string]any
}

// TestPythonErrorsReachTheFrontendCopy 让真实 Python 拒绝三个化学式，
// 检查 Go 任务保存了同一个 code + params，并确认前端语言包有对应中文句子。
func TestPythonErrorsReachTheFrontendCopy(t *testing.T) {
	computeURL := strings.TrimSpace(os.Getenv("YCSP_TEST_COMPUTE_URL"))
	if computeURL == "" {
		t.Skip("set YCSP_TEST_COMPUTE_URL to a running compute service, e.g. http://127.0.0.1:8001")
	}
	waitForCompute(t, computeURL)

	database := testutil.OpenDatabase(t)
	service, err := tasks.NewWithWorkers(database, compute.NewClient(computeURL), 2)
	if err != nil {
		t.Fatalf("NewWithWorkers: %v", err)
	}
	workerContext, cancelWorkers := context.WithCancel(context.Background())
	service.Start(workerContext)
	t.Cleanup(func() {
		cancelWorkers()
		service.Stop()
	})

	server := httptest.NewServer(api.NewRouter(api.Options{
		ComputeBaseURL: computeURL,
		Tasks:          service,
	}))
	t.Cleanup(server.Close)

	zhLocale := readLocale(t, "zh.ts")
	enLocale := readLocale(t, "en.ts")

	cases := []errorCase{
		{
			name:           "charge unbalanced",
			formula:        "NaCl2",
			code:           "formula_charge_unbalanced",
			chineseSnippet: "正负电荷加起来不是 0",
			params:         map[string]any{"formula": "NaCl2"},
		},
		{
			name:           "unsupported element",
			formula:        "Xx2",
			code:           "element_not_supported",
			chineseSnippet: "还没有收录元素",
			params:         map[string]any{"symbol": "Xx", "position": float64(1)},
		},
		{
			name:           "unsupported composition",
			formula:        "H2O",
			code:           "composition_not_supported",
			chineseSnippet: "暂时还不能处理",
			params:         map[string]any{"formula": "H2O"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			task := createTask(t, server.URL, testCase.formula)
			failed := waitForFailure(t, server.URL, task.ID)

			if failed.Error == nil {
				t.Fatalf("failed task has no error: %+v", failed)
			}
			if failed.Error.Code != testCase.code {
				t.Fatalf("error code = %q, want %q", failed.Error.Code, testCase.code)
			}
			for key, want := range testCase.params {
				got, exists := failed.Error.Params[key]
				if !exists || fmt.Sprint(got) != fmt.Sprint(want) {
					t.Errorf(
						"error param %q = %v, want %v (all params: %v)",
						key, got, want, failed.Error.Params,
					)
				}
			}

			assertLocaleContains(t, zhLocale, testCase.code, testCase.chineseSnippet)
			assertLocaleContains(t, enLocale, testCase.code, "")
		})
	}
}

// waitForCompute 等数据处理层真正开始接收请求；CI 启动服务与测试之间有几秒空档。
func waitForCompute(t *testing.T, computeURL string) {
	t.Helper()

	healthURL := strings.TrimRight(computeURL, "/") + "/health"
	deadline := time.Now().Add(waitTimeout)
	for time.Now().Before(deadline) {
		response, err := testHTTPClient.Get(healthURL)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(pollInterval)
	}
	t.Fatalf("compute service at %s is not healthy", computeURL)
}

func createTask(t *testing.T, serverURL string, formula string) taskSnapshot {
	t.Helper()

	response, err := testHTTPClient.Post(
		serverURL+"/api/tasks",
		"application/json",
		strings.NewReader(fmt.Sprintf(`{"formula":%q}`, formula)),
	)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		t.Fatalf("create task status = %d, want %d", response.StatusCode, http.StatusCreated)
	}
	var envelope taskEnvelope
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if envelope.Task == nil || envelope.Task.ID == 0 {
		t.Fatalf("create response has no task: %+v", envelope)
	}
	return *envelope.Task
}

func waitForFailure(t *testing.T, serverURL string, taskID uint) taskSnapshot {
	t.Helper()

	taskURL := fmt.Sprintf("%s/api/tasks/%d", serverURL, taskID)
	deadline := time.Now().Add(waitTimeout)
	lastStatus := ""
	for time.Now().Before(deadline) {
		response, err := testHTTPClient.Get(taskURL)
		if err != nil {
			t.Fatalf("get task %d: %v", taskID, err)
		}
		var envelope taskEnvelope
		decodeErr := json.NewDecoder(response.Body).Decode(&envelope)
		_ = response.Body.Close()
		if decodeErr != nil {
			t.Fatalf("decode task %d: %v", taskID, decodeErr)
		}
		if envelope.Task == nil {
			t.Fatalf("task %d response has no task", taskID)
		}

		lastStatus = envelope.Task.Status
		switch lastStatus {
		case store.TaskStatusFailed:
			return *envelope.Task
		case store.TaskStatusSucceeded:
			t.Fatalf("task %d unexpectedly succeeded", taskID)
		}
		time.Sleep(pollInterval)
	}
	t.Fatalf("task %d did not fail in time, last status = %q", taskID, lastStatus)
	return taskSnapshot{}
}

func readLocale(t *testing.T, fileName string) string {
	t.Helper()

	path := filepath.Join("..", "..", "frontend", "src", "locales", fileName)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read frontend locale %s: %v", path, err)
	}
	return string(content)
}

func assertLocaleContains(t *testing.T, content string, code string, wantedText string) {
	t.Helper()

	if !strings.Contains(content, code+":") {
		t.Errorf("frontend locale is missing the key %q", code)
	}
	if wantedText != "" && !strings.Contains(content, wantedText) {
		t.Errorf("frontend locale is missing the Chinese copy %q for %q", wantedText, code)
	}
}
