package clipinterface

import (
	clipapplication "PINKKER-BACKEND/internal/clip/clip-application"

	"github.com/gofiber/fiber/v2"
)

type ClipHandler struct {
	ClipService *clipapplication.ClipService
}

func NewClipHandler(ClipService *clipapplication.ClipService) *ClipHandler {
	return &ClipHandler{
		ClipService: ClipService,
	}
}

func (clip *ClipHandler) CreateClip(c *fiber.Ctx) error {
	var body struct {
		Streamer  string `json:"streamer"`
		StartTime string `json:"startTime"`
		Length    string `json:"length"`
		ClipName  string `json:"clipName"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}
	user, findUserErr := clip.ClipService.FindUser(body.Streamer)
	if findUserErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": findUserErr.Error(),
		})
	}
	err := clip.ClipService.CreateClip(user, body.StartTime, body.Length, body.ClipName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Created clip successfully!",
	})
}
