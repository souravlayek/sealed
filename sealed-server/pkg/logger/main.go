package logger

import (
	"go.uber.org/zap"
)

var Logger *zap.SugaredLogger

func InitLogger() *zap.SugaredLogger {
	logger := zap.Must(zap.NewProduction()).Sugar()
	Logger = logger
	return logger
}
