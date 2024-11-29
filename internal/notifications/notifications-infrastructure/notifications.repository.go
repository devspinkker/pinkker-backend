package Notificationstinfrastructure

import (
	notificationsdomain "PINKKER-BACKEND/internal/notifications/notifications"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type NotificationsRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewNotificationsRepository(redisClient *redis.Client, mongoClient *mongo.Client) *NotificationsRepository {
	return &NotificationsRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}

func (r *NotificationsRepository) GetRecentNotifications(userID primitive.ObjectID, page int, limit int) ([]notificationsdomain.Notification, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Notifications")

	skip := (page - 1) * limit

	// Pipeline de agregación
	pipeline := bson.A{
		// 1. Filtrar por el usuario en la colección de notificaciones
		bson.M{"$match": bson.M{"userId": userID}},
		// 2. Lookup para obtener la última conexión del usuario desde la colección Users
		bson.M{"$lookup": bson.M{
			"from":         "Users",    // Colección Users
			"localField":   "userId",   // Campo userId en Notifications
			"foreignField": "_id",      // Campo _id en Users
			"as":           "userInfo", // Campo para almacenar la información del usuario
		}},
		// 3. Descomponer el array de userInfo
		bson.M{"$unwind": "$userInfo"},
		// 4. Descomponer el array de notificaciones
		bson.M{"$unwind": "$notifications"},
		// 5. Filtrar las notificaciones cuyo timestamp sea mayor que LastConnection
		bson.M{"$match": bson.M{
			"$expr": bson.M{
				"$gt": bson.A{"$notifications.timestamp", "$userInfo.LastConnection"},
			},
		}},
		// 6. Ordenar las notificaciones por timestamp en orden descendente
		bson.M{"$sort": bson.M{"notifications.timestamp": -1}},
		// 7. Paginación: saltar y limitar
		bson.M{"$skip": skip},
		bson.M{"$limit": limit},
		// 8. Proyectar los campos finales
		bson.M{"$project": bson.M{
			"type":      "$notifications.type",
			"nameuser":  "$notifications.nameuser",
			"avatar":    "$notifications.avatar",
			"text":      "$notifications.text",
			"pixeles":   "$notifications.pixeles",
			"timestamp": "$notifications.timestamp",
		}},
	}

	// Ejecutar la consulta
	cursor, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, fmt.Errorf("error al obtener las notificaciones: %v", err)
	}
	defer cursor.Close(context.Background())

	// Decodificar resultados
	var notifications []notificationsdomain.Notification
	for cursor.Next(context.Background()) {
		var notification notificationsdomain.Notification
		if err := cursor.Decode(&notification); err != nil {
			return nil, fmt.Errorf("error al decodificar una notificación: %v", err)
		}
		notifications = append(notifications, notification)
	}

	// Verificar si hubo error en el cursor
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("error en el cursor de MongoDB: %v", err)
	}

	return notifications, nil
}
func (r *NotificationsRepository) GetOldNotifications(userID primitive.ObjectID, page int, limit int) ([]notificationsdomain.Notification, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	notificationsColl := db.Collection("Notifications")

	skip := (page - 1) * limit

	// Pipeline de agregación
	pipeline := bson.A{
		// 1. Lookup para obtener la información del usuario (LastConnection)
		bson.M{"$lookup": bson.M{
			"from":         "Users",
			"localField":   "userId",
			"foreignField": "_id",
			"as":           "userInfo",
		}},
		// 2. Descomponer el array del usuario
		bson.M{"$unwind": bson.M{
			"path":                       "$userInfo",
			"preserveNullAndEmptyArrays": false,
		}},
		// 3. Filtrar notificaciones con timestamp ANTERIOR a la última conexión
		bson.M{"$match": bson.M{
			"userId": userID, // Filtro para el usuario actual
			"$expr": bson.M{
				"$lt": bson.A{"$notifications.timestamp", "$userInfo.LastConnection"},
			},
		}},
		// 4. Descomponer las notificaciones en documentos individuales
		bson.M{"$unwind": "$notifications"},
		// 5. Ordenar las notificaciones por fecha (más recientes primero)
		bson.M{"$sort": bson.M{"notifications.timestamp": -1}},
		// 6. Aplicar skip para paginación
		bson.M{"$skip": skip},
		// 7. Limitar el número de resultados
		bson.M{"$limit": limit},
		// 8. Proyectar los datos necesarios
		bson.M{"$project": bson.M{
			"type":      "$notifications.type",
			"nameuser":  "$notifications.nameuser",
			"avatar":    "$notifications.avatar",
			"text":      "$notifications.text",
			"pixeles":   "$notifications.pixeles",
			"timestamp": "$notifications.timestamp",
		}},
	}

	// Ejecutar el pipeline de agregación
	cursor, err := notificationsColl.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando el pipeline: %v", err)
	}
	defer cursor.Close(context.Background())

	// Decodificar resultados
	var notifications []notificationsdomain.Notification
	for cursor.Next(context.Background()) {
		var notification notificationsdomain.Notification
		if err := cursor.Decode(&notification); err != nil {
			return nil, fmt.Errorf("error decodificando notificación: %v", err)
		}
		notifications = append(notifications, notification)
	}

	// Verificar errores del cursor
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("error del cursor: %v", err)
	}

	return notifications, nil
}
