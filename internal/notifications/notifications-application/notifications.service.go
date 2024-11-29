package Notificationspplication

import (
	notificationsdomain "PINKKER-BACKEND/internal/notifications/notifications"
	Notificationstinfrastructure "PINKKER-BACKEND/internal/notifications/notifications-infrastructure"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationsService struct {
	NotificationsRepository *Notificationstinfrastructure.NotificationsRepository
}

func NewNotificationsService(repo *Notificationstinfrastructure.NotificationsRepository) *NotificationsService {
	return &NotificationsService{
		NotificationsRepository: repo,
	}
}

func (s *NotificationsService) GetRecentNotifications(userID primitive.ObjectID, page int) ([]notificationsdomain.Notification, error) {
	return s.NotificationsRepository.GetRecentNotifications(userID, page, 10)
}
func (s *NotificationsService) GetOldNotifications(userID primitive.ObjectID, page int) ([]notificationsdomain.Notification, error) {
	return s.NotificationsRepository.GetOldNotifications(userID, page, 10)
}
