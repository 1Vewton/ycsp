// Package tasks 管理预测任务：创建、排队、执行、取消与查询。
//
// 任务先以 pending 落库再异步计算，接口立刻返回，前端轮询状态（见 backend/README.md）。
package tasks

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/1Vewton/ycsp/backend/internal/compute"
	"github.com/1Vewton/ycsp/backend/internal/store"
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
	// queueCapacity 是排队缓冲上限；满了任务仍留在 pending，服务重启时会重新排队。
	queueCapacity = 64
	// workerCount 是同时计算的任务数：SQLite 只允许一个写入者，教学场景 1 个足够。
	workerCount = 1
)

// Service 负责任务的落库、排队与执行。
type Service struct {
	database  *gorm.DB
	predictor Predictor
	queue     chan uint
	workers   sync.WaitGroup
}

// New 创建任务服务；调用 Start 之前任务只会排队，不会执行。
func New(database *gorm.DB, predictor Predictor) *Service {
	return &Service{
		database:  database,
		predictor: predictor,
		queue:     make(chan uint, queueCapacity),
	}
}

// Start 先把上次没算完的任务重新排队，再启动后台 worker。
// ctx 取消后 worker 退出，配合 Stop 等待它们收尾。
func (service *Service) Start(ctx context.Context) {
	service.requeueUnfinished()

	for workerIndex := 0; workerIndex < workerCount; workerIndex++ {
		service.workers.Add(1)
		go service.work(ctx)
	}
}

// Stop 等后台 worker 退出（需先取消传给 Start 的 ctx）。
func (service *Service) Stop() {
	service.workers.Wait()
}

// Create 建一条 pending 任务并排队，立即返回；userID 为空表示匿名预测。
func (service *Service) Create(formula string, userID *uint) (*store.PredictionTask, error) {
	task := store.PredictionTask{
		Formula: formula,
		UserID:  userID,
		Status:  store.TaskStatusPending,
	}
	if err := service.database.Create(&task).Error; err != nil {
		return nil, err
	}
	service.enqueue(task.ID)
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

// enqueue 把任务放进队列；队列满时只记日志，任务仍是 pending（重启会重新排队）。
func (service *Service) enqueue(taskID uint) {
	select {
	case service.queue <- taskID:
	default:
		log.Printf("tasks: queue is full, task %d stays pending", taskID)
	}
}

// requeueUnfinished 处理上次进程退出时的残留：
// running 说明算到一半进程挂了，退回 pending 重算；pending 说明还没开始。
func (service *Service) requeueUnfinished() {
	if err := service.database.Model(&store.PredictionTask{}).
		Where("status = ?", store.TaskStatusRunning).
		Update("status", store.TaskStatusPending).Error; err != nil {
		log.Printf("tasks: reset running tasks: %v", err)
	}

	var taskIDs []uint
	if err := service.database.Model(&store.PredictionTask{}).
		Where("status = ?", store.TaskStatusPending).
		Order("created_at").
		Pluck("id", &taskIDs).Error; err != nil {
		log.Printf("tasks: load pending tasks: %v", err)
		return
	}
	for _, taskID := range taskIDs {
		service.enqueue(taskID)
	}
}
