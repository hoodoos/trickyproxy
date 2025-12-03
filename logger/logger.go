package logger

import "go.uber.org/zap"

var accessLogger *zap.Logger

func InitAccessLog(logger *zap.Logger) {
	accessLogger = logger
}

func WriteAccessLog(fields ...zap.Field) {
	if accessLogger != nil {
		accessLogger.Info("access", fields...)
	}
}
