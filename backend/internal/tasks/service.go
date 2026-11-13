// Package tasks 管理预测任务：创建、排队、执行、取消与查询。
//
// 任务先以 pending 落库再异步计算，接口立刻返回，前端轮询状态（见 backend/README.md）。
package tasks

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/1Vewton/ycsp/backend/internal/compute"
	"github.com/1Vewton/ycsp/backend/internal/store"
	"github.com/1Vewton/ycsp/backend/internal/traceid"
)

// Predictor 是“把化学式算成结构”的能力：生产用 compute.Client，测试用假实现。
type Predictor interface {
	Predict(ctx context.Context, formula string) (compute.Prediction, error)
}

// ErrTaskNotFound 表示任务编号不存在。
var ErrTaskNotFound = errors.New("tasks: task not found")

// NotReadyError 表示任务还没有结果：还在排队、计算中，或已取消。
type NotReadyError struct {
	Status string
}

func (err *NotReadyError) Error() string {
	return "tasks: no result while the task is " + err.Status
}

// CancelConflictError 表示当前状态不允许取消：只有排队中的任务能取消。
type CancelConflictError struct {
	Status string
}

func (err *CancelConflictError) Error() string {
	return "tasks: cannot cancel a task that is " + err.Status
}

const (
	// QueueCapacity 是内存队列的缓冲上限；满了任务仍留在 pending，
	// 由 sweepPending 重新入队，所以不需要把内存队列做成无界的。
	QueueCapacity = 64
	// DefaultWorkerCount 是未显式配置时的 worker 数：
	// SQLite 只允许一个写入者，本地开发和测试用一个 worker 最稳。
	DefaultWorkerCount = 1
	// MaxWorkerCount 防止误配出成百上千个 worker 把数据库连接池打满。
	MaxWorkerCount = 16
	// pendingSweepInterval 是补扫 pending 任务的间隔：队列满时被丢掉的任务
	// 最迟在一个间隔后重新入队，与前端 1 秒轮询的节奏一致。
	pendingSweepInterval = time.Second
	// pendingSweepBatchSize 限制一次补扫读取的任务数，积压时也不会每秒拉全表。
	pendingSweepBatchSize = QueueCapacity

	// DefaultHistoryLimit / MaxHistoryLimit 是“我的预测记录”每页条数的默认值与上限。
	DefaultHistoryLimit = 20
	MaxHistoryLimit     = 50
)

// HistoryEntry 是一条预测记录：任务本身，以及成功时对应的结构结果。
type HistoryEntry struct {
	Task   store.PredictionTask
	Result *store.StructureResult
}

// HistoryPage 是一页预测记录；Total 是当前账号的全部任务数。
type HistoryPage struct {
	Entries []HistoryEntry
	Total   int64
	Limit   int
	Offset  int
}

// Service 负责任务的落库、排队与执行。
type Service struct {
	database  *gorm.DB
	predictor Predictor
	queue     chan uint
	workers   sync.WaitGroup
	// workerCount 是同时执行任务的 goroutine 数；由 main 按数据库类型配置。
	workerCount int

	// scheduled 记录已经交给队列、还没执行完的任务编号。
	// 补扫会反复读取 pending 任务，靠它避免同一个任务在队列里排队多次。
	scheduled   map[uint]struct{}
	scheduledMu sync.Mutex
}

// New 创建任务服务，使用默认 worker 数；调用 Start 之前任务只会排队，不会执行。
func New(database *gorm.DB, predictor Predictor) *Service {
	return newService(database, predictor, DefaultWorkerCount)
}

// NewWithWorkers 创建指定 worker 数的任务服务。
// SQLite 建议 1 个；Postgres / MySQL 可以按吞吐需要调大。
func NewWithWorkers(database *gorm.DB, predictor Predictor, workers int) (*Service, error) {
	if workers < 1 || workers > MaxWorkerCount {
		return nil, fmt.Errorf("tasks: worker count must be between 1 and %d", MaxWorkerCount)
	}
	return newService(database, predictor, workers), nil
}

