package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	ai_model "up-it-aps-api/app/models/ai"
	user_model "up-it-aps-api/app/models/user"

	"github.com/form3tech-oss/jwt-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/gofiber/fiber/v2"
)

const (
	ElevenLabsAmericanAccent   = "ErXwobaYiN019PkySvjV"
	ElevenLabsAustralianAccent = "IKne3meq5aSn9XLyUdCD"
	Role                       = "You are highly skilled software engineer for a big tech company. You're currently tutoring a candidate for a SWE (Software Engineer) role in a mock interview scenario. You are well versed in algorithms and data structures and can assist and provide feedback. Be short and concise with your responses. Never respond with more than 40 words."
	// TODO: add APS role prompt back if needed
	OpenAiCompletionsEndpoint     = "https://api.openai.com/v1/chat/completions"
	OpenAiThreadsEndpoint         = "https://api.openai.com/v1/threads/runs"
	OpenAiSingleThreadEndpoint    = "https://api.openai.com/v1/threads/%s/runs/%s/steps"
	OpenAiVoiceGenerationEndpoint = "https://api.openai.com/v1/audio/speech"
	OpenAiTranscriptionEndpoint   = "https://api.openai.com/v1/audio/transcriptions"
	GooglerCustomGptId            = "asst_bzP4wf0kl1XWZcup65OZ27Gf"
	MetaMateCustomGptId           = "asst_KeLCxkHbYlNHYKB4ccYyXQfd"
	UnrealSpeechStreamEndpoint    = "https://api.v6.unrealspeech.com/stream"
	ElevenLabsStreamEndpoint      = "https://api.elevenlabs.io/v1/text-to-speech/%s/stream"
	VertexTranscriptionEndpoint   = "https://speech.googleapis.com/v1p1beta1/speech:recognize"
	VertexTextGenerationEndpoint  = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"
)

type AiService struct {
	userService *UserService
}

// NewAiService creates a new AI service instance
func NewAiService(userService *UserService) *AiService {
	return &AiService{
		userService: userService,
	}
}

func (s *AiService) AiCreateMessage(c *fiber.Ctx, ai *ai_model.MessageReceived) (err error) {
	email := c.Query("email")
	user := s.userService.GetUserByEmail(email)
	if user.Credits <= 0 {
		return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{
			"message": "You have no more credits left",
			"status":  "error",
		})
	}

	if user.UserSettings.LlmModel == "chat-bison" || user.UserSettings.LlmModel == "gemini-pro" {
		return VertexAiCreateMessage(c, ai, user.UserSettings.LlmModel, Role)
	}

	if user.UserSettings.LlmModel == "googler" || user.UserSettings.LlmModel == "meta-mate" {
		s.OpenAiCreateThreadForAssistant(c, ai, user.UserSettings.LlmModel)
		return
	}
	return OpenAiCreateMessage(user, c, ai)
}

func OpenAiCreateMessage(user user_model.User, c *fiber.Ctx, ai *ai_model.MessageReceived) (err error) {
	apiKey := os.Getenv("OPEN_AI_API_KEY")
	agent := fiber.Post(OpenAiCompletionsEndpoint)
	auth := fmt.Sprint("Bearer ", apiKey)
	agent.Set("Authorization", auth)
	agent.Set("Content-Type", "application/json")

	jsonBody := ai_model.OpenAiRequest{
		Model: user.UserSettings.LlmModel,
		Messages: []ai_model.MessageRequest{
			{
				Role:    "system",
				Content: Role,
			},
			{
				Role:    "user",
				Content: ai.Message,
			},
		},
	}
	agent.JSON(jsonBody)

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errs": errs,
		})
	}

	var chatGptResponse ai_model.OpenAiChatResponse
	err = json.Unmarshal(body, &chatGptResponse)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"err": err,
		})
	}

	transformedData := TransformOpenAiData(chatGptResponse)
	return c.Status(statusCode).JSON(fiber.Map{
		"message": transformedData.MessageRetrieved,
		"status":  "success",
	})
}

