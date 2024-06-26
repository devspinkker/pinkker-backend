package Chatsinterface

import (
	Chatsdomain "PINKKER-BACKEND/internal/Chat/Chats"
	Chatsapplication "PINKKER-BACKEND/internal/Chat/Chats-application"
	"PINKKER-BACKEND/pkg/jwt"
	"PINKKER-BACKEND/pkg/utils"
	"encoding/json"

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

func (h *ChatsHandler) SendMessage(c *fiber.Ctx) error {
	userID := c.Context().UserValue("_id").(string)

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user ID"})
	}
	var request struct {
		SenderID   primitive.ObjectID `json:"sender_id"`
		ReceiverID primitive.ObjectID `json:"receiver_id"`
		Content    string             `json:"content"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse request"})
	}

	if request.SenderID != objID && request.ReceiverID != objID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	message, Room, err := h.Service.SendMessage(request.SenderID, request.ReceiverID, request.Content)
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

	var senderID, receiverID primitive.ObjectID
	if senderIDHex := c.Query("sender_id"); senderIDHex != "" {
		senderID, err = primitive.ObjectIDFromHex(senderIDHex)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid sender_id"})
		}
	}

	if receiverIDHex := c.Query("receiver_id"); receiverIDHex != "" {
		receiverID, err = primitive.ObjectIDFromHex(receiverIDHex)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid receiver_id"})
		}
	}

	if senderID != objID && receiverID != objID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	messages, err := h.Service.GetMessages(senderID, receiverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot get messages"})
	}

	return c.Status(fiber.StatusOK).JSON(messages)
}

func (h *ChatsHandler) GetChatsByUserID(c *fiber.Ctx) error {
	userID := c.Context().UserValue("_id").(string)

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user ID"})
	}

	messages, err := h.Service.GetChatsByUserID(objID)
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
