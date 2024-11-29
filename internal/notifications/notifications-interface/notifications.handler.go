package Notificationstinterfaces

import (
	Notificationspplication "PINKKER-BACKEND/internal/notifications/notifications-application"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationsHandler struct {
	Service *Notificationspplication.NotificationsService
}

func NewNotificationsHandler(service *Notificationspplication.NotificationsService) *NotificationsHandler {
	return &NotificationsHandler{
		Service: service,
	}
}

func (h *NotificationsHandler) GetRecentNotifications(c *fiber.Ctx) error {
	IdUserToken := c.Context().UserValue("_id").(string)

	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errinObjectID.Error(),
		})
	}

	page := c.QueryInt("page", 1)

	notifications, err := h.Service.GetRecentNotifications(IdUserTokenP, page)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"notifications": notifications})
}
func (h *NotificationsHandler) GetOldNotifications(c *fiber.Ctx) error {
	IdUserToken := c.Context().UserValue("_id").(string)

	IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"data":    errinObjectID.Error(),
		})
	}

	page := c.QueryInt("page", 1)

	notifications, err := h.Service.GetOldNotifications(IdUserTokenP, page)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"notifications": notifications})
}
