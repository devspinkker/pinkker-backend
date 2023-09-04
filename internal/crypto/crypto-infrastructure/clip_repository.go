package cryptoinfrastructure

import (
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type CryptoRepository struct {
	redisClient *redis.Client
	mongoClient *mongo.Client
}

func NewCryptoRepository(redisClient *redis.Client, mongoClient *mongo.Client) *CryptoRepository {
	return &CryptoRepository{
		redisClient: redisClient,
		mongoClient: mongoClient,
	}
}
