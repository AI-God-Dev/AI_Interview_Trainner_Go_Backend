package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func LoggingMiddleware(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		logger.Info("HTTP request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("ip", c.IP()),
			zap.String("requestID", c.Locals("requestID").(string)),
		)
		return c.Next()
	}
}