func GetCustomGptAssistant(assistant string) string {
	switch assistant {
	/*These are custom Open AI GPTs with various helpful documentation for either of these companies,
	only accessible via the original projects open ai api keys*/
	case "googler":
		return GooglerCustomGptId
	case "meta-mate":
		return MetaMateCustomGptId
	default:
		return GooglerCustomGptId
	}
}

func (s *AiService) OpenAiCreateThreadForAssistant(c *fiber.Ctx, ai *ai_model.MessageReceived, assistant string) (err error) {
	apiKey := os.Getenv("OPEN_AI_API_KEY")
	agent := fiber.Post(OpenAiThreadsEndpoint)
	auth := fmt.Sprint("Bearer ", apiKey)
	agent.Set("Authorization", auth)
	agent.Set("OpenAI-Beta", "assistants=v1")
	agent.Set("Content-Type", "application/json")

	jsonBody := ai_model.OpenAiThreadRequest{
		AssistantID: GetCustomGptAssistant(assistant),
		Thread: ai_model.OpenAiThread{Messages: []ai_model.MessageRequest{
			{
				Role:    "user",
				Content: ai.Message,
			},
		},
		},
	}
	agent.JSON(jsonBody)

	_, body, _ := agent.Bytes()

	var openAiThreadResponse ai_model.OpenAiThreadResponse
	err = json.Unmarshal(body, &openAiThreadResponse)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err,
		})
	}
	if openAiThreadResponse.Id == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Thread ID is empty",
		})
	}

	var messageId string
	var wg sync.WaitGroup
	var mu sync.Mutex
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

Loop:
	for i := 0; i < 200; i++ { //arbitrary number of times to try and get the message id
		wg.Add(1)
		time.Sleep(200 * time.Millisecond)
		go func(i int) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				singleThreadUrl := fmt.Sprintf(OpenAiSingleThreadEndpoint, openAiThreadResponse.ThreadId, openAiThreadResponse.Id)
				singleThreadAgent := fiber.Get(singleThreadUrl)
				singleThreadAgent.Set("Authorization", auth)
				singleThreadAgent.Set("Content-Type", "application/json")
				singleThreadAgent.Set("OpenAI-Beta", "assistants=v1")
				_, singleThreadBody, _ := singleThreadAgent.Bytes()

				var threadRunStepResponse ai_model.OpenAiThreadRunStepResponse
				err = json.Unmarshal(singleThreadBody, &threadRunStepResponse)

				if err != nil {
					fmt.Println(err)
				}

				if len(threadRunStepResponse.Data) > 0 && threadRunStepResponse.Data[0].StepDetails.MessageCreation.MessageID != "" {
					mu.Lock()
					messageId = threadRunStepResponse.Data[0].StepDetails.MessageCreation.MessageID
					mu.Unlock()
					cancel()
					return
				}
			}
		}(i)
		select {
		case <-ctx.Done():
			break Loop
		default:
		}
	}
	var messageBody []byte
	wg.Add(1)
	go func() {
		defer wg.Done()
		messageBody = OpenAiGetMessageFromThread(openAiThreadResponse.ThreadId, messageId)
	}()
	wg.Wait()
	var stringifiedMessage = string(messageBody)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": stringifiedMessage,
		"status":  "success",
	})
}

func OpenAiGetMessageFromThread(threadId string, messageId string) []byte {
	url := fmt.Sprintf(OpenAiSingleThreadEndpoint, threadId, messageId)
	apiKey := os.Getenv("OPEN_AI_API_KEY")
	auth := "Bearer " + apiKey
	agent := fiber.Get(url)
	agent.Set("Authorization", auth)
	agent.Set("Content-Type", "application/json")
	agent.Set("OpenAI-Beta", "assistants=v1")
	_, body, _ := agent.Bytes()

	var openAiThreadMessageResponse ai_model.OpenAiThreadMessageResponse
	err := json.Unmarshal(body, &openAiThreadMessageResponse)
	if err != nil {
		fmt.Println(err)
	}
	return []byte(openAiThreadMessageResponse.Content[0].Text.Value)
}

