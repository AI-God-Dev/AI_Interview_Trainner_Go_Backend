package service

import (
	"os"
	"testing"
	user_model "up-it-aps-api/app/models/user"
	"up-it-aps-api/pkg/config"
	"up-it-aps-api/platform/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	err = db.AutoMigrate(&user_model.User{}, &user_model.UserSettings{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func TestUserService_CreateUser(t *testing.T) {
	db := setupTestDB(t)
	database.DBConn = db
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	service := NewUserService()

	input := &user_model.InputUser{
		Email: "test@example.com",
	}

	user, err := service.CreateUser(input)
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	if user.Email != input.Email {
		t.Errorf("CreateUser() email = %v, want %v", user.Email, input.Email)
	}

	if user.Credits != 300 {
		t.Errorf("CreateUser() credits = %v, want 300", user.Credits)
	}

	if user.ID == 0 {
		t.Error("CreateUser() should set user ID")
	}
}

func TestUserService_GetUserByEmail(t *testing.T) {
	db := setupTestDB(t)
	database.DBConn = db
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	service := NewUserService()

	// Create a user first
	input := &user_model.InputUser{Email: "test@example.com"}
	created, err := service.CreateUser(input)
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	// Get the user
	user := service.GetUserByEmail("test@example.com")

	if user.Email != created.Email {
		t.Errorf("GetUserByEmail() email = %v, want %v", user.Email, created.Email)
	}

	// Test non-existent user
	nonExistent := service.GetUserByEmail("nonexistent@example.com")
	if nonExistent.Email != "" {
		t.Error("GetUserByEmail() should return empty user for non-existent email")
	}
}

func TestUserService_DecreaseTokenUsage(t *testing.T) {
	db := setupTestDB(t)
	database.DBConn = db
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	service := NewUserService()

	// Create user with credits
	input := &user_model.InputUser{Email: "test@example.com"}
	_, err := service.CreateUser(input)
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	// Decrease credits
	user := service.DecreaseTokenUsage("test@example.com")

	if user.Credits != 299 {
		t.Errorf("DecreaseTokenUsage() credits = %v, want 299", user.Credits)
	}

	// Test with zero credits
	service.UpdateTokens("test@example.com", 0)
	user = service.DecreaseTokenUsage("test@example.com")
	if user.Credits != 0 {
		t.Error("DecreaseTokenUsage() should not decrease below 0")
	}
}

func TestUserService_UpdateTokens(t *testing.T) {
	db := setupTestDB(t)
	database.DBConn = db
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	service := NewUserService()

	input := &user_model.InputUser{Email: "test@example.com"}
	_, err := service.CreateUser(input)
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	user := service.UpdateTokens("test@example.com", 100)

	if user.Credits != 400 {
		t.Errorf("UpdateTokens() credits = %v, want 400", user.Credits)
	}
}

func TestUserService_GetTokenUsage(t *testing.T) {
	db := setupTestDB(t)
	database.DBConn = db
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	service := NewUserService()

	input := &user_model.InputUser{Email: "test@example.com"}
	_, err := service.CreateUser(input)
	if err != nil {
		t.Fatalf("CreateUser() failed: %v", err)
	}

	credits := service.GetTokenUsage("test@example.com")

	if credits != 300 {
		t.Errorf("GetTokenUsage() = %v, want 300", credits)
	}
}

