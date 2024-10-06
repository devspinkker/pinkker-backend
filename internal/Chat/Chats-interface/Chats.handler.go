package Chatsinterface

import (
	Chatsdomain "PINKKER-BACKEND/internal/Chat/Chats"
	Chatsapplication "PINKKER-BACKEND/internal/Chat/Chats-application"
	"PINKKER-BACKEND/pkg/jwt"
	"PINKKER-BACKEND/pkg/utils"
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatsHandler struct {
	Service *Chatsapplication.ChatsService
}

func NewChatsHandler(service *Chatsapplication.ChatsService) *ChatsHandler {
	return &ChatsHandler{Service: service}
}

func (h *ChatsHandler) UpdateUserStatus(c *fiber.Ctx) error {
	userID := c.Context().UserValue("_id").(string)

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user ID"})
	}
	var request struct {
		Chatid  primitive.ObjectID `json:"chatid"`
		Content string             `json:"content"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse request"})
	}
	if err := h.Service.UpdateUserStatus(objID, request.Chatid, request.Content); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "StatusInternalServerError",
			"error":   err,
		})

	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "StatusOK",
	})
}

func (h *ChatsHandler) CreateChatOrGetChats(c *fiber.Ctx) error {
	// Obtener el ID del usuario actual desde el contexto
	userID := c.Context().UserValue("_id").(string)

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user ID"})
	}

	var body struct {
		OtherUserID string `json:"other_user_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	otherUserObjID, err := primitive.ObjectIDFromHex(body.OtherUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid other user ID"})
	}

	chat, err := h.Service.CreateChatOrGetChats(userObjID, otherUserObjID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot create chat"})
	}

	return c.Status(fiber.StatusOK).JSON(chat)
}
func (h *ChatsHandler) DeleteAllMessages(c *fiber.Ctx) error {
	userID := c.Context().UserValue("_id").(string)

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user ID"})
	}
	var request struct {
		ReceiverID primitive.ObjectID `json:"userid"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse request"})
	}

	err = h.Service.DeleteAllMessages(objID, request.ReceiverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "StatusOK"})
}

func (h *ChatsHandler) UpdateChatBlockStatus(c *fiber.Ctx) error {
	userID := c.Context().UserValue("_id").(string)

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user ID"})
	}
	var request struct {
		ChatID      primitive.ObjectID `json:"chatID"`
		BlockStatus bool               `json:"blockStatus"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse request"})
	}

	err = h.Service.UpdateChatBlockStatus(request.ChatID, objID, request.BlockStatus)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "StatusOK"})
}

func (h *ChatsHandler) SendMessage(c *fiber.Ctx) error {
	userID := c.Context().UserValue("_id").(string)

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user ID"})
	}
	var request struct {
		ReceiverID primitive.ObjectID `json:"chatid"`
		Content    string             `json:"content"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse request"})
	}

	message, Room, err := h.Service.SendMessage(objID, request.ReceiverID, request.Content)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	notification := map[string]interface{}{
		"action":  "new_message",
		"message": message,
	}

	clients, err := utils.NewChatService().GetWebSocketClientsInRoom(Room.Hex())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get WebSocket clients"})
	}
	connectedClientsCount := len(clients)
	if connectedClientsCount == 0 {
		err = h.Service.UpdateNotificationFlag(Room, primitive.ObjectID{})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "UpdateNotificationFlag"})
		}
	} else {
		err = h.Service.UpdateNotificationFlag(Room, request.ReceiverID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "UpdateNotificationFlag"})
		}
	}
	for _, client := range clients {
		err = client.WriteJSON(notification)
		if err != nil {
			return err
		}
	}

	return c.Status(fiber.StatusOK).JSON(message)
}

func (h *ChatsHandler) GetMessages(c *fiber.Ctx) error {
	userID := c.Context().UserValue("_id").(string)

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user ID"})
	}

	var receiverID primitive.ObjectID

	if receiverIDHex := c.Query("receiver_id"); receiverIDHex != "" {
		receiverID, err = primitive.ObjectIDFromHex(receiverIDHex)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid receiver_id"})
		}
	}
	messages, err := h.Service.GetMessages(objID, receiverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot get messages"})
	}

	return c.Status(fiber.StatusOK).JSON(messages)
}

func (h *ChatsHandler) GetChatsByUserIDWithStatus(c *fiber.Ctx) error {
	userID := c.Context().UserValue("_id").(string)

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user ID"})
	}

	status := c.Query("status", "primary")

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid page parameter"})
	}
	messages, err := h.Service.GetChatsByUserIDWithStatus(objID, page, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot get messages"})
	}

	return c.Status(fiber.StatusOK).JSON(messages)
}

func (h *ChatsHandler) MarkMessageAsSeen(c *fiber.Ctx) error {
	userID := c.Context().UserValue("_id").(string)

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "StatusBadRequest"})
	}
	messageID := c.Params("id")

	message, err := h.Service.GetMessageByID(messageID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot retrieve message"})
	}

	// Verify that the receiver is the authenticated user
	if message.ReceiverID != objID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	if err := h.Service.MarkMessageAsSeen(messageID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot mark message as seen"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "message marked as seen"})
}

func (h *ChatsHandler) WebSocketHandler(c *websocket.Conn) {
	token := c.Params("token", "null")
	var idUser primitive.ObjectID
	if token != "null" {
		_, id, _, err := jwt.ExtractDataFromToken(token)
		if err != nil {
			return
		}
		idUser, err = primitive.ObjectIDFromHex(id)
		if err != nil {
			return
		}
	}

	roomID := c.Params("roomID")
	client := &utils.Client{Connection: c}
	chatService := utils.NewChatService()
	chatService.AddClientToRoom(roomID, client)

	defer func() {
		chatService.RemoveClientFromRoom(roomID, client)
		_ = c.Close()
	}()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}

		// Procesar el mensaje recibido
		var message Chatsdomain.Message
		if err := json.Unmarshal(msg, &message); err == nil {

			if message.ReceiverID == idUser {

				notification := map[string]interface{}{
					"action":  "new_message_received",
					"message": message,
				}
				err = client.Connection.WriteJSON(notification)
				if err != nil {
					chatService.RemoveClientFromRoom(roomID, client)
					_ = c.Close()
				}
			}
		}
	}
}
