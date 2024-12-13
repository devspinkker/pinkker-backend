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

func (r *NotificationsRepository) GetOldNotifications(userID primitive.ObjectID, page int, limit int) ([]notificationsdomain.NotificationRes, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	notificationsColl := db.Collection("Notifications")
	usersColl := db.Collection("Users")

	// Obtener la lista de usuarios seguidos y la última conexión del usuario
	var user struct {
		Following      map[string]interface{} `bson:"Following"`
		LastConnection primitive.DateTime     `bson:"LastConnection"`
	}
	err := usersColl.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo los datos del usuario: %v", err)
	}
	// Si el usuario no sigue a nadie, retornar lista vacía
	if len(user.Following) == 0 {
		return []notificationsdomain.NotificationRes{}, nil
	}

	// Convertir el mapa de Following a un conjunto de IDs
	followingSet := make(map[primitive.ObjectID]bool)
	for key := range user.Following {
		id, err := primitive.ObjectIDFromHex(key)
		if err == nil {
			followingSet[id] = true
		}
	}

	skip := (page - 1) * limit

	// Pipeline de agregación
	pipeline := bson.A{
		bson.M{"$match": bson.M{"userId": userID}},
		bson.M{"$unwind": "$notifications"},
		bson.M{"$match": bson.M{
			"notifications.timestamp": bson.M{"$lt": user.LastConnection},
		}},
		bson.M{"$sort": bson.M{"notifications.timestamp": -1}},
		bson.M{"$skip": skip},
		bson.M{"$limit": limit},
		bson.M{"$project": bson.M{
			"type":      "$notifications.type",
			"nameuser":  "$notifications.nameuser",
			"avatar":    "$notifications.avatar",
			"text":      "$notifications.text",
			"pixeles":   "$notifications.pixeles",
			"timestamp": "$notifications.timestamp",
			"idUser":    "$notifications.idUser",
		}},
	}

	// Ejecutar el pipeline de agregación
	cursor, err := notificationsColl.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando el pipeline: %v", err)
	}
	defer cursor.Close(context.Background())

	// Decodificar resultados y verificar si el idUser está en la lista de Following
	var notifications []notificationsdomain.NotificationRes
	for cursor.Next(context.Background()) {
		var notification notificationsdomain.NotificationRes
		if err := cursor.Decode(&notification); err != nil {
			return nil, fmt.Errorf("error decodificando notificación: %v", err)
		}

		// Marcar si el usuario está seguido
		notification.IsFollowed = followingSet[notification.IdUser]
		notifications = append(notifications, notification)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("error del cursor: %v", err)
	}

	return notifications, nil
}

func (r *NotificationsRepository) GetRecentNotifications(userID primitive.ObjectID, page int, limit int) ([]notificationsdomain.NotificationRes, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	notificationsColl := db.Collection("Notifications")
	usersColl := db.Collection("Users")

	// Obtener la lista de usuarios seguidos
	var user struct {
		Following map[string]interface{} `bson:"Following"`
	}
	err := usersColl.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo los datos del usuario: %v", err)
	}

	// Si el usuario no sigue a nadie, retornar lista vacía
	if len(user.Following) == 0 {
		return []notificationsdomain.NotificationRes{}, nil
	}

	// Convertir el mapa de Following a un conjunto de IDs
	followingSet := make(map[primitive.ObjectID]bool)
	for key := range user.Following {
		id, err := primitive.ObjectIDFromHex(key)
		if err == nil {
			followingSet[id] = true
		}
	}

	skip := (page - 1) * limit

	// Pipeline de agregación
	pipeline := bson.A{
		bson.M{"$match": bson.M{"userId": userID}},
		bson.M{"$unwind": "$notifications"},
		bson.M{"$match": bson.M{
			"$expr": bson.M{
				"$gt": bson.A{"$notifications.timestamp", "$userInfo.LastConnection"},
			},
		}},
		bson.M{"$sort": bson.M{"notifications.timestamp": -1}},
		bson.M{"$skip": skip},
		bson.M{"$limit": limit},
		bson.M{"$project": bson.M{
			"type":      "$notifications.type",
			"nameuser":  "$notifications.nameuser",
			"avatar":    "$notifications.avatar",
			"text":      "$notifications.text",
			"pixeles":   "$notifications.pixeles",
			"timestamp": "$notifications.timestamp",
			"idUser":    "$notifications.idUser",
		}},
	}

	// Ejecutar el pipeline de agregación
	cursor, err := notificationsColl.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando el pipeline: %v", err)
	}
	defer cursor.Close(context.Background())

	// Decodificar resultados y verificar si el idUser está en la lista de Following
	var notifications []notificationsdomain.NotificationRes
	for cursor.Next(context.Background()) {
		var notification notificationsdomain.NotificationRes
		if err := cursor.Decode(&notification); err != nil {
			return nil, fmt.Errorf("error decodificando notificación: %v", err)
		}

		// Marcar si el usuario está seguido
		notification.IsFollowed = followingSet[notification.IdUser]
		notifications = append(notifications, notification)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("error del cursor: %v", err)
	}

	return notifications, nil
}
