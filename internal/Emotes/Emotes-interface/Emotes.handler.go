package Emotesinterface

import (
	Emotesapplication "PINKKER-BACKEND/internal/Emotes/Emotes-application"

	"github.com/gofiber/fiber/v2"
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
