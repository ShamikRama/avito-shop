package logger

import (
	"go.uber.org/zap"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=Logger
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
}

type LoggerStruct struct {
	logger *zap.Logger
}

func NewLogger() *LoggerStruct {
	logger, _ := zap.NewProduction(zap.AddStacktrace(zap.DPanicLevel))
	return &LoggerStruct{logger: logger}
}

func (l *LoggerStruct) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func (l *LoggerStruct) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}
