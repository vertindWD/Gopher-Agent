package repository

import (
	"agent/internal/config"
	"agent/pkg/logger"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var RDB *redis.Client

// PublishChunk 同时写入 List（持久化供重连回放）和 Pub/Sub（实时推送）
func PublishChunk(ctx context.Context, taskID, chunk string) error {
	listKey := "task:chunks:" + taskID
	pipe := RDB.Pipeline()
	pipe.RPush(ctx, listKey, chunk)
	pipe.Expire(ctx, listKey, time.Hour)
	pipe.Publish(ctx, "task:stream:"+taskID, chunk)
	_, err := pipe.Exec(ctx)
	return err
}

// GetChunks 获取已缓冲的全部 chunks，供重连时回放
func GetChunks(ctx context.Context, taskID string) ([]string, error) {
	return RDB.LRange(ctx, "task:chunks:"+taskID, 0, -1).Result()
}

func SubscribeStream(ctx context.Context, taskID string) *redis.PubSub {
	return RDB.Subscribe(ctx, "task:stream:"+taskID)
}

func InitRedis() {
	cfg := config.AppConfig.RedisConfig

	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// 测试连接
	if err := RDB.Ping(context.Background()).Err(); err != nil {
		logger.Log.Fatal("❌ Redis 连接失败", zap.Error(err))
	}

	logger.Log.Info("✅ Redis 基础设施初始化完成")
}
