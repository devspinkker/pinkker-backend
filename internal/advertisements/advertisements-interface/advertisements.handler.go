package advertisementsinterface

import (
	"PINKKER-BACKEND/internal/advertisements/advertisements"
	advertisementsapplication "PINKKER-BACKEND/internal/advertisements/advertisements-application"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdvertisementsRepository struct {
	Servise *advertisementsapplication.AdvertisementsService
}

func NewwithdrawService(Servise *advertisementsapplication.AdvertisementsService) *AdvertisementsRepository {
	return &AdvertisementsRepository{
		Servise: Servise,
	}
}

func (s *AdvertisementsRepository) BuyadCreate(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	nameUser := c.Context().UserValue("nameUser").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var req advertisements.UpdateAdvertisement

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}
	req.NameUser = nameUser
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}

	advertisementsGet, err := s.Servise.BuyadCreate(idValueObj, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    advertisementsGet,
	})
}

func (s *AdvertisementsRepository) GetAdvertisements(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	var req advertisements.AdvertisementGet

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}

	page, err := strconv.ParseInt(c.Query("page", "1"), 10, 64)
	if err != nil || page < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "InvalidPageNumber",
		})
	}

	advertisementsGet, err := s.Servise.GetAdvertisements(idValueObj, req, page)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    advertisementsGet,
	})
}
func (s *AdvertisementsRepository) AcceptPendingAds(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	var req advertisements.AdvertisementGet

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}

	nameUser := c.Query("nameUser", "1")

	err := s.Servise.AcceptPendingAds(idValueObj, req, nameUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
func (s *AdvertisementsRepository) GetAllPendingAds(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	var req advertisements.AdvertisementGet

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}

	page, err := strconv.ParseInt(c.Query("page", "1"), 10, 64)
	if err != nil || page < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "InvalidPageNumber",
		})
	}

	pendings, err := s.Servise.GetAllPendingAds(idValueObj, req, page)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    pendings,
	})
}
func (s *AdvertisementsRepository) RemovePendingAds(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	var req advertisements.AdvertisementGet

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}

	nameUser := c.Query("nameUser", "1")
	err := s.Servise.RemovePendingAds(idValueObj, req, nameUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func (s *AdvertisementsRepository) GetAdsUser(c *fiber.Ctx) error {
	nameUser := c.Context().UserValue("nameUser").(string)

	advertisementsGet, err := s.Servise.GetAdsUser(nameUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    advertisementsGet,
	})
}
func (s *AdvertisementsRepository) GetAdsUserCode(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var req advertisements.GetAdsUserCode

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}
	advertisementsGet, err := s.Servise.GetAdsUserCode(req, idValueObj)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    advertisementsGet,
	})
}
func (s *AdvertisementsRepository) CreateAdvertisement(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var req advertisements.UpdateAdvertisement

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}

	advertisementsGet, err := s.Servise.CreateAdvertisement(idValueObj, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    advertisementsGet,
	})
}

func (s *AdvertisementsRepository) IdOfTheUsersWhoClicked(c *fiber.Ctx) error {
	type request struct {
		IdAdvertisements primitive.ObjectID `json:"idAdvertisements" `
	}
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var req request

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}

	err := s.Servise.IdOfTheUsersWhoClicked(idValueObj, req.IdAdvertisements)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

func (s *AdvertisementsRepository) UpdateAdvertisement(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var req advertisements.UpdateAdvertisement

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}

	advertisementsGet, err := s.Servise.UpdateAdvertisement(idValueObj, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    advertisementsGet,
	})
}
func (s *AdvertisementsRepository) DeleteAdvertisement(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var req advertisements.DeleteAdvertisement

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"err":     err.Error(),
		})
	}
	err := s.Servise.DeleteAdvertisement(idValueObj, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