func VertexAiCreateMessage(c *fiber.Ctx, ai *ai_model.MessageReceived, llmModel string, role string) (err error) {
	jsonBody := ai_model.GoogleRequest{
		Contents: []ai_model.GoogleRequestContent{
			{
				Parts: []ai_model.GoogleRequestPart{
					{
						Text: role + " " + ai.Message,
					},
				},
			},
		},
		SafetySettings: []ai_model.GoogleRequestSafety{
			{
				Category:  "HARM_CATEGORY_DANGEROUS_CONTENT",
				Threshold: "BLOCK_ONLY_HIGH",
			},
		},
		GenerationConfig: ai_model.GoogleGenerationConfig{
			Temperature:     1.0,
			TopP:            0.8,
			TopK:            10,
			MaxOutputTokens: 125,
		},
	}
	apiKey := os.Getenv("VERTEX_AI_API_KEY") //using the old gen ai maker method
	url := fmt.Sprintf(VertexTextGenerationEndpoint, llmModel, apiKey)
	agent := fiber.Post(url)
	agent.Set("Content-Type", "application/json")
	agent.JSON(jsonBody)
	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errs": errs,
		})
	}
	if statusCode == 400 {
		return c.Status(fiber.StatusUnauthorized).JSON(body)
	}
	if statusCode == 401 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var googleResponse ai_model.GoogleResponse
	err = json.Unmarshal(body, &googleResponse)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"err": err,
		})
	}

	transformedData := TransformGoogleData(googleResponse)

	return c.Status(statusCode).JSON(fiber.Map{
		"message": transformedData.MessageRetrieved,
		"status":  "success",
	})
}

func TransformGoogleData(responseReceived ai_model.GoogleResponse) ai_model.Response {
	return ai_model.Response{
		MessageRetrieved: responseReceived.Candidates[0].Content.Parts[0].Text,
	}
}

func TransformOpenAiData(responseReceived ai_model.OpenAiChatResponse) ai_model.Response {
	return ai_model.Response{
		MessageRetrieved: responseReceived.Choices[0].Message.Content,
	}
}

func (s *AiService) VertexAiGenerateAudio(message []byte) (output []byte) {
	url := fmt.Sprintf("https://texttospeech.googleapis.com/v1/text:synthesize")
	agent := fiber.Post(url)
	apiKey := os.Getenv("GCLOUD_API_KEY")

	agent.Set("Authorization", "Bearer "+apiKey)
	agent.Set("Accept", "audio/mpeg")
	agent.Set("Content-Type", "application/json; charset=utf-8")
	agent.Set("x-goog-user-project", "up-it-aps")
	vertexAudioRequest := ai_model.GoogleVertexAiRequest{
		Input: ai_model.GoogleVertexAiAudioRequestInput{
			Text: string(message),
		},
		Voice: ai_model.GoogleVertexAiAudioRequestVoice{
			LanguageCode: "en-AU",
			Name:         "en-AU-Neural2-B",
		},
		AudioConfig: ai_model.GoogleVertexAiAudioRequestAudioConfig{
			AudioEncoding: "MP3",
			SpeakingRate:  1.0,
		},
	}

	agent.JSON(vertexAudioRequest)
	_, body, _ := agent.Bytes()
	vertexResponse := ai_model.GoogleVertexAiAudioResponse{}
	err := json.Unmarshal(body, &vertexResponse)
	if err != nil {
		return nil
	}
	decodedBytes, err := base64.StdEncoding.DecodeString(vertexResponse.AudioContent)
	if err != nil {
		fmt.Println("Error decoding Base64 string:", err)
		return
	}
	reader := io.NopCloser(bytes.NewReader(decodedBytes))
	byteArray, _ := io.ReadAll(reader)
	output = byteArray

	return output
}

