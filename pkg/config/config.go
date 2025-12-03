package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	AI       AIConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Port            string
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type AuthConfig struct {
	JWTSecret              string
	APIKey                 string
	GoogleOAuthClientID    string
	GoogleOAuthSecret      string
	SessionExpiration      time.Duration
	CookieSecure           bool
	CookieSameSite         string
}

type AIConfig struct {
	OpenAIApiKey      string
	ElevenLabsApiKey  string
	UnrealSpeechKey   string
	VertexAIKey       string
	GCloudApiKey      string
	PineconeApiKey    string
	PineconeConnection string
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedHeaders []string
	AllowedMethods []string
}

func Load() (*Config, error) {
	_ = godotenv.Load() // ignore error if .env doesn't exist

	cfg := &Config{}

	cfg.Server.Port = getEnv("PORT", "8080")
	cfg.Server.Host = getEnv("HOST", "0.0.0.0")
	cfg.Server.ReadTimeout = getDurationEnv("READ_TIMEOUT", 15*time.Second)
	cfg.Server.WriteTimeout = getDurationEnv("WRITE_TIMEOUT", 15*time.Second)
	cfg.Server.IdleTimeout = getDurationEnv("IDLE_TIMEOUT", 60*time.Second)
	cfg.Server.ShutdownTimeout = getDurationEnv("SHUTDOWN_TIMEOUT", 10*time.Second)

	cfg.Database.DSN = getRequiredEnv("DSN")
	cfg.Database.MaxOpenConns = getIntEnv("DB_MAX_OPEN_CONNS", 25)
	cfg.Database.MaxIdleConns = getIntEnv("DB_MAX_IDLE_CONNS", 5)
	cfg.Database.ConnMaxLifetime = getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute)

	cfg.Auth.JWTSecret = getRequiredEnv("JWT_SECRET")
	cfg.Auth.APIKey = getRequiredEnv("API_KEY")
	cfg.Auth.GoogleOAuthClientID = getEnv("GOOGLE_OAUTH_CLIENT_ID", "")
	cfg.Auth.GoogleOAuthSecret = getEnv("GOOGLE_OAUTH_CLIENT_SECRET", "")
	cfg.Auth.SessionExpiration = getDurationEnv("SESSION_EXPIRATION", 24*time.Hour)
	cfg.Auth.CookieSecure = getBoolEnv("COOKIE_SECURE", true)
	cfg.Auth.CookieSameSite = getEnv("COOKIE_SAME_SITE", "Lax")

	cfg.AI.OpenAIApiKey = getEnv("OPEN_AI_API_KEY", "")
	cfg.AI.ElevenLabsApiKey = getEnv("ELEVEN_LABS_API_KEY", "")
	cfg.AI.UnrealSpeechKey = getEnv("UNREAL_SPEECH_API_KEY", "")
	cfg.AI.VertexAIKey = getEnv("VERTEX_AI_API_KEY", "")
	cfg.AI.GCloudApiKey = getEnv("GCLOUD_API_KEY", "")
	cfg.AI.PineconeApiKey = getEnv("PINECONE_API_KEY", "")
	cfg.AI.PineconeConnection = getEnv("PINECONE_CONNECTION", "")

	allowedOrigins := getEnv("ALLOWED_ORIGINS", "*")
	cfg.CORS.AllowedOrigins = strings.Split(allowedOrigins, ",")
	cfg.CORS.AllowedHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		"X-API-KEY",
		"X-Request-ID",
	}
	cfg.CORS.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Auth.JWTSecret == "" || c.Auth.JWTSecret == "secret" {
		return fmt.Errorf("JWT_SECRET must be set and not use default value")
	}

	if c.Auth.APIKey == "" {
		return fmt.Errorf("API_KEY must be set")
	}

	if c.Database.DSN == "" {
		return fmt.Errorf("DSN must be set")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return value
}

func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}

