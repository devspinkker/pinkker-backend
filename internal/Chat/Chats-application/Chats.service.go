package Chatsapplication

import (
	Chatsdomain "PINKKER-BACKEND/internal/Chat/Chats"
	Chatsinfrastructure "PINKKER-BACKEND/internal/Chat/Chats-infrastructure"
	"context"
	"errors"
	"time"

	"github.com/gofiber/websocket/v2"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatsService struct {
	ChatsRepository *Chatsinfrastructure.ChatsRepository
}

func NewChatsService(ChatsRepository *Chatsinfrastructure.ChatsRepository) *ChatsService {
	return &ChatsService{
		ChatsRepository: ChatsRepository,
	}
}
func (s *ChatsService) DeleteAllMessages(senderID, receiverID primitive.ObjectID) error {
	// Eliminar mensajes entre estos dos usuarios
	err := s.ChatsRepository.DeleteMessages(senderID, receiverID)
	if err != nil {
		return err
	}

	err = s.ChatsRepository.ClearMessageIDsInChat(senderID, receiverID)
	if err != nil {
		return err
	}

	return nil
}
func (s *ChatsService) UpdateChatBlockStatus(chatID primitive.ObjectID, userID primitive.ObjectID, blockStatus bool) error {
	// Eliminar mensajes entre estos dos usuarios
	err := s.ChatsRepository.UpdateChatBlockStatus(chatID, userID, blockStatus)
	if err != nil {
		return err
	}
	return nil
}

func (s *ChatsService) SendMessage(senderID, receiverID primitive.ObjectID, content string) (*Chatsdomain.Message, string, primitive.ObjectID, error) {

	chatId, stateUser, bloqued, err := s.ChatsRepository.IsUserBlocked(senderID, receiverID)
	if bloqued || err != nil {
		return nil, stateUser, primitive.ObjectID{}, errors.New("bloqued")
	}

	message := &Chatsdomain.Message{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		Seen:       false,
		Notified:   false,
		CreatedAt:  time.Now(),
	}

	savedMessage, err := s.ChatsRepository.SaveMessage(message)
	if err != nil {
		return message, stateUser, primitive.ObjectID{}, err
	}
	id, err := s.ChatsRepository.AddMessageToChat(savedMessage.ID, chatId)
	if err != nil {
		return message, stateUser, id, err
	}

	return message, stateUser, id, nil
}

func (s *ChatsService) GetMessages(objID, receiverID primitive.ObjectID) ([]*Chatsdomain.Message, error) {
	return s.ChatsRepository.GetMessages(objID, receiverID)
}

func (s *ChatsService) MarkMessageAsSeen(messageID string) error {
	return s.ChatsRepository.MarkMessageAsSeen(messageID)
}
func (s *ChatsService) GetChatsByUserIDWithStatus(userID primitive.ObjectID, page int, status string) ([]*Chatsdomain.ChatWithUsers, error) {
	limit := 20
	ctx := context.Background()
	if status != "primary" && status != "secondary" && status != "request" {
		return nil, errors.New("estado inválido")
	}
	return s.ChatsRepository.GetChatsByUserIDWithStatus(ctx, userID, status, page, limit)

}
func (s *ChatsService) UpdateNotificationFlag(chatID primitive.ObjectID, receiverID primitive.ObjectID) error {
	return s.ChatsRepository.UpdateNotificationFlag(chatID, receiverID)
}

func (s *ChatsService) GetMessageByID(messageID string) (*Chatsdomain.Message, error) {
	return s.ChatsRepository.GetMessageByID(messageID)
}
func (s *ChatsService) CreateChatOrGetChats(Idtoken, userId primitive.ObjectID) (*Chatsdomain.ChatWithUsers, error) {
	return s.ChatsRepository.CreateChatOrGetChats(Idtoken, userId)
}

func (s *ChatsService) UpdateUserStatus(Idtoken, chatID primitive.ObjectID, newStatus string) error {
	ctx := context.Background()
	if newStatus != "primary" && newStatus != "secondary" {
		return errors.New("estado inválido")
	}
	return s.ChatsRepository.UpdateUserStatus(ctx, chatID, Idtoken, newStatus)
}

func (u *ChatsService) GetWebSocketActivityFeed(user string) ([]*websocket.Conn, error) {
	client, err := u.ChatsRepository.GetWebSocketClientsInRoom(user)
	return client, err
}
