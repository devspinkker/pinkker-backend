// clip-infrastructure/clip_repository.go
package clipinfrastructure

import (
	clipdomain "PINKKER-BACKEND/internal/clip/clip-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ClipRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewClipRepository(redisClient *redis.Client, mongoClient *mongo.Client) *ClipRepository {
	return &ClipRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}

func (c *ClipRepository) SaveClip(clip *clipdomain.Clip) (*clipdomain.Clip, error) {
	database := c.mongoClient.Database("PINKKER-BACKEND")
	clipCollection := database.Collection("Clips")
	userCollection := database.Collection("Users")

	result, err := clipCollection.InsertOne(context.Background(), clip)
	if err != nil {
		return nil, err
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("no se pudo obtener el ID insertado")
	}
	clip.ID = insertedID

	filterUser := bson.M{"_id": clip.UserID}
	update := bson.M{"$push": bson.M{"Clips": insertedID}}

	opts := options.Update().SetUpsert(false)

	resultuserCollection, err := userCollection.UpdateOne(context.Background(), filterUser, update, opts)
	if err != nil {
		return clip, err
	}

	if resultuserCollection.ModifiedCount == 0 {
		return clip, errors.New("No se encontraron documentos para actualizar.")
	}

	return clip, err
}
func (c *ClipRepository) UpdateClip(clipID primitive.ObjectID, newURL string) {
	clipCollection := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")

	filter := bson.M{"_id": clipID}

	update := bson.M{"$set": bson.M{"url": newURL}}

	opts := options.Update().SetUpsert(false)

	clipCollection.UpdateOne(context.Background(), filter, update, opts)

}
func (c *ClipRepository) FindrClipId(IdClip primitive.ObjectID) (*clipdomain.Clip, error) {
	GoMongoDBCollUsers := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")
	FindClipInDb := bson.D{
		{Key: "_id", Value: IdClip},
	}
	var findClipInDbExist *clipdomain.Clip
	err := GoMongoDBCollUsers.FindOne(context.Background(), FindClipInDb).Decode(&findClipInDbExist)
	return findClipInDbExist, err
}
func (c *ClipRepository) FindUser(NameUser string) (*userdomain.User, error) {
	GoMongoDBCollUsers := c.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	var FindUserInDb primitive.D
	FindUserInDb = bson.D{
		{Key: "NameUser", Value: NameUser},
	}
	var findUserInDbExist *userdomain.User
	errCollUsers := GoMongoDBCollUsers.FindOne(context.Background(), FindUserInDb).Decode(&findUserInDbExist)
	return findUserInDbExist, errCollUsers
}
func (c *ClipRepository) FindUserId(FindUserId string) (*userdomain.User, error) {
	GoMongoDBCollUsers := c.mongoClient.Database("PINKKER-BACKEND").Collection("Users")
	var FindUserInDb primitive.D
	FindUserInDb = bson.D{
		{Key: "NameUser", Value: FindUserId},
	}
	var findUserInDbExist *userdomain.User
	errCollUsers := GoMongoDBCollUsers.FindOne(context.Background(), FindUserInDb).Decode(&findUserInDbExist)
	return findUserInDbExist, errCollUsers
}
