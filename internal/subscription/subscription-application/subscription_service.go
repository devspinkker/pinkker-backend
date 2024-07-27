package subscriptionapplication

import (
	subscriptiondomain "PINKKER-BACKEND/internal/subscription/subscription-domain"
	subscriptioninfrastructure "PINKKER-BACKEND/internal/subscription/subscription-infrastructure"

	"github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubscriptionService struct {
	roomRepository *subscriptioninfrastructure.SubscriptionRepository
}

func NewChatService(roomRepository *subscriptioninfrastructure.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		roomRepository: roomRepository,
	}
}

func (s *SubscriptionService) GetWebSocketActivityFeed(user string) ([]*websocket.Conn, error) {
	client, err := s.roomRepository.GetWebSocketClientsInRoom(user)
	return client, err
}
func (s *SubscriptionService) Subscription(FromUser, ToUser primitive.ObjectID, text string) (string, error) {
	user, err := s.roomRepository.Subscription(FromUser, ToUser, text)
	return user, err
}
func (D *SubscriptionService) GetSubsChat(id primitive.ObjectID) ([]subscriptiondomain.ResSubscriber, error) {
	Donations, err := D.roomRepository.GetSubsChat(id)

	return Donations, err
}
func (D *SubscriptionService) PublishNotification(roomID primitive.ObjectID, noty map[string]interface{}) error {
	stream, err := D.roomRepository.GetStreamByStreamerID(roomID)
	if err != nil {
		return err
	}
	return D.roomRepository.PublishNotification(stream.ID.Hex(), noty)

}
func (D *SubscriptionService) UpdataSubsChat(id, ToUser primitive.ObjectID) error {
	err := D.roomRepository.UpdataSubsChat(id, ToUser)

	return err
}

func (s *SubscriptionService) GetSubsAct(Source, Destination primitive.ObjectID) (subscriptiondomain.Subscription, error) {
	subs, err := s.roomRepository.GetSubsAct(Source, Destination)
	return subs, err
}

func (u *SubscriptionService) DeleteRedisUserChatInOneRoom(userToDelete primitive.ObjectID, IdRoom primitive.ObjectID) error {
	err := u.roomRepository.DeleteRedisUserChatInOneRoom(userToDelete, IdRoom)

	return err
}

func (u *SubscriptionService) StateTheUserInChat(Donado primitive.ObjectID, Donante primitive.ObjectID) (bool, error) {
	return u.roomRepository.StateTheUserInChat(Donado, Donante)

}
