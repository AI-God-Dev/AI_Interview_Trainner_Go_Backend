package logger

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func New(env string) (*Logger, error) {
	var config zap.Config

	if env == "production" || env == "prod" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	log, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{Logger: log}, nil
}

func (l *Logger) FiberLogger() fiber.Handler {
	return logger.New(logger.Config{
		Format:     "${time} ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: time.RFC3339,
		Output:     os.Stdout,
	})
}

func (l *Logger) Sync() {
	_ = l.Logger.Sync()
}

