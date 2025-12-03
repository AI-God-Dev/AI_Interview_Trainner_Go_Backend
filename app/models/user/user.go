package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID           uint         `gorm:"primary_key"`
	Email        string       `json:"email"`
	Credits      uint64       `json:"credits" gorm:"default:0"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	UserSettings UserSettings `gorm:"embedded"`
}

type UserSettings struct {
	Email         string `json:"email" gorm:"primary_key"`
	LlmModel      string `json:"llm_model" gorm:"default:gpt-3.5-turbo"`
	SttModel      string `json:"stt_model" gorm:"default:whisper-1"`
	TtsModel      string `json:"tts_model" gorm:"default:elevenlabs-multilingual-v1"`
	AutoPlayAudio bool   `json:"auto_play_audio" gorm:"default:true"`
}

type InputUser struct {
	Email string `json:"email"`
}

type Tokens struct {
	Credits          int `json:"credits"`
	ListeningCredits int `json:"listeningCredits"`
	SpeakingCredits  int `json:"speakingCredits"`
}
