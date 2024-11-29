package helpers

import (
	notificationsdomain "PINKKER-BACKEND/internal/notifications/notifications"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateNotification(notificationType string, user string, avatar string, text string, pixeles float64, IdUser primitive.ObjectID) notificationsdomain.Notification {
	// Crear la base de la notificación
	notification := notificationsdomain.Notification{
		Type:      notificationType,
		NameUser:  user,
		Avatar:    avatar,
		Timestamp: time.Now(),
		Text:      text,
		Pixeles:   pixeles,
		IdUser:    IdUser,
	}

	return notification
}
