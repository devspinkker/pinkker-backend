package subscriptionapplication

import (
	subscriptiondomain "PINKKER-BACKEND/internal/subscription/subscription-domain"
	subscriptioninfrastructure "PINKKER-BACKEND/internal/subscription/subscription-infrastructure"

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

func (s *SubscriptionService) Subscription(FromUser, ToUser primitive.ObjectID, text string) error {
	err := s.roomRepository.Subscription(FromUser, ToUser, text)
	return err
}
func (D *SubscriptionService) GetSubsChat(id primitive.ObjectID) ([]subscriptiondomain.ResSubscriber, error) {
	Donations, err := D.roomRepository.GetSubsChat(id)

	return Donations, err
}
func (D *SubscriptionService) UpdataSubsChat(id, ToUser primitive.ObjectID) error {
	err := D.roomRepository.UpdataSubsChat(id, ToUser)

	return err
}
