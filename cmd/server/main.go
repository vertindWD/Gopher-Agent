package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"agent/internal/api"
	"agent/internal/config"
	"agent/internal/repository"
	"agent/internal/worker"
	"agent/pkg/kafka"
	"agent/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	logger.InitLogger()
	defer logger.Log.Sync()

	config.InitConfig()

	repository.InitDB(config.AppConfig.MySQLConfig.DSN())
	repository.InitRedis()
	kafka.InitProducer()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		worker.StartConsumer(ctx)
	}()

	r := api.SetupRouter()
	port := config.AppConfig.ServerConfig.Port
	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}

	logger.Log.Info("🚀 Gopher-Agent API Server is running on port " + port)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatal("服务启动失败", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Log.Info("收到退出信号，开始优雅退出...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("HTTP Server 关闭异常", zap.Error(err))
	}

	wg.Wait()
	logger.Log.Info("✅ 服务已安全退出")
}
