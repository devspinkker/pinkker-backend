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

func (h *ChatsHandler) SendMessage(c *fiber.Ctx) error {
	userID := c.Context().UserValue("_id").(primitive.ObjectID)

	var request struct {
		SenderID   primitive.ObjectID `json:"sender_id"`
		ReceiverID primitive.ObjectID `json:"receiver_id"`
		Content    string             `json:"content"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse request"})
	}

	if request.SenderID != userID && request.ReceiverID != userID {
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
	userID := c.Context().UserValue("_id").(primitive.ObjectID)

	senderID := c.Query("sender_id")
	receiverID := c.Query("receiver_id")

	// Verify that the sender or receiver is the authenticated user
	if senderID != userID.Hex() && receiverID != userID.Hex() {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	messages, err := h.Service.GetMessages(senderID, receiverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot get messages"})
	}

	return c.Status(fiber.StatusOK).JSON(messages)
}

func (h *ChatsHandler) GetChatsByUserID(c *fiber.Ctx) error {
	userID := c.Context().UserValue("_id").(primitive.ObjectID)

	messages, err := h.Service.GetChatsByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot get messages"})
	}

	return c.Status(fiber.StatusOK).JSON(messages)
}

func (h *ChatsHandler) MarkMessageAsSeen(c *fiber.Ctx) error {
	userID := c.Context().UserValue("_id").(primitive.ObjectID)

	messageID := c.Params("id")

	message, err := h.Service.GetMessageByID(messageID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot retrieve message"})
	}

	// Verify that the receiver is the authenticated user
	if message.ReceiverID != userID {
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
		idUser, _ = primitive.ObjectIDFromHex(id)
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
