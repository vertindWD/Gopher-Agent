package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func InitLogger() {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // 时间格式：2023-10-27T10:00:00

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // 生产环境用 JSON，方便接入 ELK
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)

	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
}
