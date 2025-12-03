package middleware

import (
	"up-it-aps-api/pkg/errors"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func ErrorHandler(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		if err != nil {
			requestID := GetRequestID(c)
			fields := []zap.Field{
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
				zap.String("requestID", requestID),
			}

			// Check if it's an AppError
			if appErr, ok := err.(*errors.AppError); ok {
				fields = append(fields,
					zap.Int("status", appErr.Code),
					zap.String("message", appErr.Message),
					zap.Error(appErr.Err),
				)
				logger.Error("Application error", fields...)
				return c.Status(appErr.Code).JSON(fiber.Map{
					"error":     true,
					"message":   appErr.Message,
					"code":      appErr.Code,
					"requestID": requestID,
				})
			}

			// Check if it's a Fiber error
			if fiberErr, ok := err.(*fiber.Error); ok {
				fields = append(fields,
					zap.Int("status", fiberErr.Code),
					zap.String("message", fiberErr.Message),
				)
				logger.Error("Fiber error", fields...)
				return c.Status(fiberErr.Code).JSON(fiber.Map{
					"error":     true,
					"message":   fiberErr.Message,
					"code":      fiberErr.Code,
					"requestID": requestID,
				})
			}

			// Unknown error
			fields = append(fields, zap.Error(err))
			logger.Error("Unknown error", fields...)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":     true,
				"message":   "Internal server error",
				"code":      fiber.StatusInternalServerError,
				"requestID": requestID,
			})
		}

		return nil
	}
}

