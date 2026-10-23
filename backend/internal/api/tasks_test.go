package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/1Vewton/ycsp/backend/internal/api"
	"github.com/1Vewton/ycsp/backend/internal/codes"
	"github.com/1Vewton/ycsp/backend/internal/compute"
	"github.com/1Vewton/ycsp/backend/internal/store"
	"github.com/1Vewton/ycsp/backend/internal/tasks"
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

	database, err := store.Open(store.Settings{
		Driver: store.DriverSQLite,
		DSN:    filepath.Join(t.TempDir(), "api-tasks.db"),
	})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := store.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() {
		sqlDatabase, err := database.DB()
		if err == nil {
			_ = sqlDatabase.Close()
		}
	})

	service := tasks.New(database, predictor)
	if startWorkers {
		workerContext, cancelWorkers := context.WithCancel(context.Background())
		service.Start(workerContext)
		t.Cleanup(func() {
			cancelWorkers()
			service.Stop()
		})
	}

	return api.NewRouter(api.Options{
		ComputeBaseURL: "http://127.0.0.1:1",
		Tasks:          service,
	})
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
