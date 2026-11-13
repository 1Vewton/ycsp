package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/1Vewton/ycsp/backend/internal/api"
	"github.com/1Vewton/ycsp/backend/internal/codes"
	"github.com/1Vewton/ycsp/backend/internal/compute"
	"github.com/1Vewton/ycsp/backend/internal/session"
	"github.com/1Vewton/ycsp/backend/internal/store"
	"github.com/1Vewton/ycsp/backend/internal/tasks"
	"github.com/1Vewton/ycsp/backend/internal/testutil"
	"github.com/1Vewton/ycsp/backend/internal/users"
)

const ionicPredictionBody = `{
  "formula": "NaCl",
  "family": "ionic",
  "candidates": [{
    "prototype": "rock_salt",
    "coordination_number": 6,
    "rule": {"code": "rule_radius_ratio", "params": {"ratio": 0.564}},
    "caveats": [],
    "cell": {
      "edge_a_angstrom": 5.66, "edge_b_angstrom": 5.66, "edge_c_angstrom": 5.66,
      "angle_alpha_degree": 90, "angle_beta_degree": 90, "angle_gamma_degree": 90
    },
    "atoms": [{"symbol": "Na", "fractional_coordinates": [0, 0, 0]}]
  }]
}`

type stubPredictor struct {
	predict func(ctx context.Context, formula string) (compute.Prediction, error)
}

func (stub stubPredictor) Predict(ctx context.Context, formula string) (compute.Prediction, error) {
	return stub.predict(ctx, formula)
}

func successPredictor() stubPredictor {
	return stubPredictor{
		predict: func(_ context.Context, formula string) (compute.Prediction, error) {
			var prediction compute.Prediction
			if err := json.Unmarshal([]byte(ionicPredictionBody), &prediction); err != nil {
				return compute.Prediction{}, err
			}
			prediction.Formula = formula
			return prediction, nil
		},
	}
}

// newTaskRouter 建一个带任务接口的路由；startWorkers 决定后台是否真的算。
func newTaskRouter(t *testing.T, predictor tasks.Predictor, startWorkers bool) http.Handler {
	t.Helper()

	router, _ := newTaskRouterWithDatabase(t, predictor, startWorkers)
	return router
}

func newTaskRouterWithDatabase(
	t *testing.T, predictor tasks.Predictor, startWorkers bool,
) (http.Handler, *gorm.DB) {
	t.Helper()

	database := testutil.OpenDatabase(t)

	service := tasks.New(database, predictor)
	if startWorkers {
		workerContext, cancelWorkers := context.WithCancel(context.Background())
		service.Start(workerContext)
		t.Cleanup(func() {
			cancelWorkers()
			service.Stop()
		})
	}

	router := api.NewRouter(api.Options{
		ComputeBaseURL: "http://127.0.0.1:1",
		Tasks:          service,
	})
	return router, database
}

func performRequest(
	t *testing.T, router http.Handler, method string, path string, body string,
) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()

	var request *http.Request
	if body == "" {
		request = httptest.NewRequest(method, path, nil)
	} else {
		request = httptest.NewRequest(method, path, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response is not JSON: %v (body=%q)", err, recorder.Body.String())
	}
	return recorder, payload
}

func taskFrom(payload map[string]any) map[string]any {
	task, ok := payload["task"].(map[string]any)
	if !ok {
		return nil
	}
	return task
}

func TestCreateTaskReturnsAPendingTask(t *testing.T) {
	router := newTaskRouter(t, successPredictor(), false)

	recorder, payload := performRequest(t, router, http.MethodPost, "/api/tasks", `{"formula":" NaCl "}`)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if payload["code"] != codes.TaskCreated {
		t.Errorf("code = %v, want %q", payload["code"], codes.TaskCreated)
	}

	task := taskFrom(payload)
	if task == nil {
		t.Fatalf("payload has no task: %v", payload)
	}
	if task["status"] != store.TaskStatusPending {
		t.Errorf("status = %v, want %q", task["status"], store.TaskStatusPending)
	}
	if task["formula"] != "NaCl" {
		t.Errorf("formula = %v, want the trimmed input", task["formula"])
	}
	if _, hasError := task["error"]; hasError {
		t.Errorf("pending task should not carry an error: %v", task["error"])
	}
}

func TestCreateTaskRejectsEmptyFormula(t *testing.T) {
	router := newTaskRouter(t, successPredictor(), false)

	recorder, payload := performRequest(t, router, http.MethodPost, "/api/tasks", `{"formula":"   "}`)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if payload["code"] != codes.TaskFormulaEmpty {
		t.Errorf("code = %v, want %q", payload["code"], codes.TaskFormulaEmpty)
	}
}

