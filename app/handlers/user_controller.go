package handler

import (
	"log"
	user_model "up-it-aps-api/app/models/user"
	service "up-it-aps-api/app/services"

	"github.com/gofiber/fiber/v2/middleware/session"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService *service.UserService
	store       *session.Store
}

func NewUserHandler(userService *service.UserService, store *session.Store) *UserHandler {
	return &UserHandler{userService: userService, store: store}
}

func (h *UserHandler) Logout(c *fiber.Ctx) error {
	log.Println("Logout")
	c.Locals("user", nil)
	c.Status(200)
	return nil
}

func (h *UserHandler) GetAllUsers(c *fiber.Ctx) error {
	log.Println("GetAllUsers")
	user := h.userService.GetAllUsers()
	return c.JSON(user)
}

func (h *UserHandler) GetUserByEmail(c *fiber.Ctx) error {
	log.Println("GetUserByEmail")
	email := c.Query("email")
	user := h.userService.GetUserByEmail(email)
	return c.JSON(user)
}

func (h *UserHandler) GetUserSettingsByEmail(c *fiber.Ctx) error {
	log.Println("GetUserSettingsByEmail")
	email := c.Query("email")
	userSettings := h.userService.GetUserSettingsByEmail(email)
	return c.JSON(userSettings)
}

func (h *UserHandler) UpdateUserSettings(c *fiber.Ctx) error {
	log.Println("UpdateUserSettings")
	email := c.Query("email")
	newUserSettings := new(user_model.UserSettings)
	if err := c.BodyParser(newUserSettings); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	userSettings, err := h.userService.UpdateUserSettings(email, newUserSettings)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   true,
			"message": "failed to update user settings",
		})
	}
	return c.JSON(userSettings)
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	log.Println("CreateUser")
	user := new(user_model.InputUser)
	if err := c.BodyParser(user); err != nil {
		return c.Status(400).SendString(err.Error())
	}
	createdUser, err := h.userService.CreateUser(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   true,
			"message": "failed to create user",
		})
	}
	return c.JSON(createdUser)
}
