package worker

import (
	"context"
	"encoding/json"
	"time"

	"agent/internal/config"
	"agent/internal/model"
	"agent/internal/repository"
	"agent/internal/service"
	"agent/pkg/logger"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// StartConsumer 启动后台监听
func StartConsumer() {
	cfg := config.AppConfig.KafkaConfig
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  "agent-worker-group", // 消费者组
		MaxBytes: 10e6,                 // 10MB
	})

	logger.Log.Info("🎧 Worker 节点已启动，正在监听 Kafka 队列...")

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			logger.Log.Error("读取 Kafka 消息失败", zap.Error(err))
			continue
		}

		var taskData map[string]string
		if err := json.Unmarshal(msg.Value, &taskData); err != nil {
			logger.Log.Error("解析消息体失败", zap.Error(err))
			continue
		}

		taskID := taskData["task_id"]
		prompt := taskData["prompt"]

		// 开启独立的 Goroutine 并发处理任务
		go processTask(taskID, prompt)
	}
}

func processTask(taskID, prompt string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	logger.Log.Info("⚙️ 开始处理任务", zap.String("task_id", taskID))

	// 1. 更新为运行中
	repository.DB.Model(&model.AgentTask{}).Where("task_id = ?", taskID).Update("status", model.TaskStatusRunning)

	// 2. 调用 Agent 引擎
	result, err := service.RunAgent(ctx, prompt)

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
