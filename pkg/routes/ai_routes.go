package routes

import (
	handler "up-it-aps-api/app/handlers"
	service "up-it-aps-api/app/services"
	"up-it-aps-api/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func AiRoutes(api fiber.Router, store *session.Store) {
	userService := service.NewUserService()
	aiService := service.NewAiService(userService)
	helperService := &service.HelperService{}
	aiHandler := handler.NewAiHandler(aiService, helperService, store)
	ai := api.Group("/ai")

	ai.Post("/generate-audio", aiHandler.GenerateChunkedAudio)
	ai.Post("/chunk", aiHandler.ChunkString)
	ai.Post("/message", aiHandler.ReceiveMessage)
	ai.Post("/speech-to-text", aiHandler.WhisperGenerateTextFromSpeech)
}
