package Emotesroutes

import (
	Emotesapplication "PINKKER-BACKEND/internal/Emotes/Emotes-application"
	Emotesinfrastructure "PINKKER-BACKEND/internal/Emotes/Emotes-infrastructure"
	Emotesinterface "PINKKER-BACKEND/internal/Emotes/Emotes-interface"
	"PINKKER-BACKEND/pkg/middleware"

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

	// autorización admin pinkker
	App.Post("Emotes/AddEmoteAut", middleware.UseExtractor(), Handler.AddEmoteAut)
	App.Post("Emotes/DeleteEmoteAut", middleware.UseExtractor(), Handler.DeleteEmoteAut)

	App.Post("Emotes/CreateOrUpdateEmote", middleware.UseExtractor(), Handler.CreateOrUpdateEmoteWithImage)
	App.Post("Emotes/GetEmoteUserandType", middleware.UseExtractor(), Handler.GetEmoteIdUserandType)
	App.Post("Emotes/DeleteEmoteForType", middleware.UseExtractor(), Handler.DeleteEmoteForType)
}
