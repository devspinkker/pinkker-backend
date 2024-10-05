package Chatsinfrastructure

import (
	Chatsdomain "PINKKER-BACKEND/internal/Chat/Chats"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChatsRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewChatsRepository(redisClient *redis.Client, mongoClient *mongo.Client) *ChatsRepository {
	return &ChatsRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}
func (r *ChatsRepository) GetMessages(objID, user2ID primitive.ObjectID) ([]*Chatsdomain.Message, error) {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("chats")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Primero, encontrar el chat entre user1ID y user2ID
	filter := bson.M{
		"$or": []bson.M{
			{"user1_id": objID, "user2_id": user2ID},
			{"user1_id": user2ID, "user2_id": objID},
		},
	}
	var chat Chatsdomain.Chat
	err := collection.FindOne(ctx, filter).Decode(&chat)
	if err != nil {
		return nil, err
	}

	// Verificar si hay mensajes en el chat
	if len(chat.MessageIDs) == 0 {
		return []*Chatsdomain.Message{}, nil
	}

	// Obtener los últimos 20 mensajes
	var messageIDs []primitive.ObjectID
	if len(chat.MessageIDs) > 20 {
		messageIDs = chat.MessageIDs[len(chat.MessageIDs)-20:]
	} else {
		messageIDs = chat.MessageIDs
	}

	// Preparar un filtro para buscar los mensajes por sus IDs
	messageIDsAsInterface := make([]interface{}, len(messageIDs))
	for i, id := range messageIDs {
		messageIDsAsInterface[i] = id
	}

	filterMessages := bson.M{"_id": bson.M{"$in": messageIDsAsInterface}}

	// Buscar los mensajes en la colección de mensajes y ordenarlos por la fecha de creación
	messageCollection := r.mongoClient.Database("PINKKER-BACKEND").Collection("messages")
	opts := options.Find().SetSort(bson.M{"created_at": 1}) // Ordenar por fecha de creación descendente
	cursor, err := messageCollection.Find(ctx, filterMessages, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*Chatsdomain.Message
	for cursor.Next(ctx) {
		var message Chatsdomain.Message
		if err := cursor.Decode(&message); err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}
	if chat.NotifyA == objID {
		err := r.UpdateNotificationFlag(chat.ID, primitive.ObjectID{})
		if err != nil {
			return nil, err
		}
	}
	return messages, nil
}

func (r *ChatsRepository) UpdateNotificationFlag(chatID primitive.ObjectID, notifyAObj primitive.ObjectID) error {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("chats")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	notifyA := notifyAObj

	// Update the NotifyA field in the chat document
	filter := bson.M{"_id": chatID}
	update := bson.M{"$set": bson.M{"NotifyA": notifyA}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}
func (r *ChatsRepository) usersExist(user1ID, user2ID primitive.ObjectID) (*userdomain.GetUser, *userdomain.GetUser, error) {

	// Verificar si user1 existe
	filter := bson.D{
		{Key: "_id", Value: user1ID},
	}
	user1, err := r.getUserAndCheckFollow(filter, user2ID)
	if err != nil {
		return nil, nil, err
	}
	// Verificar si user2 existe
	filter = bson.D{
		{Key: "_id", Value: user2ID},
	}

	user2, err := r.getUserAndCheckFollow(filter, user1ID)

	return user1, user2, err
}

func (u *ChatsRepository) getUserAndCheckFollow(filter bson.D, id primitive.ObjectID) (*userdomain.GetUser, error) {
	GoMongoDBCollUsers := u.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	// currentTime := time.Now()

	pipeline := mongo.Pipeline{
		// Filtra el usuario basado en el filtro proporcionado
		bson.D{{Key: "$match", Value: filter}},
		// Agrega campos adicionales como FollowersCount, FollowingCount, SubscribersCount
		// Verifica si el 'id' está en las claves de 'Followers'
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "isFollowedByUser", Value: bson.D{
				{Key: "$cond", Value: bson.D{
					{Key: "if", Value: bson.D{
						{Key: "$in", Value: bson.A{id.Hex(), bson.D{
							{Key: "$map", Value: bson.D{
								{Key: "input", Value: bson.D{{Key: "$objectToArray", Value: "$Followers"}}},
								{Key: "as", Value: "follower"},
								{Key: "in", Value: "$$follower.k"}, // La clave es el ObjectID
							}},
						}}},
					}},
					{Key: "then", Value: true},
					{Key: "else", Value: false},
				}},
			}},
		}}},

		// Proyección para excluir campos innecesarios
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "Followers", Value: 0},
			{Key: "Subscribers", Value: 0},
			{Key: "SubscriptionData", Value: 0}, // Excluir los datos de lookup
		}}},
	}

	cursor, err := GoMongoDBCollUsers.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var user userdomain.GetUser
	if cursor.Next(context.Background()) {
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (r *ChatsRepository) CreateChatOrGetChats(user1ID, user2ID primitive.ObjectID) (*Chatsdomain.ChatWithUsers, error) {
	user1, user2, err := r.usersExist(user1ID, user2ID)
	if err != nil {
		return nil, fmt.Errorf("error checking users existence: %v", err)
	}

	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("chats")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verificar si ya existe un chat entre los dos usuarios
	filter := bson.M{
		"$or": []bson.M{
			{"user1_id": user1ID, "user2_id": user2ID},
			{"user1_id": user2ID, "user2_id": user1ID},
		},
	}

	var existingChat Chatsdomain.Chat
	err = collection.FindOne(ctx, filter).Decode(&existingChat)
	if err == nil {
		return r.getChatWithUsers(ctx, existingChat.ID)
	} else if err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("error finding chat: %v", err)
	}
	var StatusUser1 string = "request"
	var StatusUser2 string = "request"

	if user2.IsFollowedByUser {
		StatusUser1 = "secondary"
	}
	if user1.IsFollowedByUser {
		StatusUser2 = "secondary"
	}
	newChat := Chatsdomain.Chat{
		User1ID:     user1ID,
		User2ID:     user2ID,
		CreatedAt:   time.Now(),
		MessageIDs:  []primitive.ObjectID{},
		LastMessage: time.Now(),
		StatusUser1: StatusUser1,
		StatusUser2: StatusUser2,
	}

	result, err := collection.InsertOne(ctx, newChat)
	if err != nil {
		return nil, fmt.Errorf("error creating chat: %v", err)
	}

	insertedID := result.InsertedID.(primitive.ObjectID)

	// Devolver el chat recién creado con la información de los usuarios
	return r.getChatWithUsers(ctx, insertedID)
}

