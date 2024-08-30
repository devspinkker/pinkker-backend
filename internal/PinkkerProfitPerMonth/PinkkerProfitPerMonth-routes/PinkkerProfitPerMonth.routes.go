package PinkkerProfitPerMonthroutes

import (
	PinkkerProfitPerMonthapplication "PINKKER-BACKEND/internal/PinkkerProfitPerMonth/PinkkerProfitPerMonth-application"
	PinkkerProfitPerMonthinfrastructure "PINKKER-BACKEND/internal/PinkkerProfitPerMonth/PinkkerProfitPerMonth-infrastructure"
	PinkkerProfitPerMonthinterfaces "PINKKER-BACKEND/internal/PinkkerProfitPerMonth/PinkkerProfitPerMonth-interface"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func PinkkerProfitPerMonthRoutes(App *fiber.App, redisClient *redis.Client, newMongoDB *mongo.Client) {

	Repository := PinkkerProfitPerMonthinfrastructure.NewPinkkerProfitPerMonthRepository(redisClient, newMongoDB)
	Service := PinkkerProfitPerMonthapplication.NewPinkkerProfitPerMonthService(Repository)
	_ = PinkkerProfitPerMonthinterfaces.NewPinkkerProfitPerMonthService(Service)

	// App.Get("/vods/vod_streamer", Handler.GetVodtreamer)
	// App.Get("/vods/get_vod", Handler.GetVodWithId)

}
