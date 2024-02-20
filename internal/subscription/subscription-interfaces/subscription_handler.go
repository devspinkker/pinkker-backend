package subscriptioninterfaces

import (
	subscriptionapplication "PINKKER-BACKEND/internal/subscription/subscription-application"
	subscriptiondomain "PINKKER-BACKEND/internal/subscription/subscription-domain"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubscriptionHandler struct {
	subscriptionService *subscriptionapplication.SubscriptionService
}

func NewSubscriptionHandler(Service *subscriptionapplication.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: Service,
	}
}

func (h *SubscriptionHandler) Suscribirse(c *fiber.Ctx) error {

	var idReq subscriptiondomain.ReqCreateSuscribirse
	if err := c.BodyParser(&idReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}

	IdUserToken := c.Context().UserValue("_id").(string)
	FromUser, errinObjectID := primitive.ObjectIDFromHex(IdUserToken)
	if errinObjectID != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
		})
	}
	if len(idReq.Text) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    "text < 100",
		})
	}
	errdonatePixels := h.subscriptionService.Subscription(FromUser, idReq.ToUser, idReq.Text)
	if errdonatePixels != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": errdonatePixels.Error(),
		})
	}

	errupdataSubsChat := h.subscriptionService.UpdataSubsChat(FromUser, idReq.ToUser)
	if errupdataSubsChat != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": errupdataSubsChat.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}

type ReqChatSubs struct {
	Id primitive.ObjectID `json:"-" query:"Toid"`
}

func (d *SubscriptionHandler) GetSubsChat(c *fiber.Ctx) error {
	var req ReqChatSubs
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
		})
	}
	donations, err := d.subscriptionService.GetSubsChat(req.Id)
	if err != nil {
		if err.Error() == "no documents found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    donations,
	})
}
