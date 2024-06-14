package Emotesinfrastructure

import (
	EmotesDomain "PINKKER-BACKEND/internal/Emotes/Emotes"
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

type EmotesRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewEmotesRepository(redisClient *redis.Client, mongoClient *mongo.Client) *EmotesRepository {
	return &EmotesRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}

func (r *EmotesRepository) CreateEmote(emote EmotesDomain.Emote) (*EmotesDomain.Emote, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Emotes")

	emote.ID = primitive.NewObjectID()
	emote.CreatedAt = time.Now()

	_, err := collection.InsertOne(context.Background(), emote)
	if err != nil {
		return nil, err
	}
	return &emote, nil
}

func (r *EmotesRepository) AddEmoteAut(emote EmotesDomain.EmoteUpdate) (*EmotesDomain.Emote, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Emotes")

	filter := bson.M{"_id": emote.ID}
	update := bson.M{
		"$push": bson.M{"emotes": emote.Emotes},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedEmote EmotesDomain.Emote

	err := collection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&updatedEmote)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &updatedEmote, nil
}
func (r *EmotesRepository) DeleteEmoteAut(emote EmotesDomain.EmoteUpdate) (*EmotesDomain.Emote, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Emotes")

	filter := bson.M{"_id": emote.ID}
	update := bson.M{
		"$pull": bson.M{"emotes": emote.Emotes},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedEmote EmotesDomain.Emote

	err := collection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&updatedEmote)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &updatedEmote, nil
}

// DeleteEmote elimina un emote de la base de datos
func (r *EmotesRepository) DeleteEmote(emoteID primitive.ObjectID) error {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Emotes")

	filter := bson.M{"_id": emoteID}

	_, err := collection.DeleteOne(context.Background(), filter)
	return err
}

// GetEmote obtiene un emote por su ID
func (r *EmotesRepository) GetEmote(emoteID primitive.ObjectID) (*EmotesDomain.Emote, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Emotes")

	var emote EmotesDomain.Emote
	err := collection.FindOne(context.Background(), bson.M{"_id": emoteID}).Decode(&emote)
	if err != nil {
		return nil, err
	}
	return &emote, nil
}
func (r *EmotesRepository) DeleteEmoteForType(userId primitive.ObjectID, emoteType string, emoteName string) error {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Emotes")

	partnerFilter := bson.M{"_id": userId, "Partner.Active": true}
	collectionUsers := db.Collection("Users")
	var user userdomain.User
	err := collectionUsers.FindOne(context.Background(), partnerFilter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("el usuario no es un Partner activo")
		}
		return err
	}

	filter := bson.M{"userId": userId, "type": emoteType}
	update := bson.M{
		"$pull": bson.M{"emotes": bson.M{"name": emoteName}},
	}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("no se encontró ningún documento con el userId y emoteType especificados")
		}
		return err
	}

	return nil
}

func (r *EmotesRepository) UpdateOrCreateEmoteByUserAndType(userId primitive.ObjectID, emoteType string, emote EmotesDomain.EmotePair) (EmotesDomain.Emote, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Emotes")

	partnerFilter := bson.M{"_id": userId, "Partner.Active": true}
	collectionUsers := db.Collection("Users")
	var user userdomain.User
	err := collectionUsers.FindOne(context.Background(), partnerFilter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return EmotesDomain.Emote{}, fmt.Errorf("el usuario no es un Partner activo")
		}
		return EmotesDomain.Emote{}, err
	}

	// Construir el filtro y la actualización
	filter := bson.M{"userId": userId, "type": emoteType}
	update := bson.M{
		"$push": bson.M{"emotes": emote},
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var updatedEmote EmotesDomain.Emote

	// Realizar la operación de findOneAndUpdate
	err = collection.FindOneAndUpdate(context.Background(), filter, update, opts).Decode(&updatedEmote)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Si no existe, crear un nuevo documento
			newEmote := EmotesDomain.Emote{
				UserID:    userId,
				Type:      emoteType,
				Emotes:    []EmotesDomain.EmotePair{emote},
				CreatedAt: time.Now(),
			}
			newEmote.ID = primitive.NewObjectID()

			_, err := collection.InsertOne(context.Background(), newEmote)
			if err != nil {
				return EmotesDomain.Emote{}, err
			}
			return newEmote, nil
		}
		return EmotesDomain.Emote{}, err
	}

	return updatedEmote, nil
}

func (r *EmotesRepository) GetEmoteIdUserandType(userId primitive.ObjectID, emoteType string) (EmotesDomain.Emote, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Emotes")

	var emote EmotesDomain.Emote
	filter := bson.M{"userId": userId, "type": emoteType}
	err := collection.FindOne(context.Background(), filter).Decode(&emote)
	if err != nil {
		return EmotesDomain.Emote{}, err
	}
	return emote, nil
}

// GetAllEmotes obtiene todos los emotes de la colección
func (r *EmotesRepository) GetAllEmotes() ([]EmotesDomain.Emote, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Emotes")

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var emotes []EmotesDomain.Emote
	for cursor.Next(context.Background()) {
		var emote EmotesDomain.Emote
		if err := cursor.Decode(&emote); err != nil {
			return nil, err
		}
		emotes = append(emotes, emote)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return emotes, nil
}

// ChangeEmoteTypeToGlobal cambia el tipo de un emote a "global"
func (r *EmotesRepository) ChangeEmoteTypeToGlobal(emoteID primitive.ObjectID) (*EmotesDomain.Emote, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Emotes")

	filter := bson.M{"_id": emoteID}
	update := bson.M{
		"$set": bson.M{"type": "global"},
	}

	var emote EmotesDomain.Emote
	err := collection.FindOneAndUpdate(context.Background(), filter, update).Decode(&emote)
	if err != nil {
		return nil, err
	}
	return &emote, nil
}

// ChangeEmoteTypeToPinkker cambia el tipo de un emote a "pinkker"
func (r *EmotesRepository) ChangeEmoteTypeToPinkker(emoteID primitive.ObjectID) (*EmotesDomain.Emote, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Emotes")

	filter := bson.M{"_id": emoteID}
	update := bson.M{
		"$set": bson.M{"type": "pinkker"},
	}

	var emote EmotesDomain.Emote
	err := collection.FindOneAndUpdate(context.Background(), filter, update).Decode(&emote)
	if err != nil {
		return nil, err
	}
	return &emote, nil
}
func (r *EmotesRepository) GetEmotesByType(emoteType string) ([]EmotesDomain.Emote, error) {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collection := db.Collection("Emotes")

	filter := bson.M{"type": emoteType}
	options := options.Find()
	cursor, err := collection.Find(context.Background(), filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var emotes []EmotesDomain.Emote
	for cursor.Next(context.Background()) {
		var emote EmotesDomain.Emote
		if err := cursor.Decode(&emote); err != nil {
			return nil, err
		}
		emotes = append(emotes, emote)
	}

	return emotes, nil
}
func (r *EmotesRepository) AutCode(id primitive.ObjectID, code string) error {
	db := r.mongoClient.Database("PINKKER-BACKEND")
	collectionUsers := db.Collection("Users")
	var User userdomain.User

	err := collectionUsers.FindOne(context.Background(), bson.M{"_id": id}).Decode(&User)
	if err != nil {
		return err
	}

	if User.PanelAdminPinkker.Level != 1 || !User.PanelAdminPinkker.Asset || User.PanelAdminPinkker.Code != code {
		return fmt.Errorf("usuario no autorizado")
	}
	return nil
}
