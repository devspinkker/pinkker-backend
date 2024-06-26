package Chatsapplication

import (
	Chatsdomain "PINKKER-BACKEND/internal/Chat/Chats"
	Chatsinfrastructure "PINKKER-BACKEND/internal/Chat/Chats-infrastructure"
	"time"

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

func (s *ChatsService) SendMessage(senderID, receiverID primitive.ObjectID, content string) (*Chatsdomain.Message, primitive.ObjectID, error) {
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
		return message, primitive.ObjectID{}, err
	}

	id, err := s.ChatsRepository.AddMessageToChat(senderID, receiverID, savedMessage.ID)
	if err != nil {
		return message, id, err
	}

	return message, id, nil
}

func (s *ChatsService) GetMessages(senderID, receiverID primitive.ObjectID) ([]*Chatsdomain.Message, error) {
	return s.ChatsRepository.GetMessages(senderID, receiverID)
}

func (s *ChatsService) MarkMessageAsSeen(messageID string) error {
	return s.ChatsRepository.MarkMessageAsSeen(messageID)
}
func (s *ChatsService) GetChatsByUserID(messageID primitive.ObjectID) ([]*Chatsdomain.ChatWithUsers, error) {
	return s.ChatsRepository.GetChatsByUserID(messageID)
}

func (s *ChatsService) GetMessageByID(messageID string) (*Chatsdomain.Message, error) {
	return s.ChatsRepository.GetMessageByID(messageID)
}
func (s *ChatsService) CreateChatOrGetChats(Idtoken, userId primitive.ObjectID) (*Chatsdomain.ChatWithUsers, error) {
	return s.ChatsRepository.CreateChatOrGetChats(Idtoken, userId)
}
