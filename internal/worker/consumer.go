package worker

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"agent/internal/config"
	"agent/internal/model"
	"agent/internal/repository"
	"agent/internal/service"
	"agent/pkg/logger"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// StartConsumer 启动后台监听，ctx 取消时等待所有进行中的任务完成再退出
func StartConsumer(ctx context.Context) {
	cfg := config.AppConfig.KafkaConfig
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  "agent-worker-group",
		MaxBytes: 10e6,
	})
	defer reader.Close()

	logger.Log.Info("🎧 Worker 节点已启动，正在监听 Kafka 队列...")

	var wg sync.WaitGroup
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				logger.Log.Info("Worker 停止接收新任务，等待进行中任务完成...")
				wg.Wait()
				logger.Log.Info("✅ Worker 已安全退出")
				return
			}
			logger.Log.Error("读取 Kafka 消息失败", zap.Error(err))
			continue
		}

		var taskData map[string]string
		if err := json.Unmarshal(msg.Value, &taskData); err != nil {
			logger.Log.Error("解析消息体失败", zap.Error(err))
			continue
		}

		wg.Add(1)
		go func(taskID, prompt string) {
			defer wg.Done()
			processTask(taskID, prompt)
		}(taskData["task_id"], taskData["prompt"])
	}
}

func processTask(taskID, prompt string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	logger.Log.Info("⚙️ 开始处理任务", zap.String("task_id", taskID))

	// 1. 更新为运行中
	repository.DB.Model(&model.AgentTask{}).Where("task_id = ?", taskID).Update("status", model.TaskStatusRunning)

	// 2. 调用 Agent 引擎
	result, err := service.RunAgent(ctx, prompt, func(chunk string) {
		repository.PublishChunk(context.Background(), taskID, chunk)
	})

	// 无论成功失败，通知订阅方任务已结束
	repository.PublishChunk(context.Background(), taskID, "[DONE]")

	// 3. 更新最终状态
	if err != nil {
		repository.DB.Model(&model.AgentTask{}).Where("task_id = ?", taskID).Updates(map[string]interface{}{
			"status":    model.TaskStatusFailed,
			"error_msg": err.Error(),
		})
		logger.Log.Error("❌ 任务执行失败", zap.String("task_id", taskID), zap.Error(err))
		return
	}

	repository.DB.Model(&model.AgentTask{}).Where("task_id = ?", taskID).Updates(map[string]interface{}{
		"status": model.TaskStatusCompleted,
		"result": result,
	})
	logger.Log.Info("✅ 任务执行完成", zap.String("task_id", taskID))
}
