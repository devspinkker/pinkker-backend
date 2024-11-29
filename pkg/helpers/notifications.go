package helpers

import (
	notificationsdomain "PINKKER-BACKEND/internal/notifications/notifications"
	"time"
)

func CreateNotification(notificationType string, user string, avatar string, text string, pixeles float64) notificationsdomain.Notification {
	// Crear la base de la notificación
	notification := notificationsdomain.Notification{
		Type:      notificationType,
		NameUser:  user,
		Avatar:    avatar,
		Timestamp: time.Now(),
		Text:      text,
		Pixeles:   pixeles,
	}

	return notification
}