func TestCreateTaskRejectsTooLongFormula(t *testing.T) {
	router := newTaskRouter(t, successPredictor(), false)
	tooLong := strings.Repeat("Na", 33) // 66 个字符，超过 64 的上限

	recorder, payload := performRequest(
		t, router, http.MethodPost, "/api/tasks", fmt.Sprintf(`{"formula":%q}`, tooLong),
	)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if payload["code"] != codes.TaskFormulaTooLong {
		t.Errorf("code = %v, want %q", payload["code"], codes.TaskFormulaTooLong)
	}
	params, ok := payload["params"].(map[string]any)
	if !ok || params["max_length"] != float64(64) {
		t.Errorf("params = %v, want max_length 64", payload["params"])
	}
}

func TestCreateTaskRejectsInvalidJSON(t *testing.T) {
	router := newTaskRouter(t, successPredictor(), false)

	recorder, payload := performRequest(t, router, http.MethodPost, "/api/tasks", "not json")

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if payload["code"] != codes.RequestInvalidBody {
		t.Errorf("code = %v, want %q", payload["code"], codes.RequestInvalidBody)
	}
}

func TestGetTaskReportsNotFound(t *testing.T) {
	router := newTaskRouter(t, successPredictor(), false)

	for _, path := range []string{"/api/tasks/9999", "/api/tasks/abc"} {
		recorder, payload := performRequest(t, router, http.MethodGet, path, "")

		if recorder.Code != http.StatusNotFound {
			t.Errorf("%s: status = %d, want %d", path, recorder.Code, http.StatusNotFound)
		}
		if payload["code"] != codes.TaskNotFound {
			t.Errorf("%s: code = %v, want %q", path, payload["code"], codes.TaskNotFound)
		}
	}
}

func TestResultIsNotReadyWhilePending(t *testing.T) {
	router := newTaskRouter(t, successPredictor(), false)
	_, created := performRequest(t, router, http.MethodPost, "/api/tasks", `{"formula":"NaCl"}`)
	taskID := int(taskFrom(created)["id"].(float64))

	recorder, payload := performRequest(
		t, router, http.MethodGet, fmt.Sprintf("/api/tasks/%d/result", taskID), "",
	)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}
	if payload["code"] != codes.TaskNotReady {
		t.Errorf("code = %v, want %q", payload["code"], codes.TaskNotReady)
	}
	params, ok := payload["params"].(map[string]any)
	if !ok || params["status"] != store.TaskStatusPending {
		t.Errorf("params = %v, want status pending", payload["params"])
	}
}

