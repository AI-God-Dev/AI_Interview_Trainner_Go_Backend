//go:build integration
// +build integration

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"up-it-aps-api/pkg/config"
	"up-it-aps-api/pkg/logger"
	"up-it-aps-api/pkg/middleware"
	"up-it-aps-api/pkg/routes"
	"up-it-aps-api/platform/database"

	"github.com/gofiber/fiber/v2/middleware/session"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	user_model "up-it-aps-api/app/models/user"
)

func setupTestApp(t *testing.T) *fiber.App {
	// Set test env vars
	os.Setenv("DSN", "test.db")
	os.Setenv("JWT_SECRET", "test-secret-key-for-integration-tests")
	os.Setenv("API_KEY", "test-api-key-for-integration")
	os.Setenv("ENV", "test")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config load failed: %v", err)
	}

	appLogger, err := logger.New("test")
	if err != nil {
		t.Fatalf("logger init failed: %v", err)
	}

	// Use in-memory SQLite for tests
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db init failed: %v", err)
	}

	err = db.AutoMigrate(&user_model.User{}, &user_model.UserSettings{})
	if err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	database.DBConn = db

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler(appLogger.Logger),
	})

	app.Use(middleware.Recovery(appLogger.Logger))
	app.Use(middleware.RequestID())

	store := session.New(session.Config{
		Expiration:     cfg.Auth.SessionExpiration,
		KeyLookup:      "cookie:session",
		CookieSecure:   false,
		CookieSameSite: "Lax",
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api")
	api.Use(middleware.APIKeyAuth(cfg.Auth.APIKey, appLogger.Logger))

	routes.UserRoutes(api, store)

	return app
}

func TestHealthCheck(t *testing.T) {
	app := setupTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestCreateUser(t *testing.T) {
	app := setupTestApp(t)

	userData := map[string]string{
		"email": "test@example.com",
	}
	body, _ := json.Marshal(userData)

	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", "test-api-key-for-integration")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestGetUserByEmail(t *testing.T) {
	app := setupTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/users?email=test@example.com", nil)
	req.Header.Set("x-api-key", "test-api-key-for-integration")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestAPIKeyAuth(t *testing.T) {
	app := setupTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/users?email=test@example.com", nil)
	// No API key header

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test() failed: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}