func (r *ChatsRepository) getChatWithUsers(ctx context.Context, chatID primitive.ObjectID) (*Chatsdomain.ChatWithUsers, error) {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("chats")

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.M{"_id": chatID}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "let", Value: bson.D{
				{Key: "user1_id", Value: "$user1_id"},
				{Key: "user2_id", Value: "$user2_id"},
			}},
			{Key: "pipeline", Value: bson.A{
				bson.D{{Key: "$match", Value: bson.M{
					"$expr": bson.M{
						"$or": bson.A{
							bson.M{"$eq": bson.A{"$_id", "$$user1_id"}},
							bson.M{"$eq": bson.A{"$_id", "$$user2_id"}},
						},
					},
				}}},
				bson.D{{Key: "$project", Value: bson.M{
					"_id":      1,
					"NameUser": 1,
					"Avatar":   1,
					"Partner":  1,
				}}},
			}},
			{Key: "as", Value: "users"},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var chatWithUsers Chatsdomain.ChatWithUsers
	if cursor.Next(ctx) {
		if err := cursor.Decode(&chatWithUsers); err != nil {
			return nil, err
		}
	}

	return &chatWithUsers, nil
}
func (r *ChatsRepository) SaveMessage(message *Chatsdomain.Message) (*Chatsdomain.Message, error) {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, message)
	if err != nil {
		return nil, err
	}

	insertedID := result.InsertedID.(primitive.ObjectID)

	filter := bson.M{"_id": insertedID}

	var savedMessage Chatsdomain.Message
	err = collection.FindOne(ctx, filter).Decode(&savedMessage)
	if err != nil {
		return nil, err
	}

	return &savedMessage, nil
}

func (r *ChatsRepository) AddMessageToChat(user1ID, user2ID, messageID primitive.ObjectID) (primitive.ObjectID, error) {
	_, _, err := r.usersExist(user1ID, user2ID)
	if err != nil {
		return primitive.ObjectID{}, fmt.Errorf("error checking users existence: %v", err)
	}
	// if !exist {
	// 	return primitive.ObjectID{}, fmt.Errorf("one or both users do not exist")
	// }
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("chats")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"$or": []bson.M{
		{"user1_id": user1ID, "user2_id": user2ID},
		{"user1_id": user2ID, "user2_id": user1ID},
	}}

	update := bson.M{
		"$setOnInsert": bson.M{
			"user1_id":   user1ID,
			"user2_id":   user2ID,
			"created_at": time.Now(),
		},
		"$push": bson.M{"message_ids": messageID},
		"$set":  bson.M{"LastMessage": time.Now()}, // Actualizar LastMessage a la fecha actual
	}

	opts := options.Update().SetUpsert(true)

	// Realizar la actualización
	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return primitive.ObjectID{}, err
	}

	// Si fue una inserción (upsert), obtener el ID del documento insertado
	var result bson.M
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return primitive.ObjectID{}, err
	}

	// Extraer el ID del documento
	objectID, ok := result["_id"].(primitive.ObjectID)
	if !ok {
		return primitive.ObjectID{}, fmt.Errorf("could not convert _id to ObjectID")
	}

	return objectID, nil
}

