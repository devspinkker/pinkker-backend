package Emotesinterface

import (
	EmotesDomain "PINKKER-BACKEND/internal/Emotes/Emotes"
	Emotesapplication "PINKKER-BACKEND/internal/Emotes/Emotes-application"
	"PINKKER-BACKEND/pkg/helpers"

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
func (s *EmotesRepository) CreateOrUpdateEmoteWithImage(c *fiber.Ctx) error {
	IdUserToken := c.Context().UserValue("_id").(string)
	nameUser := c.Context().UserValue("nameUser").(string)

	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}

	var req EmotesDomain.EmoteUpdateOrCreate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
		})
	}
	fileHeader, _ := c.FormFile("emoteImage")
	PostImageChanel := make(chan string)
	errChanel := make(chan error)
	go helpers.ProcessImageEmotes(fileHeader, PostImageChanel, errChanel, nameUser, req.TypeEmote)

	select {
	case imageUrl := <-PostImageChanel:
		emote := EmotesDomain.EmotePair{
			Name: req.Name,
			URL:  imageUrl,
		}

		createdEmote, err := s.Servise.CreateOrUpdateEmote(IdUserTokenP, req.TypeEmote, emote)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Internal Server Error",
				"error":   err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Emote creado o actualizado exitosamente",
			"data":    createdEmote,
		})

	case err := <-errChanel:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error al procesar la imagen",
			"error":   err.Error(),
		})
	}
}
func (s *EmotesRepository) DeleteEmoteForType(c *fiber.Ctx) error {
	IdUserToken := c.Context().UserValue("_id").(string)

	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}

	var req EmotesDomain.EmoteUpdateOrCreate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
		})
	}
	err := s.Servise.DeleteEmoteForType(IdUserTokenP, req.Name, req.TypeEmote)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
func (s *EmotesRepository) GetEmoteIdUserandType(c *fiber.Ctx) error {

	var req EmotesDomain.GetEmoteIdUserandType
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
		})
	}

	Emote, err := s.Servise.GetEmoteIdUserandType(req.IdUser, req.TypeEmote)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    Emote,
	})

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
func (s *EmotesRepository) AddEmoteAut(c *fiber.Ctx) error {
	// Extract user ID from context
	idValue := c.Context().UserValue("_id").(string)
	nameUser := c.Context().UserValue("nameUser").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
		})
	}

	// Parse request body into EmoteUpdate struct
	var req EmotesDomain.EmoteUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
		})
	}

	fileHeader, _ := c.FormFile("emoteImage")
	PostImageChanel := make(chan string)
	errChanel := make(chan error)

	go helpers.ProcessImageEmotes(fileHeader, PostImageChanel, errChanel, nameUser+"aut", req.Type)

	select {
	case imageUrl := <-PostImageChanel:
		req.Emotes = EmotesDomain.EmotePair{
			URL:  imageUrl,
			Name: req.Name,
		}
		updatedEmote, err := s.Servise.AddEmoteAut(req, idValueObj)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Internal Server Error",
				"error":   err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "OK",
			"data":    updatedEmote,
		})

	case err := <-errChanel:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error processing image",
			"error":   err.Error(),
		})
	}
}
func (s *EmotesRepository) DeleteEmoteAut(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
		})
	}

	var req EmotesDomain.EmoteUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
		})
	}

	req.Emotes = EmotesDomain.EmotePair{
		URL:  "",
		Name: req.Name,
	}
	updatedEmote, err := s.Servise.DeleteEmoteAut(req, idValueObj)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "OK",
		"data":    updatedEmote,
	})

}
