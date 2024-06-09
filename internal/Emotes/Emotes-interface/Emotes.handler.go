package Emotesinterface

import (
	EmotesDomain "PINKKER-BACKEND/internal/Emotes/Emotes"
	Emotesapplication "PINKKER-BACKEND/internal/Emotes/Emotes-application"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmotesRepository struct {
	Servise *Emotesapplication.EmotesService
}

func NewwithdrawService(Servise *Emotesapplication.EmotesService) *EmotesRepository {
	return &EmotesRepository{
		Servise: Servise,
	}
}
func (s *EmotesRepository) GetGlobalEmotes(c *fiber.Ctx) error {
	GlobalEmotes, err := s.Servise.GetGlobalEmotes()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    GlobalEmotes,
	})
}
func (s *EmotesRepository) GetPinkkerEmotes(c *fiber.Ctx) error {
	GlobalEmotes, err := s.Servise.GetPinkkerEmotes()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    GlobalEmotes,
	})
}
func (s *EmotesRepository) UpdateEmoteAut(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var req EmotesDomain.EmoteUpdate

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	Emotes, err := s.Servise.UpdateEmoteAut(req, idValueObj)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Emotes,
	})
}
