package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set required env vars
	os.Setenv("DSN", "test:test@tcp(localhost:3306)/test")
	os.Setenv("JWT_SECRET", "test-secret-key-12345")
	os.Setenv("API_KEY", "test-api-key-12345")
	defer func() {
		os.Unsetenv("DSN")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("API_KEY")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	if cfg.Database.DSN == "" {
		t.Error("DSN should not be empty")
	}

	if cfg.Auth.JWTSecret == "" {
		t.Error("JWTSecret should not be empty")
	}

	if cfg.Auth.APIKey == "" {
		t.Error("APIKey should not be empty")
	}
}

func TestLoad_Defaults(t *testing.T) {
	os.Setenv("DSN", "test:test@tcp(localhost:3306)/test")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("API_KEY", "test-key")
	defer func() {
		os.Unsetenv("DSN")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("API_KEY")
		os.Unsetenv("PORT")
	}()

	os.Unsetenv("PORT")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.Server.Port)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				Database: DatabaseConfig{DSN: "test:test@tcp(localhost:3306)/test"},
				Auth: AuthConfig{
					JWTSecret: "valid-secret-key",
					APIKey:    "valid-api-key",
				},
			},
			wantErr: false,
		},
		{
			name: "missing DSN",
			cfg: &Config{
				Database: DatabaseConfig{DSN: ""},
				Auth: AuthConfig{
					JWTSecret: "valid-secret",
					APIKey:    "valid-key",
				},
			},
			wantErr: true,
		},
		{
			name: "default JWT secret",
			cfg: &Config{
				Database: DatabaseConfig{DSN: "test:test@tcp(localhost:3306)/test"},
				Auth: AuthConfig{
					JWTSecret: "secret",
					APIKey:    "valid-key",
				},
			},
			wantErr: true,
		},
		{
			name: "missing API key",
			cfg: &Config{
				Database: DatabaseConfig{DSN: "test:test@tcp(localhost:3306)/test"},
				Auth: AuthConfig{
					JWTSecret: "valid-secret",
					APIKey:    "",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "test-value")
	defer os.Unsetenv("TEST_VAR")

	if got := getEnv("TEST_VAR", "default"); got != "test-value" {
		t.Errorf("getEnv() = %v, want test-value", got)
	}

	if got := getEnv("NONEXISTENT", "default"); got != "default" {
		t.Errorf("getEnv() = %v, want default", got)
	}
}

func TestGetIntEnv(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")

	if got := getIntEnv("TEST_INT", 0); got != 42 {
		t.Errorf("getIntEnv() = %v, want 42", got)
	}

	if got := getIntEnv("NONEXISTENT", 10); got != 10 {
		t.Errorf("getIntEnv() = %v, want 10", got)
	}

	os.Setenv("INVALID_INT", "not-a-number")
	defer os.Unsetenv("INVALID_INT")
	if got := getIntEnv("INVALID_INT", 99); got != 99 {
		t.Errorf("getIntEnv() = %v, want 99", got)
	}
}

