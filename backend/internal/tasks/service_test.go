package tasks_test

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/1Vewton/ycsp/backend/internal/compute"
	"github.com/1Vewton/ycsp/backend/internal/store"
	"github.com/1Vewton/ycsp/backend/internal/tasks"
)

type stubPredictor struct {
	predict func(ctx context.Context, formula string) (compute.Prediction, error)
}

func (stub stubPredictor) Predict(ctx context.Context, formula string) (compute.Prediction, error) {
	return stub.predict(ctx, formula)
}

func samplePrediction(formula string) compute.Prediction {
	return compute.Prediction{
		Formula: formula,
		Family:  "ionic",
		Candidates: []compute.Candidate{{
			Prototype:          "rock_salt",
			CoordinationNumber: 6,
			Rule:               json.RawMessage(`{"code":"rule_radius_ratio","params":{"ratio":0.564}}`),
			Caveats:            json.RawMessage(`[]`),
			Cell: compute.Cell{
				EdgeAAngstrom: 5.66, EdgeBAngstrom: 5.66, EdgeCAngstrom: 5.66,
				AngleAlphaDegree: 90, AngleBetaDegree: 90, AngleGammaDegree: 90,
			},
			Atoms: json.RawMessage(`[{"symbol":"Na","fractional_coordinates":[0,0,0]}]`),
		}},
	}
}

func openTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	database, err := store.Open(store.Settings{
		Driver: store.DriverSQLite,
		DSN:    filepath.Join(t.TempDir(), "tasks-test.db"),
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
	return database
}

func startService(t *testing.T, database *gorm.DB, predictor tasks.Predictor) *tasks.Service {
	t.Helper()

	service := tasks.New(database, predictor)
	workerContext, cancelWorkers := context.WithCancel(context.Background())
	service.Start(workerContext)
	t.Cleanup(func() {
		cancelWorkers()
		service.Stop()
	})
	return service
}

// waitForStatus 轮询数据库直到任务到达期望状态（后台执行是异步的）。
func waitForStatus(t *testing.T, database *gorm.DB, taskID uint, want string) store.PredictionTask {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var task store.PredictionTask
		if err := database.First(&task, taskID).Error; err != nil {
			t.Fatalf("load task %d: %v", taskID, err)
		}
		if task.Status == want {
			return task
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("task %d did not reach %q in time", taskID, want)
	return store.PredictionTask{}
}

func TestCreateStoresAPendingTask(t *testing.T) {
	database := openTestDatabase(t)
	service := tasks.New(database, stubPredictor{})

	task, err := service.Create("NaCl")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if task.ID == 0 || task.Status != store.TaskStatusPending || task.Formula != "NaCl" {
		t.Fatalf("task = %+v", task)
	}

	var stored store.PredictionTask
	if err := database.First(&stored, task.ID).Error; err != nil {
		t.Fatalf("load stored task: %v", err)
	}
	if stored.Status != store.TaskStatusPending {
		t.Errorf("stored status = %q, want %q", stored.Status, store.TaskStatusPending)
	}
}

func TestServiceExecutesTheTaskAndStoresTheResult(t *testing.T) {
	database := openTestDatabase(t)
	service := startService(t, database, stubPredictor{
		predict: func(_ context.Context, formula string) (compute.Prediction, error) {
			return samplePrediction(formula), nil
		},
	})

	task, err := service.Create("NaCl")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	finished := waitForStatus(t, database, task.ID, store.TaskStatusSucceeded)
	if finished.StartedAt == nil || finished.FinishedAt == nil {
		t.Errorf("timestamps not set: %+v", finished)
	}
	if finished.DurationMS == nil {
		t.Errorf("duration_ms not set: %+v", finished)
	}

	result, err := service.Result(task.ID)
	if err != nil {
		t.Fatalf("Result: %v", err)
	}
	if result.Family != "ionic" || result.Prototype != "rock_salt" || result.CoordinationNumber != 6 {
		t.Errorf("result = %+v", result)
	}
	if result.EdgeAAngstrom != 5.66 || result.AngleGammaDegree != 90 {
		t.Errorf("cell not stored: %+v", result)
	}
	if !strings.Contains(result.Atoms, `"symbol":"Na"`) {
		t.Errorf("atoms = %s", result.Atoms)
	}
	if result.Caveats != "[]" {
		t.Errorf("caveats = %s, want []", result.Caveats)
	}
}

func TestPredictionErrorIsStoredOnTheTask(t *testing.T) {
	database := openTestDatabase(t)
	service := startService(t, database, stubPredictor{
		predict: func(_ context.Context, _ string) (compute.Prediction, error) {
			return compute.Prediction{}, &compute.PredictionError{
				Code:   "formula_charge_unbalanced",
				Params: map[string]any{"formula": "NaCl2"},
			}
		},
	})

	task, err := service.Create("NaCl2")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	failed := waitForStatus(t, database, task.ID, store.TaskStatusFailed)
	if failed.ErrorCode != "formula_charge_unbalanced" {
		t.Errorf("error_code = %q", failed.ErrorCode)
	}
	if !strings.Contains(failed.ErrorParams, `"formula":"NaCl2"`) {
		t.Errorf("error_params = %s", failed.ErrorParams)
	}

	_, resultErr := service.Result(task.ID)
	var notReady *tasks.NotReadyError
	if !errors.As(resultErr, &notReady) || notReady.Status != store.TaskStatusFailed {
		t.Errorf("Result error = %v, want NotReadyError(failed)", resultErr)
	}
}

func TestCancelledTaskIsNotExecuted(t *testing.T) {
	database := openTestDatabase(t)
	executed := make(chan string, 1)
	service := tasks.New(database, stubPredictor{
		predict: func(_ context.Context, formula string) (compute.Prediction, error) {
			executed <- formula
			return samplePrediction(formula), nil
		},
	})

	task, err := service.Create("NaCl")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	cancelled, err := service.Cancel(task.ID)
	if err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if cancelled.Status != store.TaskStatusCancelled {
		t.Fatalf("status = %q, want cancelled", cancelled.Status)
	}

	// 取消之后才启动 worker：队列里还有这个编号，但 claim 必须失败。
	workerContext, cancelWorkers := context.WithCancel(context.Background())
	service.Start(workerContext)
	t.Cleanup(func() {
		cancelWorkers()
		service.Stop()
	})

	select {
	case formula := <-executed:
		t.Fatalf("cancelled task was executed with %q", formula)
	case <-time.After(200 * time.Millisecond):
	}

	var stored store.PredictionTask
	if err := database.First(&stored, task.ID).Error; err != nil {
		t.Fatalf("load task: %v", err)
	}
	if stored.Status != store.TaskStatusCancelled {
		t.Errorf("status = %q, want cancelled", stored.Status)
	}
}

func TestCancelRejectsFinishedTask(t *testing.T) {
	database := openTestDatabase(t)
	service := startService(t, database, stubPredictor{
		predict: func(_ context.Context, formula string) (compute.Prediction, error) {
			return samplePrediction(formula), nil
		},
	})

	task, err := service.Create("NaCl")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	waitForStatus(t, database, task.ID, store.TaskStatusSucceeded)

	_, cancelErr := service.Cancel(task.ID)
	var conflict *tasks.CancelConflictError
	if !errors.As(cancelErr, &conflict) || conflict.Status != store.TaskStatusSucceeded {
		t.Errorf("Cancel error = %v, want CancelConflictError(succeeded)", cancelErr)
	}
}

func TestGetAndResultReportMissingTask(t *testing.T) {
	database := openTestDatabase(t)
	service := tasks.New(database, stubPredictor{})

	if _, err := service.Get(9999); !errors.Is(err, tasks.ErrTaskNotFound) {
		t.Errorf("Get error = %v, want ErrTaskNotFound", err)
	}
	if _, err := service.Result(9999); !errors.Is(err, tasks.ErrTaskNotFound) {
		t.Errorf("Result error = %v, want ErrTaskNotFound", err)
	}
}

func TestResultIsNotReadyWhilePending(t *testing.T) {
	database := openTestDatabase(t)
	service := tasks.New(database, stubPredictor{})

	task, err := service.Create("NaCl")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, resultErr := service.Result(task.ID)
	var notReady *tasks.NotReadyError
	if !errors.As(resultErr, &notReady) || notReady.Status != store.TaskStatusPending {
		t.Errorf("Result error = %v, want NotReadyError(pending)", resultErr)
	}
}

func TestStartRequeuesTasksLeftRunning(t *testing.T) {
	database := openTestDatabase(t)
	startedAt := time.Now().UTC().Add(-time.Minute)
	// 模拟上次进程在计算中途退出：状态停在 running。
	task := store.PredictionTask{
		Formula:   "NaCl",
		Status:    store.TaskStatusRunning,
		StartedAt: &startedAt,
	}
	if err := database.Create(&task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}

	startService(t, database, stubPredictor{
		predict: func(_ context.Context, formula string) (compute.Prediction, error) {
			return samplePrediction(formula), nil
		},
	})

	waitForStatus(t, database, task.ID, store.TaskStatusSucceeded)
}

func TestStartRequeuesPendingTasks(t *testing.T) {
	database := openTestDatabase(t)
	// 直接落库的 pending 任务没有进队列（例如队列满时），启动时要补上。
	task := store.PredictionTask{Formula: "NaCl", Status: store.TaskStatusPending}
	if err := database.Create(&task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}

	startService(t, database, stubPredictor{
		predict: func(_ context.Context, formula string) (compute.Prediction, error) {
			return samplePrediction(formula), nil
		},
	})

	waitForStatus(t, database, task.ID, store.TaskStatusSucceeded)
}
