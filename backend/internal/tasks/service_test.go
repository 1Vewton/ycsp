package tasks_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/1Vewton/ycsp/backend/internal/codes"
	"github.com/1Vewton/ycsp/backend/internal/compute"
	"github.com/1Vewton/ycsp/backend/internal/store"
	"github.com/1Vewton/ycsp/backend/internal/tasks"
	"github.com/1Vewton/ycsp/backend/internal/testutil"
	"github.com/1Vewton/ycsp/backend/internal/traceid"
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
	return testutil.OpenDatabase(t)
}

func startService(t *testing.T, database *gorm.DB, predictor tasks.Predictor) *tasks.Service {
	t.Helper()
	return startServiceWithWorkers(t, database, predictor, tasks.DefaultWorkerCount)
}

func startServiceWithWorkers(
	t *testing.T, database *gorm.DB, predictor tasks.Predictor, workers int,
) *tasks.Service {
	t.Helper()

	service, err := tasks.NewWithWorkers(database, predictor, workers)
	if err != nil {
		t.Fatalf("NewWithWorkers(%d): %v", workers, err)
	}
	workerContext, cancelWorkers := context.WithCancel(context.Background())
	service.Start(workerContext)
	t.Cleanup(func() {
		cancelWorkers()
		service.Stop()
	})
	return service
}

const (
	// taskStatusWaitTimeout 是等待异步任务结束的上限；CI 上 SQLite 写入较慢时留足余量。
	taskStatusWaitTimeout = 5 * time.Second
	// workerStartTimeout 是等待 worker 真正开始第一条任务的上限。
	workerStartTimeout = time.Second
	// concurrentOverflowTasks 让并发任务数超过队列容量，覆盖多人同时使用时的排队与补扫。
	concurrentOverflowTasks = 16
	// concurrentComputeDelay 模拟每次计算占用 worker 的时间。
	concurrentComputeDelay = 2 * time.Millisecond
)

