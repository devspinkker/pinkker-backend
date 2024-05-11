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
	user, errdonatePixels := h.subscriptionService.Subscription(FromUser, idReq.ToUser, idReq.Text)
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
	errdDeleteRedisUserChatInOneRoom := h.subscriptionService.DeleteRedisUserChatInOneRoom(FromUser, idReq.ToUser)
	if errdDeleteRedisUserChatInOneRoom != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": errupdataSubsChat,
		})
	}
	h.NotifyActivityFeed(FromUser.Hex()+"ActivityFeed", user, idReq.Text)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
func (h *SubscriptionHandler) NotifyActivityFeed(room, user string, text string) error {
	clients, err := h.subscriptionService.GetWebSocketActivityFeed(room)
	if err != nil {
		return err
	}

	notification := map[string]interface{}{
		"action":  "Suscribirse",
		"Pixeles": text,
		"data":    user,
	}
	for _, client := range clients {
		err = client.WriteJSON(notification)
		if err != nil {
			return err
		}
	}

	return nil
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

type ReqGetSubsAct struct {
	Source primitive.ObjectID `json:"-" query:"Source"`
	Desti  primitive.ObjectID `json:"-" query:"Desti"`
}

func (h *SubscriptionHandler) GetSubsAct(c *fiber.Ctx) error {

	var idReq ReqGetSubsAct
	if err := c.QueryParser(&idReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "StatusBadRequest",
		})
	}
	sub, err := h.subscriptionService.GetSubsAct(idReq.Source, idReq.Desti)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":    sub,
		"message": "ok",
	})
}
