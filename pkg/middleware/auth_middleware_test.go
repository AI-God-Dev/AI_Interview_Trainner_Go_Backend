package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestAPIKeyAuth(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validAPIKey := "test-api-key-12345"

	app := fiber.New()
	app.Use(APIKeyAuth(validAPIKey, logger))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	tests := []struct {
		name           string
		apiKey         string
		expectedStatus int
	}{
		{
			name:           "valid API key",
			apiKey:         validAPIKey,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing API key",
			apiKey:         "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid API key",
			apiKey:         "wrong-key",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.apiKey != "" {
				req.Header.Set("x-api-key", tt.apiKey)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test() failed: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestRequestID(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/test", func(c *fiber.Ctx) error {
		requestID := GetRequestID(c)
		if requestID == "" {
			return c.Status(500).SendString("no request ID")
		}
		return c.SendString(requestID)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Check if request ID is in response header
	requestID := resp.Header.Get("X-Request-ID")
	if requestID == "" {
		t.Error("X-Request-ID header should be set")
	}
}