func newService(database *gorm.DB, predictor Predictor, workers int) *Service {
	return &Service{
		database:    database,
		predictor:   predictor,
		queue:       make(chan uint, QueueCapacity),
		workerCount: workers,
		scheduled:   make(map[uint]struct{}, QueueCapacity),
	}
}

// Start 先把上次没算完的任务重新排队，再启动补扫与后台 worker。
// ctx 取消后它们退出，配合 Stop 等待收尾。
func (service *Service) Start(ctx context.Context) {
	service.requeueUnfinished()

	service.workers.Add(1)
	go service.sweepPending(ctx)

	for workerIndex := 0; workerIndex < service.workerCount; workerIndex++ {
		service.workers.Add(1)
		go service.work(ctx)
	}
}

// Stop 等补扫与 worker 退出（需先取消传给 Start 的 ctx）。
func (service *Service) Stop() {
	service.workers.Wait()
}

// Create 建一条 pending 任务并排队，立即返回；userID 为空表示匿名预测。
// ctx 里带着这次请求的 trace_id，worker 调用计算层时继续使用。
func (service *Service) Create(
	ctx context.Context, formula string, userID *uint,
) (*store.PredictionTask, error) {
	task := store.PredictionTask{
		Formula: formula,
		UserID:  userID,
		Status:  store.TaskStatusPending,
		TraceID: traceid.From(ctx),
	}
	if err := service.database.Create(&task).Error; err != nil {
		return nil, err
	}
	if !service.enqueue(task.ID) {
		log.Printf("tasks: queue is full, task %d stays pending for the sweep", task.ID)
	}
	return &task, nil
}

