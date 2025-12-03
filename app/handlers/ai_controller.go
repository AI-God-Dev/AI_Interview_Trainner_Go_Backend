package handler

import (
	"bufio"
	"log"
	ai_model "up-it-aps-api/app/models/ai"
	service "up-it-aps-api/app/services"

	"github.com/gofiber/fiber/v2/middleware/session"

	"github.com/gofiber/fiber/v2"
)

type AiHandler struct {
	aiService     *service.AiService
	helperService *service.HelperService
	userService   *service.UserService
	store         *session.Store
}

func NewAiHandler(aiService *service.AiService, helperService *service.HelperService, store *session.Store) *AiHandler {
	return &AiHandler{
		aiService:     aiService,
		helperService: helperService,
		userService:   aiService.userService, // Get userService from aiService
		store:         store,
	}
}

// Placeholder
// @Param request body main.MyHandler.request true "query params"
// @Success 200 {object} main.MyHandler.response
// @Router /test [post]
func (h *AiHandler) ReceiveMessage(c *fiber.Ctx) error {
	log.Println("ReceiveMessage")
	message := new(ai_model.MessageReceived)
	if err := c.BodyParser(message); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	return h.aiService.AiCreateMessage(c, message)

}

func (h *AiHandler) ChunkString(c *fiber.Ctx) error {
	log.Println("ChunkString")
	query := c.Query("text")
	chunkedString := h.aiService.Chunking(query)
	return h.helperService.ChunkData(c, chunkedString)
}

func (h *AiHandler) GenerateChunkedAudio(ctx *fiber.Ctx) (err error) {
	// TODO: refactor this, it's getting messy
	log.Println("GenerateChunkedAudio")
	message := new(ai_model.MessageReceived)
	email := ctx.Query("email")
	if err := ctx.BodyParser(message); err != nil {
		log.Println(err)
		return ctx.Status(400).SendString(err.Error())
	}
	chunkedMessage := h.aiService.Chunking(message.Message)
	ctx.Set("Transfer-Encoding", "chunked")
	userSettings := h.userService.GetUserSettingsByEmail(email)

	// Unreal is faster if it's not chunked, unless the responses are BIG
	if userSettings.TtsModel == "unreal-speech" {
		ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			var audio = h.aiService.UnrealSpeechGenerateAudio([]byte(message.Message), email)
			_, err := w.Write(audio)
			if err != nil {
				log.Printf("Error writing audio: %v", err)
				return
			}
			_ = w.Flush()
		})
		return nil
	}
	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		doneCh := make(chan bool)

		for i := 0; i < len(chunkedMessage); i++ {
			go func(index int) {
				var audio []byte
				switch userSettings.TtsModel {
				case "tts-1":
					audio = h.aiService.OpenAiGenerateAudio(chunkedMessage[index], email)
				case "vertex":
					audio = h.aiService.VertexAiGenerateAudio(chunkedMessage[index])
				case "elevenlabs-multilingual-v1":
					audio = h.aiService.ElevenLabsGenerateAudio(chunkedMessage[index], email)
				default:
					audio = h.aiService.OpenAiGenerateAudio(chunkedMessage[index], email)
				}
				_, err := w.Write(audio)
				if err != nil {
					doneCh <- false
					log.Fatal(err)
					return
				}
				err = w.Flush()
				log.Println("Sending chunk")
				if err != nil {
					log.Printf("Error flushing chunk: %v", err)
					doneCh <- false
					return
				}
				doneCh <- true
			}(i)

			if !<-doneCh { // wait for the goroutine to signal completion
				return
			}
		}
	})
	return nil
}

func (h *AiHandler) WhisperGenerateTextFromSpeech(c *fiber.Ctx) error {
	log.Println("WhisperGenerateTextFromSpeech")
	email := c.Query("email")
	userSettings := h.userService.GetUserSettingsByEmail(email)
	if userSettings.SttModel == "vertex" {
		return h.aiService.VertexAiCreateTranscription(c)
	}
	return h.aiService.OpenAiCreateTranscription(c)
}
