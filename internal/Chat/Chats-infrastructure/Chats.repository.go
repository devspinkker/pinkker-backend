package Chatsinfrastructure

import (
	Chatsdomain "PINKKER-BACKEND/internal/Chat/Chats"
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
func (r *ChatsRepository) usersExist(user1ID, user2ID primitive.ObjectID) (bool, error) {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var userCount int64
	var err error

	// Verificar si user1 existe
	userCount, err = collection.CountDocuments(ctx, bson.M{"_id": user1ID})
	if err != nil || userCount == 0 {
		return false, err
	}

	// Verificar si user2 existe
	userCount, err = collection.CountDocuments(ctx, bson.M{"_id": user2ID})
	if err != nil || userCount == 0 {
		return false, err
	}

	return true, nil
}

func (r *ChatsRepository) CreateChatOrGetChats(user1ID, user2ID primitive.ObjectID) (*Chatsdomain.ChatWithUsers, error) {
	exist, err := r.usersExist(user1ID, user2ID)
	if err != nil {
		return nil, fmt.Errorf("error checking users existence: %v", err)
	}
	if !exist {
		return nil, fmt.Errorf("one or both users do not exist")
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

	// El chat no existe, crearlo
	newChat := Chatsdomain.Chat{
		User1ID:     user1ID,
		User2ID:     user2ID,
		CreatedAt:   time.Now(),
		MessageIDs:  []primitive.ObjectID{},
		LastMessage: time.Now(),
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
	exist, err := r.usersExist(user1ID, user2ID)
	if err != nil {
		return primitive.ObjectID{}, fmt.Errorf("error checking users existence: %v", err)
	}
	if !exist {
		return primitive.ObjectID{}, fmt.Errorf("one or both users do not exist")
	}
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

func (r *ChatsRepository) GetChatsByUserID(userID primitive.ObjectID) ([]*Chatsdomain.ChatWithUsers, error) {
	// Convertir el userID de string a ObjectID

	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("chats")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.M{
			"$or": bson.A{
				bson.M{"user1_id": userID}, // Aquí utilizamos objID, que es de tipo primitive.ObjectID
				bson.M{"user2_id": userID}, // Aquí también
			},
		}}},
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
		bson.D{{Key: "$limit", Value: 20}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "LastMessage", Value: -1},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

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
		err := r.UpdateNotificationFlag(chat.NotifyA, primitive.ObjectID{})
		if err != nil {
			return nil, err
		}
	}
	return messages, nil
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
