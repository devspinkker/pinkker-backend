package cliproutes

import (
	clipapplication "PINKKER-BACKEND/internal/clip/clip-application"
	clipinfrastructure "PINKKER-BACKEND/internal/clip/clip-infrastructure"
	clipinterface "PINKKER-BACKEND/internal/clip/clip-interface"
	"PINKKER-BACKEND/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func ClipRoutes(App *fiber.App, redisClient *redis.Client, newMongoDB *mongo.Client) {
	clipRepository := clipinfrastructure.NewClipRepository(redisClient, newMongoDB)
	clipService := clipapplication.NewClipService(clipRepository)
	clipHandler := clipinterface.NewClipHandler(clipService)

	App.Post("/clips/create-clips", middleware.UseExtractor(), clipHandler.CreateClips)
	App.Get("/clips/GetClipId", clipHandler.GetClipId)

	App.Get("/clips/GetClipsNameUser", clipHandler.GetClipsNameUser)
	App.Get("/clips/GetClipsCategory", clipHandler.GetClipsCategory)
	App.Post("/clips/ClipLike", clipHandler.CliptLike)
	App.Post("/clips/DisLike", clipHandler.ClipDislike)

}
