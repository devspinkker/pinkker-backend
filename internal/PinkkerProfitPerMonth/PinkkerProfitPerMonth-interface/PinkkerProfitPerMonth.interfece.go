package PinkkerProfitPerMonthinterfaces

import (
	PinkkerProfitPerMonthapplication "PINKKER-BACKEND/internal/PinkkerProfitPerMonth/PinkkerProfitPerMonth-application"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PinkkerProfitPerMonthHandler struct {
	PinkkerProfitPerMonthServise *PinkkerProfitPerMonthapplication.PinkkerProfitPerMonthService
}

func NewPinkkerProfitPerMonthService(PinkkerProfitPerMonthServise *PinkkerProfitPerMonthapplication.PinkkerProfitPerMonthService) *PinkkerProfitPerMonthHandler {
	return &PinkkerProfitPerMonthHandler{
		PinkkerProfitPerMonthServise: PinkkerProfitPerMonthServise,
	}
}
func (h *PinkkerProfitPerMonthHandler) GetEarningsByWeek(c *fiber.Ctx) error {
	weekStr := c.Query("week")
	code := c.Query("code")

	idValue := c.Context().UserValue("_id").(string)
	streamerID, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	// Parse the week start date from the query parameter
	week, err := time.Parse("2006-01-02", weekStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid week format, use YYYY-MM-DD",
		})
	}

	earnings, err := h.PinkkerProfitPerMonthServise.GetProfitByMonth(streamerID, code, week)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get earnings by week",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    earnings,
	})
}
func (h *PinkkerProfitPerMonthHandler) GetEarningsByMonthRange(c *fiber.Ctx) error {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	code := c.Query("code")

	idValue := c.Context().UserValue("_id").(string)
	streamerID, errorID := primitive.ObjectIDFromHex(idValue)
	if errorID != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid start date format, use YYYY-MM-DD",
		})
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid end date format, use YYYY-MM-DD",
		})
	}

	earnings, err := h.PinkkerProfitPerMonthServise.GetProfitByMonthRange(streamerID, code, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get earnings by month range",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    earnings,
	})
}
