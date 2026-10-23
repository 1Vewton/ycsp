package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"github.com/1Vewton/ycsp/backend/internal/codes"
	"github.com/1Vewton/ycsp/backend/internal/store"
	"github.com/1Vewton/ycsp/backend/internal/tasks"
)

// maxFormulaLength 与数据库列宽一致：化学式最长 64 个字符。
const maxFormulaLength = 64

type createTaskRequest struct {
	Formula string `json:"formula"`
}

// taskPayload 是任务状态的对外形状；错误信息只在失败时出现。
type taskPayload struct {
	ID         uint              `json:"id"`
	Formula    string            `json:"formula"`
	Status     string            `json:"status"`
	CreatedAt  time.Time         `json:"created_at"`
	StartedAt  *time.Time        `json:"started_at"`
	FinishedAt *time.Time        `json:"finished_at"`
	DurationMS *int64            `json:"duration_ms"`
	Error      *taskErrorPayload `json:"error,omitempty"`
}

type taskErrorPayload struct {
	Code   string         `json:"code"`
	Params map[string]any `json:"params,omitempty"`
}

// taskResultPayload 沿用数据处理层的字段名，前端拿到的结构与 POST /predict 一致。
type taskResultPayload struct {
	Family             string          `json:"family"`
	Prototype          string          `json:"prototype"`
	CoordinationNumber int             `json:"coordination_number"`
	Cell               taskResultCell  `json:"cell"`
	Atoms              json.RawMessage `json:"atoms"`
	Rule               json.RawMessage `json:"rule"`
	Caveats            json.RawMessage `json:"caveats"`
}

type taskResultCell struct {
	EdgeAAngstrom    float64 `json:"edge_a_angstrom"`
	EdgeBAngstrom    float64 `json:"edge_b_angstrom"`
	EdgeCAngstrom    float64 `json:"edge_c_angstrom"`
	AngleAlphaDegree float64 `json:"angle_alpha_degree"`
	AngleBetaDegree  float64 `json:"angle_beta_degree"`
	AngleGammaDegree float64 `json:"angle_gamma_degree"`
}

func registerTaskRoutes(router *gin.Engine, service *tasks.Service) {
	router.POST("/api/tasks", createTaskHandler(service))
	router.GET("/api/tasks/:id", getTaskHandler(service))
	router.GET("/api/tasks/:id/result", getTaskResultHandler(service))
	router.POST("/api/tasks/:id/cancel", cancelTaskHandler(service))
}

func createTaskHandler(service *tasks.Service) gin.HandlerFunc {
	return func(context *gin.Context) {
		var request createTaskRequest
		if err := context.ShouldBindJSON(&request); err != nil {
			writeError(context, http.StatusBadRequest, codes.RequestInvalidBody, nil)
			return
		}

		formula := strings.TrimSpace(request.Formula)
		if formula == "" {
			writeError(context, http.StatusBadRequest, codes.TaskFormulaEmpty, nil)
			return
		}
		if utf8.RuneCountInString(formula) > maxFormulaLength {
			writeError(context, http.StatusBadRequest, codes.TaskFormulaTooLong, gin.H{
				"max_length": maxFormulaLength,
			})
			return
		}

		task, err := service.Create(formula)
		if err != nil {
			writeServerError(context, err)
			return
		}
		writeTask(context, http.StatusCreated, codes.TaskCreated, task)
	}
}

func getTaskHandler(service *tasks.Service) gin.HandlerFunc {
	return func(context *gin.Context) {
		taskID, ok := parseTaskID(context)
		if !ok {
			return
		}

		task, err := service.Get(taskID)
		if err != nil {
			writeTaskError(context, err, context.Param("id"))
			return
		}
		writeTask(context, http.StatusOK, codes.TaskOK, task)
	}
}

