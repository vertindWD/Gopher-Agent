package api

import (
	"context"
	"net/http"

	"agent/internal/model"
	"agent/internal/repository"
	"agent/pkg/kafka"
	"agent/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TaskRequest 定义前端传来的 JSON 格式
type TaskRequest struct {
	Prompt string `json:"prompt" binding:"required"`
}

// SubmitTask 接收用户指令并进入异步队列
func SubmitTask(c *gin.Context) {
	var req TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误，需要提供 prompt"})
		return
	}

	// 1. 生成全局唯一任务ID
	taskID := uuid.New().String()

	// 2. 第一时间落库，状态设为 Pending (待处理)
	task := model.AgentTask{
		TaskID: taskID,
		Prompt: req.Prompt,
		Status: model.TaskStatusPending,
	}

	if err := repository.DB.Create(&task).Error; err != nil {
		logger.Log.Error("任务落库失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统内部错误"})
		return
	}

	// 3. 将任务投递到 Kafka
	if err := kafka.SendTaskMessage(context.Background(), taskID, req.Prompt); err != nil {
		// 投递失败时的处理
		repository.DB.Model(&task).Update("status", model.TaskStatusFailed)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统繁忙，任务排队失败"})
		return
	}

	// 4. 投递成功，更新数据库状态为 InQueue (已入队)
	repository.DB.Model(&task).Update("status", model.TaskStatusRunning)

	// 5. 立即返回 TaskID 给前端，不让前端等 AI 思考
	c.JSON(http.StatusOK, gin.H{
		"message": "任务已成功提交至后台队列",
		"task_id": taskID,
	})
}

// QueryTask 查询任务执行状态与结果
func QueryTask(c *gin.Context) {
	taskID := c.Param("task_id")

	var task model.AgentTask
	// 去数据库里找这个任务
	if err := repository.DB.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到该任务"})
		return
	}

	// 组装返回给前端的数据结构
	response := gin.H{
		"task_id": task.TaskID,
		"status":  task.Status,
	}

	// 如果任务完成了，就把 AI 的思考结果带上
	if task.Status == model.TaskStatusCompleted {
		response["result"] = task.Result
	} else if task.Status == model.TaskStatusFailed {
		response["error_msg"] = task.ErrorMsg
	}

	c.JSON(http.StatusOK, response)
}