func (s *AiService) UnrealSpeechGenerateAudio(message []byte, email string) (output []byte) {
	url := fmt.Sprintf(UnrealSpeechStreamEndpoint)
	agent := fiber.Post(url)
	apiKey := os.Getenv("UNREAL_SPEECH_API_KEY")

	agent.Set("Authorization", "Bearer "+apiKey)
	agent.Set("Accept", "audio/mpeg")
	agent.Set("Content-Type", "application/json; charset=utf-8")
	agent.Set("x-goog-user-project", "up-it-aps")
	unrealSpeechAudioRequest := ai_model.UnrealSpeechRequest{
		Text:    string(message),
		VoiceId: "Liv",
		Bitrate: "64k",
		Speed:   "0",
		Pitch:   "1",
		Codec:   "libmp3lame",
	}

	agent.JSON(unrealSpeechAudioRequest)
	_, body, _ := agent.Bytes()
	s.userService.DecreaseTokenUsage(email)
	reader := io.NopCloser(bytes.NewReader(body))
	byteArray, _ := io.ReadAll(reader)
	output = byteArray
	return output
}

func (s *AiService) OpenAiGenerateAudio(message []byte, email string) (output []byte) {
	url := fmt.Sprintf(OpenAiVoiceGenerationEndpoint)
	agent := fiber.Post(url)
	apiKey := os.Getenv("OPEN_AI_API_KEY")
	agent.Set("Authorization", "Bearer "+apiKey)
	agent.Set("Accept", "audio/mpeg")
	agent.Set("Content-Type", "application/json")
	messageReceived := string(message)

	jsonBody := ai_model.OpenAiTtsRequest{
		Model: "tts-1",
		Input: messageReceived,
		Voice: "alloy",
	}
	agent.JSON(jsonBody)
	_, body, _ := agent.Bytes()
	s.userService.DecreaseTokenUsage(email)
	reader := io.NopCloser(bytes.NewReader(body))
	byteArray, _ := io.ReadAll(reader)
	output = byteArray
	return output
}

func (s *AiService) ElevenLabsGenerateAudio(message []byte, email string) (output []byte) {
	url := fmt.Sprintf(ElevenLabsStreamEndpoint, ElevenLabsAmericanAccent)
	agent := fiber.Post(url)
	apiKey := os.Getenv("ELEVEN_LABS_API_KEY")

	agent.Set("xi-api-key", apiKey)
	agent.Set("Accept", "audio/mpeg")
	agent.Set("Content-Type", "application/json")
	messageReceived := string(message)

	jsonBody := ai_model.ElevenLabsRequest{
		Text:                     messageReceived,
		ModelID:                  "eleven_turbo_v2",
		OptimizeStreamingLatency: 3,
		VoiceSettings: ai_model.ElevenLabsVoiceSettings{
			Stability:       0.95,
			SimilarityBoost: 0.95,
		},
	}
	agent.JSON(jsonBody)
	_, body, _ := agent.Bytes()
	s.userService.DecreaseTokenUsage(email)
	reader := io.NopCloser(bytes.NewReader(body))
	byteArray, _ := io.ReadAll(reader)
	output = byteArray
	return output
}

func (s *AiService) Chunking(input string) (output [][]byte) {
	r := regexp.MustCompile(`[!?.,]`)
	result := r.Split(input, -1)

	for _, v := range result {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			fmt.Println(trimmed)
			output = append(output, []byte(trimmed))
		}
	}

	return output
}

