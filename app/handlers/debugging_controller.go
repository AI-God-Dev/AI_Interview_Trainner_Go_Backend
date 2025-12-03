package handler

import (
	"strconv"
	service "up-it-aps-api/app/services"

	"github.com/gofiber/fiber/v2"
)

type DebuggingHandler struct {
	userService *service.UserService
}

func NewDebuggingHandler() *DebuggingHandler {
	return &DebuggingHandler{}
}
func (h *DebuggingHandler) Debugging(c *fiber.Ctx) error {
	// quick debug endpoint, remove in prod
	request := c.Request()
	response := c.Response()
	contentType := request.Header.ContentType()
	contentLength := request.Header.ContentLength()

	multipartForm, err := c.MultipartForm()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"contentType":    string(contentType),
		"contentLength":  contentLength,
		"hasMultipart":   multipartForm != nil,
		"statusCode":     response.StatusCode(),
	})
}

func (h *DebuggingHandler) GetUserDetails(c *fiber.Ctx) error {
	email := c.Query("email")
	user := h.userService.GetUserByEmail(email)
	return c.JSON(user)
}

func (h *DebuggingHandler) UpdateUserTokens(c *fiber.Ctx) error {
	email := c.Query("email")
	tokenAmount := c.Query("tokenAmount")
	token, _ := strconv.ParseInt(tokenAmount, 10, 64)
	intToken := uint64(token)
	user := h.userService.UpdateTokens(email, intToken)
	return c.JSON(user)
}
