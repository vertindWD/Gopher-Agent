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

	// 4. 立即返回 TaskID 给前端，不让前端傻等 AI 思考
	c.JSON(http.StatusOK, gin.H{
		"message": "任务已成功提交至后台队列",
		"task_id": taskID,
	})
}

// StreamTask 通过 SSE 实时推送 Agent 的流式输出
func StreamTask(c *gin.Context) {
	taskID := c.Param("task_id")

	var task model.AgentTask
	if err := repository.DB.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到该任务"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 任务已完成，直接返回完整结果，无需订阅
	if task.Status == model.TaskStatusCompleted {
		c.SSEvent("result", task.Result)
		c.Writer.Flush()
		return
	}

	ctx := c.Request.Context()

	// 先订阅，再 replay 缓冲，保证不丢实时消息
	pubsub := repository.SubscribeStream(ctx, taskID)
	defer pubsub.Close()

	// 回放断线前已缓冲的 chunks
	buffered, _ := repository.GetChunks(ctx, taskID)
	for _, chunk := range buffered {
		if chunk == "[DONE]" {
			c.SSEvent("done", "")
			c.Writer.Flush()
			return
		}
		c.SSEvent("chunk", chunk)
	}
	c.Writer.Flush()

	// 继续接收实时 chunks
	for {
		select {
		case msg := <-pubsub.Channel():
			if msg.Payload == "[DONE]" {
				c.SSEvent("done", "")
				c.Writer.Flush()
				return
			}
			c.SSEvent("chunk", msg.Payload)
			c.Writer.Flush()
		case <-ctx.Done():
			return
		}
	}
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
