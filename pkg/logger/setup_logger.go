package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xyroscar/common-lib/internal/constants"
	"github.com/xyroscar/common-lib/pkg/config"
	logger "github.com/xyroscar/common-lib/pkg/logger/temp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitializeLogger() {
	c := config.GetConfig()

	encoderConfig := zap.NewProductionEncoderConfig()
	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()

	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)

	logLevel := constants.DEFAULT_LOG_LEVEL

	switch c.LoggingConfig.Level {
	case "debug":
		logLevel = zapcore.DebugLevel
	case "info":
		logLevel = zapcore.InfoLevel
	case "warn":
		logLevel = zapcore.WarnLevel
	case "error":
		logLevel = zapcore.ErrorLevel
	}

	var cores []zapcore.Core

	stdOut := zapcore.Lock(os.Stdout)
	cores = append(cores, zapcore.NewCore(
		consoleEncoder,
		stdOut,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl <= zapcore.InfoLevel && lvl >= logLevel
		}),
	))

	stdErr := zapcore.Lock(os.Stderr)
	cores = append(cores, zapcore.NewCore(
		consoleEncoder,
		stdErr,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.WarnLevel && lvl >= logLevel
		}),
	))

	logFileName := fmt.Sprintf("%s-%s", c.AppName, time.Now().Format("2006-01-02"))
	logFilePath := filepath.Join(c.BaseDir, c.LoggingConfig.LogsDir, logFileName)
	fileWriter := zapcore.AddSync(configureLogFiles(&c.LoggingConfig, logFilePath))
	cores = append(cores, zapcore.NewCore(
		jsonEncoder,
		fileWriter,
		logLevel,
	))

	combined := zapcore.NewTee(cores...)

	finalLoggerGlobal := zap.New(
		combined,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	zap.ReplaceGlobals(finalLoggerGlobal)

	finalLoggerWrapper := zap.New(
		combined,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	logger.SetLogger(finalLoggerWrapper)

	zap.L().Info("Logger initialized", zap.String("level", logLevel.String()), zap.String("log file", logFilePath))
}

func configureLogFiles(c *config.LoggingConfig, fileName string) *lumberjack.Logger {
	l := lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    c.MaxFileSize,
		MaxAge:     c.MaxAge,
		MaxBackups: c.MaxNumFiles,
		Compress:   c.Compress,
	}

	return &l
}
