package cryptoroutes

import (
	cryptoinfrastructure "PINKKER-BACKEND/internal/Crypto/Crypto-infrastructure"
	cryptoapplication "PINKKER-BACKEND/internal/crypto/crypto-application"
	cryptopinterface "PINKKER-BACKEND/internal/crypto/crypto-interface"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func CryptoRoutes(App *fiber.App, redisClient *redis.Client, newMongoDB *mongo.Client) {
	cryptoRepository := cryptoinfrastructure.NewCryptoRepository(redisClient, newMongoDB)
	cryptoService := cryptoapplication.NewryptoService(cryptoRepository)
	cryptoHandler := cryptopinterface.NewCryptoHandler(cryptoService)

	App.Post("/crypto/Subscription", cryptoHandler.Subscription)
}
