// clip-infrastructure/clip_repository.go
package clipinfrastructure

import (
	clipdomain "PINKKER-BACKEND/internal/clip/clip-domain"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"context"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

func (c *ClipRepository) SaveClip(clip *clipdomain.Clip) error {
	clipCollection := c.mongoClient.Database("PINKKER-BACKEND").Collection("Clips")
	_, err := clipCollection.InsertOne(context.Background(), clip)
	if err != nil {
		return err
	}

	return nil
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
