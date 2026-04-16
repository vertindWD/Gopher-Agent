package kafka

import (
	"agent/internal/config"
	"agent/pkg/logger"
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var Writer *kafka.Writer

// InitProducer 初始化 Kafka 生产者
func InitProducer() {
	cfg := config.AppConfig.KafkaConfig
	Writer = &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.Topic,
		Balancer: &kafka.LeastBytes{}, // 负载均衡策略：优先发给数据量最少的分区
	}
	logger.Log.Info("✅ Kafka Producer 基础设施初始化完成")
}

// SendTaskMessage 投递任务到消息队列
func SendTaskMessage(ctx context.Context, taskID string, prompt string) error {
	// 构造消息体
	msg := map[string]string{
		"task_id": taskID,
		"prompt":  prompt,
	}
	msgBytes, _ := json.Marshal(msg)

	// 发送消息
	err := Writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(taskID), // 工业细节：用 taskID 作 Key，保证同一任务的后续消息能落到同一个分区（保证顺序性）
		Value: msgBytes,
	})

	if err != nil {
		logger.Log.Error("❌ 投递 Kafka 消息失败", zap.Error(err), zap.String("task_id", taskID))
		return err
	}
	return nil
}
