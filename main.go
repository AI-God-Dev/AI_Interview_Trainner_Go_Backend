package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	user_model "up-it-aps-api/app/models/user"
	service "up-it-aps-api/app/services"
	_ "up-it-aps-api/docs"
	"up-it-aps-api/pkg/config"
	"up-it-aps-api/pkg/logger"
	"up-it-aps-api/pkg/middleware"
	"up-it-aps-api/pkg/routes"
	"up-it-aps-api/platform/database"

	"github.com/form3tech-oss/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/swagger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"go.uber.org/zap"
)

// @title AI Interview Trainer API
// @version 1.0
// @host localhost:8080
func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("config load failed: %v", err))
	}

	appLogger, err := logger.New(os.Getenv("ENV"))
	if err != nil {
		panic(fmt.Sprintf("logger init failed: %v", err))
	}
	defer appLogger.Sync()

	appLogger.Info("starting app", zap.String("env", os.Getenv("ENV")))

	db, err := initDatabase(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}()

	app := fiber.New(fiber.Config{
		DisablePreParseMultipartForm: true,
		StreamRequestBody:            true,
		PassLocalsToViews:            true,
		ReadTimeout:                  cfg.Server.ReadTimeout,
		WriteTimeout:                 cfg.Server.WriteTimeout,
		IdleTimeout:                  cfg.Server.IdleTimeout,
		ErrorHandler: middleware.ErrorHandler(appLogger),
	})

	setupMiddleware(app, cfg, appLogger)

	store := session.New(session.Config{
		Expiration:     cfg.Auth.SessionExpiration,
		KeyLookup:      "cookie:session",
		CookieSecure:   cfg.Auth.CookieSecure,
		CookieSameSite: cfg.Auth.CookieSameSite,
	})

	setupRoutes(app, store, cfg, appLogger)

	go func() {
		addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
		appLogger.Info("server starting", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			appLogger.Fatal("server start failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	appLogger.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		appLogger.Error("shutdown error", zap.Error(err))
	}

	appLogger.Info("exited")
}

func setupMiddleware(app *fiber.App, cfg *config.Config, appLogger *logger.Logger) {
	app.Use(middleware.Recovery(appLogger.Logger))
	app.Use(middleware.RequestID())
	app.Use(appLogger.FiberLogger())

	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     cfg.CORS.AllowedOrigins[0],
		AllowHeaders:     cfg.CORS.AllowedHeaders,
		AllowMethods:     cfg.CORS.AllowedMethods,
	}))

	app.Get("/swagger/*", swagger.HandlerDefault)
}

func setupRoutes(app *fiber.App, store *session.Store, cfg *config.Config, appLogger *logger.Logger) {
	app.Get("/", healthCheck)
	app.Get("/health", healthCheck)

	api := app.Group("/api")
	api.Use(middleware.APIKeyAuth(cfg.Auth.APIKey, appLogger.Logger))

	auth := api.Group("/auth")
	auth.Get("/callback", handleLoginCallback(cfg.Auth.JWTSecret, appLogger))
	auth.Get("/logout", handleLogout(store))

	routes.AiRoutes(api, store)
	routes.UserRoutes(api, store)
	routes.DebuggingRoutes(api, store)
}

func initDatabase(cfg *config.Config, appLogger *logger.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("db connect failed: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get db instance failed: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	database.DBConn = db

	if err := db.AutoMigrate(&user_model.User{}, &user_model.UserSettings{}); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	appLogger.Info("db connected")
	return db, nil
}

func healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func handleLogout(store *session.Store) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := store.Get(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to get session",
			})
		}

		if err := sess.Destroy(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to destroy session",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Logged out successfully",
		})
	}
}

func handleLoginCallback(jwtSecret string, appLogger *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userService := service.NewUserService()
		tokenString := c.Get("Authorization")

		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Missing authorization token",
			})
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			appLogger.Warn("invalid jwt", zap.Error(err))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "invalid token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "invalid token claims",
			})
		}

		email, ok := claims["email"].(string)
		if !ok || email == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "invalid email in token",
			})
		}

		name, _ := claims["name"].(string)

		retrievedUser := userService.GetUserByEmail(email)
		if retrievedUser.Email == "" {
			newUser := user_model.InputUser{Email: email}
			created, err := userService.CreateUser(&newUser)
			if err != nil {
				appLogger.Error("failed to create user", zap.Error(err), zap.String("email", email))
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   true,
					"message": "failed to create user",
				})
			}
			retrievedUser = created
			appLogger.Info("new user created", zap.String("email", email))
		}

		return c.JSON(fiber.Map{
			"email": email,
			"name":  name,
		})
	}
}