func TestCreatePollAndFetchTheResult(t *testing.T) {
	router := newTaskRouter(t, successPredictor(), true)

	_, created := performRequest(t, router, http.MethodPost, "/api/tasks", `{"formula":"NaCl"}`)
	taskID := int(taskFrom(created)["id"].(float64))

	deadline := time.Now().Add(3 * time.Second)
	status := ""
	for time.Now().Before(deadline) {
		_, payload := performRequest(t, router, http.MethodGet, fmt.Sprintf("/api/tasks/%d", taskID), "")
		status, _ = taskFrom(payload)["status"].(string)
		if status == store.TaskStatusSucceeded {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if status != store.TaskStatusSucceeded {
		t.Fatalf("task did not succeed in time, last status = %q", status)
	}

	recorder, payload := performRequest(
		t, router, http.MethodGet, fmt.Sprintf("/api/tasks/%d/result", taskID), "",
	)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if payload["code"] != codes.TaskResult {
		t.Errorf("code = %v, want %q", payload["code"], codes.TaskResult)
	}

	result, ok := payload["result"].(map[string]any)
	if !ok {
		t.Fatalf("payload has no result: %v", payload)
	}
	if result["prototype"] != "rock_salt" || result["family"] != "ionic" {
		t.Errorf("result = %v", result)
	}
	cell, ok := result["cell"].(map[string]any)
	if !ok || cell["edge_a_angstrom"] != 5.66 {
		t.Errorf("cell = %v", result["cell"])
	}
	atoms, ok := result["atoms"].([]any)
	if !ok || len(atoms) != 1 {
		t.Errorf("atoms = %v", result["atoms"])
	}
}

func TestCancelPendingTaskAndRejectTheSecondAttempt(t *testing.T) {
	router := newTaskRouter(t, successPredictor(), false)
	_, created := performRequest(t, router, http.MethodPost, "/api/tasks", `{"formula":"NaCl"}`)
	taskID := int(taskFrom(created)["id"].(float64))
	path := fmt.Sprintf("/api/tasks/%d/cancel", taskID)

	recorder, payload := performRequest(t, router, http.MethodPost, path, "")
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if payload["code"] != codes.TaskCancelled {
		t.Errorf("code = %v, want %q", payload["code"], codes.TaskCancelled)
	}
	if taskFrom(payload)["status"] != store.TaskStatusCancelled {
		t.Errorf("status = %v, want cancelled", taskFrom(payload)["status"])
	}

	conflictRecorder, conflictPayload := performRequest(t, router, http.MethodPost, path, "")
	if conflictRecorder.Code != http.StatusConflict {
		t.Fatalf("second cancel status = %d, want %d", conflictRecorder.Code, http.StatusConflict)
	}
	if conflictPayload["code"] != codes.TaskCancelConflict {
		t.Errorf("code = %v, want %q", conflictPayload["code"], codes.TaskCancelConflict)
	}
	params, ok := conflictPayload["params"].(map[string]any)
	if !ok || params["status"] != store.TaskStatusCancelled {
		t.Errorf("params = %v, want status cancelled", conflictPayload["params"])
	}
}

func TestCreateTaskStoresRequestID(t *testing.T) {
	router, database := newTaskRouterWithDatabase(t, successPredictor(), false)

	request := httptest.NewRequest(
		http.MethodPost, "/api/tasks", strings.NewReader(`{"formula":"NaCl"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Request-ID", "web-test-789")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if got := recorder.Header().Get("X-Request-ID"); got != "web-test-789" {
		t.Errorf("response request id = %q, want web-test-789", got)
	}

	var stored store.PredictionTask
	if err := database.First(&stored).Error; err != nil {
		t.Fatalf("load task: %v", err)
	}
	if stored.TraceID != "web-test-789" {
		t.Errorf("stored trace_id = %q, want web-test-789", stored.TraceID)
	}
}

func newAuthedTaskRouter(t *testing.T) (http.Handler, *tasks.Service) {
	t.Helper()

	database := testutil.OpenDatabase(t)
	service := tasks.New(database, successPredictor())
	workerContext, cancelWorkers := context.WithCancel(context.Background())
	service.Start(workerContext)
	t.Cleanup(func() {
		cancelWorkers()
		service.Stop()
	})

	router := api.NewRouter(api.Options{
		ComputeBaseURL: "http://127.0.0.1:1",
		Tasks:          service,
		Auth: api.AuthOptions{
			Users:           users.New(database),
			Sessions:        session.NewMemoryStore(),
			SessionLifetime: time.Hour,
		},
	})
	return router, service
}

func TestTaskHistoryRequiresSignIn(t *testing.T) {
	router, _ := newAuthedTaskRouter(t)

	recorder, payload := performAuthRequest(
		t, router, http.MethodGet, "/api/users/me/tasks", "", nil,
	)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
	if payload["code"] != codes.NotSignedIn {
		t.Errorf("code = %v, want %q", payload["code"], codes.NotSignedIn)
	}
}

func TestTaskHistoryListsOnlyTheSignedInUsersTasks(t *testing.T) {
	router, _ := newAuthedTaskRouter(t)
	cookie := registerUser(t, router, "小明")

	createRecorder, createPayload := performAuthRequest(
		t, router, http.MethodPost, "/api/tasks", `{"formula":"NaCl"}`, cookie,
	)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", createRecorder.Code, createRecorder.Body.String())
	}
	taskID := uint(createPayload["task"].(map[string]any)["id"].(float64))

	deadline := time.Now().Add(3 * time.Second)
	status := ""
	for time.Now().Before(deadline) {
		_, payload := performAuthRequest(
			t, router, http.MethodGet, fmt.Sprintf("/api/tasks/%d", taskID), "", cookie,
		)
		status, _ = payload["task"].(map[string]any)["status"].(string)
		if status == store.TaskStatusSucceeded {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if status != store.TaskStatusSucceeded {
		t.Fatalf("task did not succeed in time, last status = %q", status)
	}

	historyRecorder, historyPayload := performAuthRequest(
		t, router, http.MethodGet, "/api/users/me/tasks", "", cookie,
	)
	if historyRecorder.Code != http.StatusOK {
		t.Fatalf("history status = %d, body = %s", historyRecorder.Code, historyRecorder.Body.String())
	}
	if historyPayload["code"] != codes.TaskHistory {
		t.Errorf("code = %v, want %q", historyPayload["code"], codes.TaskHistory)
	}
	if historyPayload["total"] != float64(1) {
		t.Errorf("total = %v, want 1", historyPayload["total"])
	}
	items, ok := historyPayload["tasks"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("tasks = %v, want one item", historyPayload["tasks"])
	}
	item := items[0].(map[string]any)
	if item["formula"] != "NaCl" || item["status"] != store.TaskStatusSucceeded {
		t.Errorf("history item = %v", item)
	}
	result, ok := item["result"].(map[string]any)
	if !ok || result["prototype"] != "rock_salt" {
		t.Errorf("history result = %v, want rock_salt", item["result"])
	}

	// 匿名任务没有归属，不应出现在个人历史里。
	anonymousRecorder, _ := performAuthRequest(
		t, router, http.MethodPost, "/api/tasks", `{"formula":"KCl"}`, nil,
	)
	if anonymousRecorder.Code != http.StatusCreated {
		t.Fatalf("anonymous create status = %d", anonymousRecorder.Code)
	}
	_, afterAnonymous := performAuthRequest(
		t, router, http.MethodGet, "/api/users/me/tasks", "", cookie,
	)
	if afterAnonymous["total"] != float64(1) {
		t.Errorf("total after anonymous task = %v, want 1", afterAnonymous["total"])
	}
}
