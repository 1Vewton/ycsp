package tasks

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/1Vewton/ycsp/backend/internal/codes"
	"github.com/1Vewton/ycsp/backend/internal/compute"
	"github.com/1Vewton/ycsp/backend/internal/store"
)

// work 从队列取任务并执行，直到 ctx 取消。
func (service *Service) work(ctx context.Context) {
	defer service.workers.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case taskID := <-service.queue:
			service.runTask(ctx, taskID)
		}
	}
}

// runTask 执行一条任务：抢占 → 调数据处理层 → 写结果或错误。
func (service *Service) runTask(ctx context.Context, taskID uint) {
	task, err := service.claim(taskID)
	if err != nil {
		log.Printf("tasks: claim %d: %v", taskID, err)
		return
	}
	if task == nil {
		return // 已被取消或删除
	}

	startedAt := time.Now()
	if task.StartedAt != nil {
		startedAt = *task.StartedAt
	}

	prediction, predictionErr := service.predictor.Predict(ctx, task.Formula)
	finishedAt := time.Now().UTC()
	durationMilliseconds := finishedAt.Sub(startedAt).Milliseconds()

	if predictionErr != nil {
		code, params := compute.Describe(predictionErr)
		service.fail(taskID, code, params, finishedAt, durationMilliseconds)
		return
	}

	if err := service.succeed(taskID, prediction, finishedAt, durationMilliseconds); err != nil {
		log.Printf("tasks: save result of %d: %v", taskID, err)
		service.fail(taskID, codes.InternalError, nil, finishedAt, durationMilliseconds)
	}
}

// claim 把 pending 改成 running；返回 nil 表示任务已被取消或不存在。
// 用条件更新而不是先读后写：取消与执行并发时，才不会把已取消的任务又跑起来。
func (service *Service) claim(taskID uint) (*store.PredictionTask, error) {
	update := service.database.Model(&store.PredictionTask{}).
		Where("id = ? AND status = ?", taskID, store.TaskStatusPending).
		Updates(map[string]any{
			"status":     store.TaskStatusRunning,
			"started_at": time.Now().UTC(),
		})
	if update.Error != nil {
		return nil, update.Error
	}
	if update.RowsAffected == 0 {
		return nil, nil
	}
	return service.Get(taskID)
}

// succeed 在同一个事务里写结果并把任务标成 succeeded，避免“有结果但状态没改”。
func (service *Service) succeed(
	taskID uint, prediction compute.Prediction, finishedAt time.Time, durationMilliseconds int64,
) error {
	candidate := prediction.Candidates[0]
	result := store.StructureResult{
		TaskID:             taskID,
		Family:             prediction.Family,
		Prototype:          candidate.Prototype,
		CoordinationNumber: candidate.CoordinationNumber,
		EdgeAAngstrom:      candidate.Cell.EdgeAAngstrom,
		EdgeBAngstrom:      candidate.Cell.EdgeBAngstrom,
		EdgeCAngstrom:      candidate.Cell.EdgeCAngstrom,
		AngleAlphaDegree:   candidate.Cell.AngleAlphaDegree,
		AngleBetaDegree:    candidate.Cell.AngleBetaDegree,
		AngleGammaDegree:   candidate.Cell.AngleGammaDegree,
		Atoms:              string(candidate.Atoms),
		Rule:               string(candidate.Rule),
		Caveats:            string(candidate.Caveats),
	}

	return service.database.Transaction(func(transaction *gorm.DB) error {
		if err := transaction.Create(&result).Error; err != nil {
			return err
		}
		return transaction.Model(&store.PredictionTask{}).
			Where("id = ?", taskID).
			Updates(map[string]any{
				"status":      store.TaskStatusSucceeded,
				"finished_at": finishedAt,
				"duration_ms": durationMilliseconds,
			}).Error
	})
}

// fail 记录失败原因：code 与 params 是语言无关的，文案由前端翻译。
func (service *Service) fail(
	taskID uint, code string, params map[string]any, finishedAt time.Time, durationMilliseconds int64,
) {
	encodedParams := []byte("{}")
	if len(params) > 0 {
		encoded, err := json.Marshal(params)
		if err != nil {
			log.Printf("tasks: encode error params of %d: %v", taskID, err)
		} else {
			encodedParams = encoded
		}
	}

	update := service.database.Model(&store.PredictionTask{}).
		Where("id = ?", taskID).
		Updates(map[string]any{
			"status":       store.TaskStatusFailed,
			"error_code":   code,
			"error_params": string(encodedParams),
			"finished_at":  finishedAt,
			"duration_ms":  durationMilliseconds,
		})
	if update.Error != nil {
		log.Printf("tasks: mark %d failed: %v", taskID, update.Error)
	}
}
