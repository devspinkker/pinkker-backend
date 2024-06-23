package Chatsapplication

import (
	Chatsdomain "PINKKER-BACKEND/internal/Chat/Chats"
	Chatsinfrastructure "PINKKER-BACKEND/internal/Chat/Chats-infrastructure"
	"time"
)

type ChatsService struct {
	ChatsRepository *Chatsinfrastructure.ChatsRepository
}

func NewChatsService(ChatsRepository *Chatsinfrastructure.ChatsRepository) *ChatsService {
	return &ChatsService{
		ChatsRepository: ChatsRepository,
	}
}

func (s *ChatsService) SendMessage(senderID, receiverID, content string) (*Chatsdomain.Message, string, error) {
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
		return message, "", err
	}

	id, err := s.ChatsRepository.AddMessageToChat(senderID, receiverID, savedMessage.ID)
	if err != nil {
		return message, id, err
	}

	return message, id, nil
}

func (s *ChatsService) GetMessages(senderID, receiverID string) ([]*Chatsdomain.Message, error) {
	return s.ChatsRepository.GetMessages(senderID, receiverID)
}

func (s *ChatsService) MarkMessageAsSeen(messageID string) error {
	return s.ChatsRepository.MarkMessageAsSeen(messageID)
}

func (s *ChatsService) GetMessageByID(messageID string) (*Chatsdomain.Message, error) {
	return s.ChatsRepository.GetMessageByID(messageID)
}
