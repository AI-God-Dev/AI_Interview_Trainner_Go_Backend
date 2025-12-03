package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger) fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			requestID := GetRequestID(c)
			logger.Error("Panic recovered",
				zap.Any("error", e),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
				zap.String("requestID", requestID),
			)
		},
	})
}

