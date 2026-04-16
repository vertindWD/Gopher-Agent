package repository

import (
	"agent/internal/model"
	"agent/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(dsn string) {
	var err error

	// 配置 GORM 日志级别，生产环境通常只记录 Warn/Error
	config := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	}

	DB, err = gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		logger.Log.Fatal("Failed to connect to database", zap.Error(err))
	}

	// 自动迁移表结构 (自动在 MySQL 中创建 agent_tasks 表)
	err = DB.AutoMigrate(&model.AgentTask{})
	if err != nil {
		logger.Log.Fatal("Failed to auto migrate database schema", zap.Error(err))
	}

	logger.Log.Info("Database initialized and schema migrated successfully")
}
