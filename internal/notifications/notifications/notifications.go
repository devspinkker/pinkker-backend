package notificationsdomain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	Type      string             `bson:"type" json:"type"`                           // Tipo de notificación (e.g., follow, DonatePixels, etc.)
	NameUser  string             `bson:"nameuser" json:"nameuser"`                   // Nombre del usuario que genera la notificación
	IdUser    primitive.ObjectID `bson:"idUser" json:"idUser"`                       // Nombre del usuario que genera la notificación
	Avatar    string             `bson:"avatar" json:"avatar"`                       // URL del avatar del usuario
	Text      string             `bson:"text,omitempty" json:"text,omitempty"`       // Texto adicional (opcional)
	Pixeles   float64            `bson:"pixeles,omitempty" json:"pixeles,omitempty"` // Cantidad de píxeles (opcional)
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`                 // Marca de tiempo de la notificación
}

// UserNotifications representa el documento de notificaciones de un usuario.
type UserNotifications struct {
	UserID        primitive.ObjectID `bson:"userId" json:"userId"`               // ID del usuario propietario de las notificaciones
	Notifications []Notification     `bson:"notifications" json:"notifications"` // Lista de notificaciones
}
