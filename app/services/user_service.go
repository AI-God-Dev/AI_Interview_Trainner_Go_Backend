package service

import (
	user_model "up-it-aps-api/app/models/user"
	"up-it-aps-api/platform/database"
)

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) GetTokenUsage(email string) uint64 {
	user := s.GetUserByEmail(email)
	return user.Credits
}

func (s *UserService) DecreaseTokenUsage(email string) user_model.User {
	var db = database.DBConn
	user := s.GetUserByEmail(email)
	if user.Credits == 0 {
		return user
	}
	db.Where("email = ?", email).First(&user).Update("credits", user.Credits-1)
	return user
}

func (s *UserService) UpdateTokens(email string, newTokens uint64) user_model.User {
	var db = database.DBConn
	user := s.GetUserByEmail(email)
	db.Where("email = ?", email).First(&user).Update("credits", user.Credits+newTokens)
	return user
}

func (s *UserService) GetAllUsers() []user_model.User {
	var db = database.DBConn
	var users []user_model.User
	db.Find(&users)
	return users
}

func (s *UserService) GetUserByEmail(email string) user_model.User {
	var db = database.DBConn
	var user user_model.User
	db.Where("email = ?", email).First(&user)
	return user
}

func (s *UserService) GetUserSettingsByEmail(email string) user_model.UserSettings {
	var db = database.DBConn
	var user user_model.User
	db.Where("email = ?", email).First(&user)
	return user.UserSettings
}

func (s *UserService) UpdateUserSettings(email string, newUserSettings *user_model.UserSettings) user_model.UserSettings {
	var db = database.DBConn
	var user user_model.User
	result := db.Where("email = ?", email).First(&user).Select("llm_model", "stt_model", "tts_model", "auto_play_audio").Updates(newUserSettings)
	if result.Error != nil {
		panic(result.Error)
	}
	return *newUserSettings
}

func (s *UserService) CreateUser(user *user_model.InputUser) user_model.User {
	var db = database.DBConn
	var newUser user_model.User
	newUser.Email = user.Email
	newUser.Credits = 300
	result := db.Create(&newUser)
	if result.Error != nil {
		panic(result.Error)
	}
	return newUser
}