// Get 按编号取任务；不存在时返回 ErrTaskNotFound。
func (service *Service) Get(taskID uint) (*store.PredictionTask, error) {
	var task store.PredictionTask
	if err := service.database.First(&task, taskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

// Result 取任务的结构结果；还没算完时返回 NotReadyError。
func (service *Service) Result(taskID uint) (*store.StructureResult, error) {
	task, err := service.Get(taskID)
	if err != nil {
		return nil, err
	}
	if task.Status != store.TaskStatusSucceeded {
		return nil, &NotReadyError{Status: task.Status}
	}

	var result store.StructureResult
	if err := service.database.Where("task_id = ?", taskID).First(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 理论上不会发生：结果与 succeeded 在同一事务里写。兜住即可。
			return nil, &NotReadyError{Status: task.Status}
		}
		return nil, err
	}
	return &result, nil
}

// History 按账号分页取预测记录，最新的在前；匿名任务（user_id 为空）不会出现在这里。
func (service *Service) History(userID uint, limit int, offset int) (*HistoryPage, error) {
	if limit <= 0 {
		limit = DefaultHistoryLimit
	}
	if limit > MaxHistoryLimit {
		limit = MaxHistoryLimit
	}
	if offset < 0 {
		offset = 0
	}

	var total int64
	if err := service.database.Model(&store.PredictionTask{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, err
	}

	var tasks []store.PredictionTask
	if err := service.database.
		Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	resultsByTask := make(map[uint]*store.StructureResult, len(tasks))
	if len(tasks) > 0 {
		taskIDs := make([]uint, 0, len(tasks))
		for _, task := range tasks {
			taskIDs = append(taskIDs, task.ID)
		}

		var results []store.StructureResult
		if err := service.database.
			Where("task_id IN ?", taskIDs).
			Find(&results).Error; err != nil {
			return nil, err
		}
		for index := range results {
			result := results[index]
			resultsByTask[result.TaskID] = &result
		}
	}

	entries := make([]HistoryEntry, 0, len(tasks))
	for _, task := range tasks {
		entries = append(entries, HistoryEntry{
			Task:   task,
			Result: resultsByTask[task.ID],
		})
	}
	return &HistoryPage{
		Entries: entries,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

// Cancel 只允许取消排队中的任务；已经开始或结束的任务返回 CancelConflictError。
func (service *Service) Cancel(taskID uint) (*store.PredictionTask, error) {
	update := service.database.Model(&store.PredictionTask{}).
		Where("id = ? AND status = ?", taskID, store.TaskStatusPending).
		Updates(map[string]any{
			"status":      store.TaskStatusCancelled,
			"finished_at": time.Now().UTC(),
		})
	if update.Error != nil {
		return nil, update.Error
	}
	if update.RowsAffected == 0 {
		task, err := service.Get(taskID)
		if err != nil {
			return nil, err
		}
		return nil, &CancelConflictError{Status: task.Status}
	}
	return service.Get(taskID)
}

// enqueue 把任务放进队列；同一任务只会放一次。
// 队列满时返回 false，任务仍是 pending，由 sweepPending 稍后重试。
func (service *Service) enqueue(taskID uint) bool {
	if !service.markScheduled(taskID) {
		return true // 已经在队列或执行中
	}
	select {
	case service.queue <- taskID:
		return true
	default:
		service.forgetScheduled(taskID)
		return false
	}
}

// markScheduled 记录任务已经交给队列；重复调用返回 false。
func (service *Service) markScheduled(taskID uint) bool {
	service.scheduledMu.Lock()
	defer service.scheduledMu.Unlock()
	if _, exists := service.scheduled[taskID]; exists {
		return false
	}
	service.scheduled[taskID] = struct{}{}
	return true
}

// forgetScheduled 在任务执行完（或入队失败）后清掉记录。
func (service *Service) forgetScheduled(taskID uint) {
	service.scheduledMu.Lock()
	delete(service.scheduled, taskID)
	service.scheduledMu.Unlock()
}

// sweepPending 周期性补扫 pending 任务：
// 队列满时 enqueue 会放弃，只靠创建时那一次入队会让任务永远停在 pending。
func (service *Service) sweepPending(ctx context.Context) {
	defer service.workers.Done()
	ticker := time.NewTicker(pendingSweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			service.enqueuePendingTasks()
		}
	}
}

// enqueuePendingTasks 读取最早的一批 pending 任务补入队列；
// 队列仍满时剩下的任务留给下一轮，不在这里死等。
func (service *Service) enqueuePendingTasks() {
	// 队列满时补扫也放不进去，等 worker 腾出位置再查数据库。
	if len(service.queue) == cap(service.queue) {
		return
	}

	taskIDs, err := service.pendingTaskIDs()
	if err != nil {
		log.Printf("tasks: load pending tasks for the sweep: %v", err)
		return
	}
	waiting := 0
	for _, taskID := range taskIDs {
		if !service.enqueue(taskID) {
			waiting++
		}
	}
	if waiting > 0 {
		log.Printf("tasks: queue is full, %d pending task(s) wait for the next sweep", waiting)
	}
}

// pendingTaskIDs 按创建顺序取一批 pending 任务编号。
func (service *Service) pendingTaskIDs() ([]uint, error) {
	var taskIDs []uint
	err := service.database.Model(&store.PredictionTask{}).
		Where("status = ?", store.TaskStatusPending).
		Order("created_at, id").
		Limit(pendingSweepBatchSize).
		Pluck("id", &taskIDs).Error
	return taskIDs, err
}

// requeueUnfinished 处理上次进程退出时的残留：
// running 说明算到一半进程挂了，退回 pending 重算；pending 说明还没开始。
func (service *Service) requeueUnfinished() {
	if err := service.database.Model(&store.PredictionTask{}).
		Where("status = ?", store.TaskStatusRunning).
		Update("status", store.TaskStatusPending).Error; err != nil {
		log.Printf("tasks: reset running tasks: %v", err)
	}

	taskIDs, err := service.pendingTaskIDs()
	if err != nil {
		log.Printf("tasks: load pending tasks: %v", err)
		return
	}
	waiting := 0
	for _, taskID := range taskIDs {
		if !service.enqueue(taskID) {
			waiting++
		}
	}
	if waiting > 0 {
		log.Printf("tasks: queue is full at startup, %d pending task(s) wait for the sweep", waiting)
	}
}