func getTaskResultHandler(service *tasks.Service) gin.HandlerFunc {
	return func(context *gin.Context) {
		taskID, ok := parseTaskID(context)
		if !ok {
			return
		}

		result, err := service.Result(taskID)
		if err != nil {
			writeTaskError(context, err, context.Param("id"))
			return
		}
		context.JSON(http.StatusOK, gin.H{
			"service": ServiceName,
			"status":  "ok",
			"code":    codes.TaskResult,
			"result":  resultToPayload(result),
		})
	}
}

func cancelTaskHandler(service *tasks.Service) gin.HandlerFunc {
	return func(context *gin.Context) {
		taskID, ok := parseTaskID(context)
		if !ok {
			return
		}

		task, err := service.Cancel(taskID)
		if err != nil {
			writeTaskError(context, err, context.Param("id"))
			return
		}
		writeTask(context, http.StatusOK, codes.TaskCancelled, task)
	}
}

func parseTaskID(context *gin.Context) (uint, bool) {
	rawTaskID := context.Param("id")
	taskID, err := strconv.ParseUint(rawTaskID, 10, 64)
	if err != nil || taskID == 0 {
		writeError(context, http.StatusNotFound, codes.TaskNotFound, gin.H{"task_id": rawTaskID})
		return 0, false
	}
	return uint(taskID), true
}

func writeTask(context *gin.Context, statusCode int, code string, task *store.PredictionTask) {
	context.JSON(statusCode, gin.H{
		"service": ServiceName,
		"status":  "ok",
		"code":    code,
		"task":    taskToPayload(task),
	})
}

func writeTaskError(context *gin.Context, err error, rawTaskID string) {
	var notReady *tasks.NotReadyError
	var cancelConflict *tasks.CancelConflictError

	switch {
	case errors.Is(err, tasks.ErrTaskNotFound):
		writeError(context, http.StatusNotFound, codes.TaskNotFound, gin.H{"task_id": rawTaskID})
	case errors.As(err, &notReady):
		writeError(context, http.StatusConflict, codes.TaskNotReady, gin.H{"status": notReady.Status})
	case errors.As(err, &cancelConflict):
		writeError(context, http.StatusConflict, codes.TaskCancelConflict, gin.H{
			"status": cancelConflict.Status,
		})
	default:
		writeServerError(context, err)
	}
}

func taskToPayload(task *store.PredictionTask) taskPayload {
	payload := taskPayload{
		ID:         task.ID,
		Formula:    task.Formula,
		Status:     task.Status,
		CreatedAt:  task.CreatedAt.UTC(),
		StartedAt:  utcPointer(task.StartedAt),
		FinishedAt: utcPointer(task.FinishedAt),
		DurationMS: task.DurationMS,
	}
	if task.ErrorCode != "" {
		payload.Error = &taskErrorPayload{
			Code:   task.ErrorCode,
			Params: decodeErrorParams(task.ErrorParams),
		}
	}
	return payload
}

func resultToPayload(result *store.StructureResult) taskResultPayload {
	return taskResultPayload{
		Family:             result.Family,
		Prototype:          result.Prototype,
		CoordinationNumber: result.CoordinationNumber,
		Cell: taskResultCell{
			EdgeAAngstrom:    result.EdgeAAngstrom,
			EdgeBAngstrom:    result.EdgeBAngstrom,
			EdgeCAngstrom:    result.EdgeCAngstrom,
			AngleAlphaDegree: result.AngleAlphaDegree,
			AngleBetaDegree:  result.AngleBetaDegree,
			AngleGammaDegree: result.AngleGammaDegree,
		},
		Atoms:   json.RawMessage(result.Atoms),
		Rule:    json.RawMessage(result.Rule),
		Caveats: json.RawMessage(result.Caveats),
	}
}

func utcPointer(moment *time.Time) *time.Time {
	if moment == nil {
		return nil
	}
	utc := moment.UTC()
	return &utc
}

func decodeErrorParams(rawParams string) map[string]any {
	if rawParams == "" {
		return nil
	}
	params := map[string]any{}
	if err := json.Unmarshal([]byte(rawParams), &params); err != nil {
		return nil
	}
	return params
}
