package middleware

import (
	"up-it-aps-api/pkg/errors"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func APIKeyAuth(apiKey string, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestKey := c.Get("x-api-key")
		if requestKey == "" {
			logger.Warn("Missing API key",
				zap.String("path", c.Path()),
				zap.String("ip", c.IP()),
			)
			return errors.ErrUnauthorized
		}

		if requestKey != apiKey {
			logger.Warn("Invalid API key",
				zap.String("path", c.Path()),
				zap.String("ip", c.IP()),
			)
			return errors.ErrUnauthorized
		}

		return c.Next()
	}
}
