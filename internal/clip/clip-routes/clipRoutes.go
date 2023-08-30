// vod-routes.go (o el archivo correspondiente)
package main

import (
	clipapplication "PINKKER-BACKEND/internal/clip/clip-application"
	clipinfrastructure "PINKKER-BACKEND/internal/clip/clip-infrastructure"
	clipinterface "PINKKER-BACKEND/internal/clip/clip-interface"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func ClipRoutes(App *fiber.App, redisClient *redis.Client, newMongoDB *mongo.Client) {
	clipRepository := clipinfrastructure.NewClipRepository(redisClient, newMongoDB)
	clipService := clipapplication.NewClipService(clipRepository)
	clipHandler := clipinterface.NewClipHandler(clipService)

	App.Post("/createClip", clipHandler.CreateClip)
}
