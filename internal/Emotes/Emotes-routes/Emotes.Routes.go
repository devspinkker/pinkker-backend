package Emotesroutes

import (
	Emotesapplication "PINKKER-BACKEND/internal/Emotes/Emotes-application"
	Emotesinfrastructure "PINKKER-BACKEND/internal/Emotes/Emotes-infrastructure"
	Emotesinterface "PINKKER-BACKEND/internal/Emotes/Emotes-interface"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func EmotesRoutes(App *fiber.App, redisClient *redis.Client, newMongoDB *mongo.Client) {

	Repository := Emotesinfrastructure.NewEmotesRepository(redisClient, newMongoDB)
	Service := Emotesapplication.NewEmotesService(Repository)
	Handler := Emotesinterface.NewwithdrawService(Service)

	App.Get("Emotes/GetGlobalEmotes", Handler.GetGlobalEmotes)
	App.Get("Emotes/GetPinkkerEmotes", Handler.GetPinkkerEmotes)
}
