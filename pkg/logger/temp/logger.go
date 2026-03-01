package logger

import (
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

func SetLogger(l *zap.Logger) {
	logger = l
}

func Debug(message string, fields ...zap.Field) {
	if logger == nil {
		l, err := zap.NewDevelopment(zap.AddCallerSkip(1))
		if err == nil {
			l.Debug(message, fields...)
		}
	} else {
		logger.Debug(message, fields...)
	}
}

func Info(message string, fields ...zap.Field) {
	if logger == nil {
		l, err := zap.NewDevelopment(zap.AddCallerSkip(1))
		if err == nil {
			l.Info(message, fields...)
		}
	} else {
		logger.Info(message, fields...)
	}
}

func Warn(message string, fields ...zap.Field) {
	if logger == nil {
		l, err := zap.NewDevelopment(zap.AddCallerSkip(1))
		if err == nil {
			l.Warn(message, fields...)
		}
	} else {
		logger.Warn(message, fields...)
	}
}

func Error(message string, fields ...zap.Field) {
	if logger == nil {
		if l, err := zap.NewDevelopment(zap.AddCallerSkip(1)); err == nil {
			l.Error(message, fields...)
		}
	} else {
		logger.Error(message, fields...)
	}
}
