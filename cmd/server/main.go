package main

import (
	"agent/internal/api"
	"agent/internal/config"
	"agent/internal/repository"
	"agent/internal/worker" // 【新增】引入 worker 包
	"agent/pkg/kafka"
	"agent/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	// 1. 启动日志引擎
	logger.InitLogger()
	defer logger.Log.Sync()

	// 2. 加载配置文件
	config.InitConfig()

	// 3. 初始化基础设施
	repository.InitDB(config.AppConfig.MySQLConfig.DSN())
	repository.InitRedis()
	kafka.InitProducer()

	// 【新增】启动 Kafka 消费者 (必须放在 go 协程中运行)
	go worker.StartConsumer()

	// 4. 启动 HTTP API 服务
	r := api.SetupRouter()
	port := config.AppConfig.ServerConfig.Port
	logger.Log.Info("🚀 Gopher-Agent API Server is running on port " + port)

	if err := r.Run(port); err != nil {
		logger.Log.Fatal("服务启动失败: ", zap.Error(err))
	}
}