func (r *ChatsRepository) GetChatsByUserIDWithStatus(ctx context.Context, userID primitive.ObjectID, status string, page, limit int) ([]*Chatsdomain.ChatWithUsers, error) {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("chats")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Determinar el estado que queremos filtrar (según sea user1 o user2)
	matchStatus := bson.M{
		"$or": bson.A{
			bson.M{"user1_id": userID, "status_user1": status}, // Filtrar por el estado si es user1
			bson.M{"user2_id": userID, "status_user2": status}, // Filtrar por el estado si es user2
		},
	}

	// Calcular el salto para la paginación
	skip := (page - 1) * limit

	pipeline := bson.A{
		// Filtrar los chats por el userID y el estado
		bson.D{{Key: "$match", Value: matchStatus}},

		// Hacer un lookup para traer los detalles de los usuarios
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "Users"},
			{Key: "let", Value: bson.D{
				{Key: "user1_id", Value: "$user1_id"},
				{Key: "user2_id", Value: "$user2_id"},
			}},
			{Key: "pipeline", Value: bson.A{
				bson.D{{Key: "$match", Value: bson.M{
					"$expr": bson.M{
						"$or": bson.A{
							bson.M{"$eq": bson.A{"$_id", "$$user1_id"}},
							bson.M{"$eq": bson.A{"$_id", "$$user2_id"}},
						},
					},
				}}},
				bson.D{{Key: "$project", Value: bson.M{
					"_id":      1,
					"NameUser": 1,
					"Avatar":   1,
					"Partner":  1,
				}}},
			}},
			{Key: "as", Value: "users"},
		}}},

		// Ordenar por la fecha del último mensaje
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "LastMessage", Value: -1},
		}}},

		// Aplicar paginación: saltar y limitar el número de documentos
		bson.D{{Key: "$skip", Value: skip}},
		bson.D{{Key: "$limit", Value: limit}},
	}

	// Ejecutar la consulta agregada
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decodificar los resultados
	var chats []*Chatsdomain.ChatWithUsers
	for cursor.Next(ctx) {
		var chat Chatsdomain.ChatWithUsers
		if err := cursor.Decode(&chat); err != nil {
			return nil, err
		}
		chats = append(chats, &chat)
	}

	return chats, nil
}

func (r *ChatsRepository) MarkMessageAsSeen(messageID string) error {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": messageID}
	update := bson.M{"$set": bson.M{"seen": true}}

	_, err := collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(false))
	return err
}

func (r *ChatsRepository) GetMessageByID(messageID string) (*Chatsdomain.Message, error) {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var message Chatsdomain.Message
	err := collection.FindOne(ctx, bson.M{"_id": messageID}).Decode(&message)
	if err != nil {
		return nil, err
	}

	return &message, nil
}
func (r *ChatsRepository) UpdateUserStatus(ctx context.Context, chatID, userID primitive.ObjectID, newStatus string) error {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("chats")

	// Obtener el chat para verificar qué usuario es
	var chat Chatsdomain.Chat
	err := collection.FindOne(ctx, bson.M{"_id": chatID}).Decode(&chat)
	if err != nil {
		return fmt.Errorf("error al buscar el chat: %v", err)
	}

	// Crear la consulta de actualización según el usuario
	var update bson.M
	if chat.User1ID == userID {
		// Si es el user1, actualiza el status_user1
		update = bson.M{"$set": bson.M{"status_user1": newStatus}}
	} else if chat.User2ID == userID {
		// Si es el user2, actualiza el status_user2
		update = bson.M{"$set": bson.M{"status_user2": newStatus}}
	} else {
		return fmt.Errorf("el usuario %v no es parte de este chat", userID)
	}

	// Actualizar el chat en la base de datos
	_, err = collection.UpdateOne(ctx, bson.M{"_id": chatID}, update)
	if err != nil {
		return fmt.Errorf("error al actualizar el estado del usuario: %v", err)
	}

	return nil
}
