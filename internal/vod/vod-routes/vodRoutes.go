package vodroutes

import (
	vodsapplication "PINKKER-BACKEND/internal/vod/vod-application"
	vodinfrastructure "PINKKER-BACKEND/internal/vod/vod-infrastructure"
	vodinterfaces "PINKKER-BACKEND/internal/vod/vod-interface"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func VodRoutes(App *fiber.App, redisClient *redis.Client, newMongoDB *mongo.Client) {

	vodRepository := vodinfrastructure.NewVodRepository(redisClient, newMongoDB)
	vodService := vodsapplication.NewVodService(vodRepository)
	vodHandler := vodinterfaces.NewVodService(vodService)

	App.Get("/vods/vod_streamer", vodHandler.GetVodtreamer)
	App.Get("/vods/get_vod", vodHandler.GetVodWithId)
	// getVodTrending
	// getClipsTrending
	App.Post("/vods/createVod", vodHandler.CreateVod)

}