// waitForStatus 轮询数据库直到任务到达期望状态（后台执行是异步的）。
func waitForStatus(t *testing.T, database *gorm.DB, taskID uint, want string) store.PredictionTask {
	t.Helper()

	deadline := time.Now().Add(taskStatusWaitTimeout)
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

// waitForAllSucceeded 等待指定数量的任务全部成功，并确认每条都有结构结果。
func waitForAllSucceeded(t *testing.T, database *gorm.DB, want int64) {
	t.Helper()

	deadline := time.Now().Add(taskStatusWaitTimeout)
	var succeeded int64
	for time.Now().Before(deadline) {
		if err := database.Model(&store.PredictionTask{}).
			Where("status = ?", store.TaskStatusSucceeded).
			Count(&succeeded).Error; err != nil {
			t.Fatalf("count succeeded tasks: %v", err)
		}
		if succeeded == want {
			var results int64
			if err := database.Model(&store.StructureResult{}).Count(&results).Error; err != nil {
				t.Fatalf("count results: %v", err)
			}
			if results != want {
				t.Fatalf("results = %d, want %d", results, want)
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("only %d/%d tasks succeeded in time", succeeded, want)
}

func TestCreateStoresAPendingTask(t *testing.T) {
	database := openTestDatabase(t)
	service := tasks.New(database, stubPredictor{})

	task, err := service.Create(context.Background(), "NaCl", nil)
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

func TestCreateStoresRequestTraceID(t *testing.T) {
	database := openTestDatabase(t)
	service := tasks.New(database, stubPredictor{})

	ctx := traceid.With(context.Background(), "web-test-123")
	task, err := service.Create(ctx, "NaCl", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if task.TraceID != "web-test-123" {
		t.Errorf("task trace_id = %q, want web-test-123", task.TraceID)
	}

	var stored store.PredictionTask
	if err := database.First(&stored, task.ID).Error; err != nil {
		t.Fatalf("load stored task: %v", err)
	}
	if stored.TraceID != "web-test-123" {
		t.Errorf("stored trace_id = %q, want web-test-123", stored.TraceID)
	}
}

func TestWorkerPassesRequestTraceIDToCompute(t *testing.T) {
	database := openTestDatabase(t)
	received := make(chan string, 1)
	service := startService(t, database, stubPredictor{
		predict: func(ctx context.Context, formula string) (compute.Prediction, error) {
			received <- traceid.From(ctx)
			return samplePrediction(formula), nil
		},
	})

	ctx := traceid.With(context.Background(), "web-test-456")
	task, err := service.Create(ctx, "NaCl", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	waitForStatus(t, database, task.ID, store.TaskStatusSucceeded)

	select {
	case got := <-received:
		if got != "web-test-456" {
			t.Errorf("worker trace_id = %q, want web-test-456", got)
		}
	case <-time.After(time.Second):
		t.Fatal("predictor was not called")
	}
}

func TestServiceExecutesTheTaskAndStoresTheResult(t *testing.T) {
	database := openTestDatabase(t)
	service := startService(t, database, stubPredictor{
		predict: func(_ context.Context, formula string) (compute.Prediction, error) {
			return samplePrediction(formula), nil
		},
	})

	task, err := service.Create(context.Background(), "NaCl", nil)
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

	task, err := service.Create(context.Background(), "NaCl2", nil)
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

	task, err := service.Create(context.Background(), "NaCl", nil)
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

	task, err := service.Create(context.Background(), "NaCl", nil)
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

	task, err := service.Create(context.Background(), "NaCl", nil)
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

// 回归：队列满时任务曾经只能等重启补算；现在补扫会自动重新入队。
func TestQueueFullTaskIsRetriedWithoutRestart(t *testing.T) {
	database := openTestDatabase(t)
	predictorStarted := make(chan struct{})
	releasePredictor := make(chan struct{}, 1)
	var predictorCalls atomic.Int32
	service := startService(t, database, stubPredictor{
		predict: func(_ context.Context, formula string) (compute.Prediction, error) {
			// 第一条任务卡住 worker，后面的任务才有机会把队列填满。
			if predictorCalls.Add(1) == 1 {
				close(predictorStarted)
				<-releasePredictor
			}
			return samplePrediction(formula), nil
		},
	})
	// 测试中途失败时也要放行 worker，否则 Stop 会一直等它。
	t.Cleanup(func() {
		select {
		case releasePredictor <- struct{}{}:
		default:
		}
	})

	first, err := service.Create(context.Background(), "NaCl", nil)
	if err != nil {
		t.Fatalf("Create(first): %v", err)
	}
	select {
	case <-predictorStarted:
	case <-time.After(workerStartTimeout):
		t.Fatalf("worker did not start task %d", first.ID)
	}

	// 正好填满队列，再建一条：这一条必然因为队列满而留在 pending。
	for index := 0; index < tasks.QueueCapacity; index++ {
		if _, err := service.Create(context.Background(), "NaCl", nil); err != nil {
			t.Fatalf("Create(fill %d): %v", index, err)
		}
	}
	overflow, err := service.Create(context.Background(), "NaCl", nil)
	if err != nil {
		t.Fatalf("Create(overflow): %v", err)
	}

	var stored store.PredictionTask
	if err := database.First(&stored, overflow.ID).Error; err != nil {
		t.Fatalf("load overflow task: %v", err)
	}
	if stored.Status != store.TaskStatusPending {
		t.Fatalf("overflow status = %q, want pending while the queue is full", stored.Status)
	}

	// 放行 worker；队列腾出后补扫应重新入队，不需要重启服务。
	releasePredictor <- struct{}{}
	waitForStatus(t, database, overflow.ID, store.TaskStatusSucceeded)
}

// 超时不是“连不上”：任务要落成 compute_timeout，前端才能给出“等太久”的提示。
func TestTimedOutPredictionIsRecordedAsComputeTimeout(t *testing.T) {
	database := openTestDatabase(t)
	release := make(chan struct{})
	stalled := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		// 一直不回话，直到客户端超时；测试结束时由 cleanup 放行 handler。
		select {
		case <-request.Context().Done():
		case <-release:
		}
	}))
	t.Cleanup(stalled.Close)
	t.Cleanup(func() { close(release) })

	service := startService(
		t,
		database,
		compute.NewClientWithTimeout(stalled.URL, 25*time.Millisecond),
	)
	task, err := service.Create(context.Background(), "NaCl", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	failed := waitForStatus(t, database, task.ID, store.TaskStatusFailed)
	if failed.ErrorCode != codes.ComputeTimeout {
		t.Fatalf("error code = %q, want %q", failed.ErrorCode, codes.ComputeTimeout)
	}
	if failed.DurationMS == nil || *failed.DurationMS > 1000 {
		t.Errorf("duration_ms = %v, want a short timeout", failed.DurationMS)
	}
}

// 多人同时提交时，所有任务都要落库并被 worker 处理完，不能丢任务或卡在 pending。
func TestConcurrentCreatesAreAllProcessed(t *testing.T) {
	database := openTestDatabase(t)
	service := startService(t, database, stubPredictor{
		predict: func(_ context.Context, formula string) (compute.Prediction, error) {
			time.Sleep(concurrentComputeDelay)
			return samplePrediction(formula), nil
		},
	})

	taskCount := tasks.QueueCapacity + concurrentOverflowTasks
	createErrors := make(chan error, taskCount)
	var creators sync.WaitGroup
	for index := 0; index < taskCount; index++ {
		creators.Add(1)
		go func() {
			defer creators.Done()
			if _, err := service.Create(context.Background(), "NaCl", nil); err != nil {
				createErrors <- err
			}
		}()
	}
	creators.Wait()
	close(createErrors)

	for err := range createErrors {
		t.Errorf("Create: %v", err)
	}
	if t.Failed() {
		return
	}

	waitForAllSucceeded(t, database, int64(taskCount))
}

// 配置多个 worker 时，两条任务应能同时进入计算；只有 1 个 worker 时这个测试会超时。
func TestConfiguredWorkersRunTasksConcurrently(t *testing.T) {
	database := openTestDatabase(t)

	started := make(chan struct{}, 2)
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseAll := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(releaseAll)

	service := startServiceWithWorkers(t, database, stubPredictor{
		predict: func(_ context.Context, formula string) (compute.Prediction, error) {
			started <- struct{}{}
			<-release
			return samplePrediction(formula), nil
		},
	}, 2)

	for index := 0; index < 2; index++ {
		if _, err := service.Create(context.Background(), "NaCl", nil); err != nil {
			t.Fatalf("Create(%d): %v", index, err)
		}
	}

	for index := 0; index < 2; index++ {
		select {
		case <-started:
		case <-time.After(workerStartTimeout):
			t.Fatalf("only %d/2 workers started a task", index)
		}
	}

	releaseAll()
	waitForAllSucceeded(t, database, 2)
}

func TestNewWithWorkersRejectsInvalidCounts(t *testing.T) {
	for _, workers := range []int{0, -1, tasks.MaxWorkerCount + 1} {
		if _, err := tasks.NewWithWorkers(nil, stubPredictor{}, workers); err == nil {
			t.Errorf("NewWithWorkers(%d) returned no error", workers)
		}
	}
}

func createOwnedTask(t *testing.T, database *gorm.DB, service *tasks.Service, userID uint, formula string) store.PredictionTask {
	t.Helper()

	task, err := service.Create(context.Background(), formula, &userID)
	if err != nil {
		t.Fatalf("Create(%s): %v", formula, err)
	}
	return waitForStatus(t, database, task.ID, store.TaskStatusSucceeded)
}

func TestHistoryListsOnlyOwnedTasksWithResults(t *testing.T) {
	database := openTestDatabase(t)
	firstUser := store.User{DisplayName: "小明"}
	secondUser := store.User{DisplayName: "小红"}
	if err := database.Create(&firstUser).Error; err != nil {
		t.Fatalf("create first user: %v", err)
	}
	if err := database.Create(&secondUser).Error; err != nil {
		t.Fatalf("create second user: %v", err)
	}

	service := startService(t, database, stubPredictor{
		predict: func(_ context.Context, formula string) (compute.Prediction, error) {
			return samplePrediction(formula), nil
		},
	})

	older := createOwnedTask(t, database, service, firstUser.ID, "NaCl")
	newer := createOwnedTask(t, database, service, firstUser.ID, "CaF2")
	other := createOwnedTask(t, database, service, secondUser.ID, "KCl")
	if _, err := service.Create(context.Background(), "MgO", nil); err != nil {
		t.Fatalf("Create anonymous: %v", err)
	}

	page, err := service.History(firstUser.ID, 10, 0)
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if page.Total != 2 {
		t.Errorf("total = %d, want 2 (anonymous tasks must be excluded)", page.Total)
	}
	if len(page.Entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(page.Entries))
	}
	if page.Entries[0].Task.ID != newer.ID || page.Entries[1].Task.ID != older.ID {
		t.Errorf("order = [%d, %d], want newest first [%d, %d]",
			page.Entries[0].Task.ID, page.Entries[1].Task.ID, newer.ID, older.ID)
	}
	for _, entry := range page.Entries {
		if entry.Result == nil || entry.Result.Prototype != "rock_salt" {
			t.Errorf("entry %d has no expected result: %+v", entry.Task.ID, entry.Result)
		}
	}

	otherPage, err := service.History(secondUser.ID, 10, 0)
	if err != nil {
		t.Fatalf("History(second user): %v", err)
	}
	if otherPage.Total != 1 || len(otherPage.Entries) != 1 || otherPage.Entries[0].Task.ID != other.ID {
		t.Errorf("second user history = %+v, want only task %d", otherPage, other.ID)
	}
}

func TestHistoryPaginates(t *testing.T) {
	database := openTestDatabase(t)
	user := store.User{DisplayName: "小明"}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	service := startService(t, database, stubPredictor{
		predict: func(_ context.Context, formula string) (compute.Prediction, error) {
			return samplePrediction(formula), nil
		},
	})
	for _, formula := range []string{"NaCl", "CaF2", "KCl"} {
		createOwnedTask(t, database, service, user.ID, formula)
	}

	firstPage, err := service.History(user.ID, 2, 0)
	if err != nil {
		t.Fatalf("first page: %v", err)
	}
	if firstPage.Total != 3 || firstPage.Limit != 2 || firstPage.Offset != 0 {
		t.Errorf("first page metadata = %+v", firstPage)
	}
	if len(firstPage.Entries) != 2 {
		t.Errorf("first page entries = %d, want 2", len(firstPage.Entries))
	}

	secondPage, err := service.History(user.ID, 2, 2)
	if err != nil {
		t.Fatalf("second page: %v", err)
	}
	if len(secondPage.Entries) != 1 {
		t.Errorf("second page entries = %d, want 1", len(secondPage.Entries))
	}

	clamped, err := service.History(user.ID, 999, 0)
	if err != nil {
		t.Fatalf("clamped page: %v", err)
	}
	if clamped.Limit != tasks.MaxHistoryLimit {
		t.Errorf("limit = %d, want %d", clamped.Limit, tasks.MaxHistoryLimit)
	}
}
