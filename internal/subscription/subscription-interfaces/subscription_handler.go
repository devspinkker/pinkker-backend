package subscriptioninterfaces

import (
	subscriptionapplication "PINKKER-BACKEND/internal/subscription/subscription-application"
	subscriptiondomain "PINKKER-BACKEND/internal/subscription/subscription-domain"
	"PINKKER-BACKEND/pkg/helpers"
	"fmt"
	"time"

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
	if FromUser == idReq.ToUser {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "StatusBadRequest",
			"data":    "toUser !== ",
		})
	}

	banned, err := h.subscriptionService.StateTheUserInChat(idReq.ToUser, FromUser)
	if err != nil || banned {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message": "StatusConflict",
			"data":    "baneado",
		})
	}

	user, avatar, errdonatePixels := h.subscriptionService.Subscription(FromUser, idReq.ToUser, idReq.Text)
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
	IsFollowing, _ := h.subscriptionService.IsFollowing(FromUser, idReq.ToUser)
	h.NotifyActivityFeed(idReq.ToUser.Hex()+"ActivityFeed", user, avatar, idReq.Text, IsFollowing)
	h.NotifyActivityToChat(idReq.ToUser, user, idReq.Text, FromUser)
	Notification := helpers.CreateNotification("Suscribirse", user, avatar, idReq.Text, 0, FromUser)
	err = h.subscriptionService.SaveNotification(idReq.ToUser, Notification)
	if err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "ok",
			"data":    "SaveNotification error",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
	})
}
func (h *SubscriptionHandler) NotifyActivityFeed(room, user, Avatar, text string, IsFollowing bool) error {
	clients, err := h.subscriptionService.GetWebSocketActivityFeed(room)
	if err != nil {
		return err

	}

	notification := map[string]interface{}{
		"type":        "Suscribirse",
		"nameuser":    user,
		"Text":        text,
		"avatar":      Avatar,
		"isFollowing": IsFollowing,
		"timestamp":   time.Now(),
	}

	for _, client := range clients {
		err = client.WriteJSON(notification)
		if err != nil {
			return err
		}
	}

	return nil
}
func (h *SubscriptionHandler) NotifyActivityToChat(UserToken primitive.ObjectID, user string, text string, id primitive.ObjectID) error {

	notification := map[string]interface{}{
		"action": "Subs",
		"Text":   text,
		"data":   user,
		"IdUser": id,
	}
	h.subscriptionService.PublishNotification(UserToken, notification)
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