func (s *AiService) OpenAiCreateTranscription(c *fiber.Ctx) (err error) {
	apiKey := os.Getenv("OPEN_AI_API_KEY")
	fmt.Println("Running OpenAiCreateTranscription")
	jsonData := c.Body()

	var audio struct {
		AudioData []byte `json:"audioData"`
	}

	if err := json.Unmarshal([]byte(jsonData), &audio); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil
	}

	uint8Array := audio.AudioData
	audioFile, err := os.Create("audio.wav")

	if err != nil {
		fmt.Println("Error creating audio file:", err)
		return nil
	}
	fmt.Println("Created audio file")
	defer func(audioFile *os.File) {
		err := audioFile.Close()
		if err != nil {
			fmt.Println("Error closing audio file:", err)
		}
	}(audioFile)

	if _, err := audioFile.Write(uint8Array); err != nil {
		fmt.Println("Error writing to audio file:", err)
		return nil
	}

	url := OpenAiTranscriptionEndpoint
	agent := fiber.Post(url)
	agent.Set("Authorization", "Bearer "+apiKey)
	agent.Set("Content-Type", "multipart/form-data")

	var args = fiber.AcquireArgs()
	args.Add("model", "whisper-1")
	args.Add("language", "en")

	var formFile = fiber.AcquireFormFile()
	formFile.Name = "audio.wav"
	formFile.Fieldname = "file"
	formFile.Content = uint8Array

	agent.FileData(formFile).MultipartForm(args)

	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errs": errs,
		})
	}
	return c.Status(statusCode).Send(body)
}

func (s *AiService) VertexAiCreateTranscription(c *fiber.Ctx) (err error) {
	url := fmt.Sprintf(VertexTranscriptionEndpoint)
	agent := fiber.Post(url)
	apiKey := os.Getenv("GCLOUD_API_KEY")
	agent.Set("Authorization", "Bearer "+apiKey)
	agent.Set("Content-Type", "application/json; charset=utf-8")
	agent.Set("x-goog-user-project", "up-it-aps") //replace with your project id

	retrievedJson := c.Body()
	var audio struct {
		AudioData []byte `json:"audioData"`
	}

	if err := json.Unmarshal([]byte(retrievedJson), &audio); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil
	}

	encodedString := base64.StdEncoding.EncodeToString(audio.AudioData)
	if err != nil {
		fmt.Println("Error decoding Base64 string:", err)
		return
	}

	jsonBody := ai_model.GoogleVertexAiSpeechToTextRequest{
		Config: ai_model.GoogleVertexAiSpeechToTextRequestConfig{
			LanguageCode:          "en-AU",
			EnableWordTimeOffsets: true,
			EnableWordConfidence:  true,
			Model:                 "default",
			Encoding:              "MP3",
			SampleRateHertz:       24000,
			AudioChannelCount:     1,
		},
		Audio: ai_model.GoogleVertexAiSpeechToTextAudio{
			Content: encodedString,
		},
	}
	response := agent.JSON(jsonBody)
	_, body, errs := response.Bytes()
	if len(errs) > 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errs": errs,
		})
	}

	var vertexResponse ai_model.GoogleVertexAiSpeechToTextResponse
	err = json.Unmarshal(body, &vertexResponse)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"err": err,
		})
	}
	return c.Status(200).JSON(fiber.Map{
		"text": vertexResponse.VertexAiSpeechToTextResponseResults[0].Alternatives[0].Transcript,
	})
}

func validateToken(tokenString string) (*jwt.Token, error) {
	// TODO: implement proper token validation
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		googleJson, _ := os.ReadFile("./google.json")
		googleJsonParsed, _ := google.JWTConfigFromJSON(googleJson)
		return googleJsonParsed.PrivateKey, nil
	})

	if err != nil {
		return nil, err
	}

	// Check token validity
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// You can add more validations based on your use case here
		// For example, checking expiration, issuer, subject, etc.
		if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
			return nil, fmt.Errorf("Token is expired")
		}
		return token, nil
	}

	return nil, fmt.Errorf("Invalid token")
}

func getIdToken() (*oauth2.Token, error) {
	// TODO: implement OAuth token generation
	data, err := os.ReadFile("./google.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read the key file: %v", err)
	}

	// Generate the token
	tokenSource, err := google.CredentialsFromJSON(context.Background(), data, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("unable to generate token: %v", err)
	}
	token, err := tokenSource.TokenSource.Token()
	return token, nil
}
