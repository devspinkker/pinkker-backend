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

func (r *ChatsRepository) SaveMessage(message *Chatsdomain.Message) (*Chatsdomain.Message, error) {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, message)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": message.ID}

	var savedMessage Chatsdomain.Message
	err = collection.FindOne(ctx, filter).Decode(&savedMessage)
	if err != nil {
		return nil, err
	}

	return &savedMessage, nil
}

func (r *ChatsRepository) AddMessageToChat(user1ID, user2ID, messageID string) (string, error) {
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
	}

	opts := options.Update().SetUpsert(true)

	// Realizar la actualización
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return "", err
	}

	// Si fue una inserción (upsert), obtener el ID del documento insertado
	var result bson.M
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return "", err
	}

	// Extraer el ID del documento
	objectID, ok := result["_id"].(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("could not convert _id to ObjectID")
	}

	return objectID.Hex(), nil
}

func (r *ChatsRepository) GetMessages(senderID, receiverID string) ([]*Chatsdomain.Message, error) {
	collection := r.mongoClient.Database("PINKKER-BACKEND").Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"sender_id": senderID, "receiver_id": receiverID}
	cursor, err := collection.Find(ctx, filter)
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
