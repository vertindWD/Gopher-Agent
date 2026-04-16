package repository

import (
	"agent/internal/config"
	"agent/pkg/logger"
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var RDB *redis.Client

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
