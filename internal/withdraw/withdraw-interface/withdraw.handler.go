package withdrawtinterfaces

import (
	withdrawalsdomain "PINKKER-BACKEND/internal/withdraw/withdraw"
	withdrawapplication "PINKKER-BACKEND/internal/withdraw/withdraw-application"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WithdrawalsRepository struct {
	Servise *withdrawapplication.WithdrawalsService
}

func NewwithdrawService(Servise *withdrawapplication.WithdrawalsService) *WithdrawalsRepository {
	return &WithdrawalsRepository{
		Servise: Servise,
	}
}

func (s *WithdrawalsRepository) WithdrawalRequest(c *fiber.Ctx) error {
	idValue := c.Context().UserValue("_id").(string)
	idValueObj, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	var req withdrawalsdomain.WithdrawalRequestReq

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	err := s.Servise.WithdrawalRequest(idValueObj, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
