package StreamSummaryroutes

import (
	StreamSummaryapplication "PINKKER-BACKEND/internal/StreamSummary.repository/StreamSummary-application"
	StreamSummaryinfrastructure "PINKKER-BACKEND/internal/StreamSummary.repository/StreamSummary-infrastructure"
	StreamSummaryinterfaces "PINKKER-BACKEND/internal/StreamSummary.repository/StreamSummary-interface"
	"PINKKER-BACKEND/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func StreamSummaryRoutes(App *fiber.App, redisClient *redis.Client, newMongoDB *mongo.Client) {

	StreamSummaryRepository := StreamSummaryinfrastructure.NewStreamSummaryRepository(redisClient, newMongoDB)
	StreamSummaryService := StreamSummaryapplication.NewStreaSummaryService(StreamSummaryRepository)
	StreamSummary := StreamSummaryinterfaces.NewStreamSummaryService(StreamSummaryService)

	App.Post("StreamSummary/Update", middleware.UseExtractor(), StreamSummary.UpdateStreamSummary)
	App.Post("StreamSummary/AdsAdd", middleware.UseExtractor(), StreamSummary.AddAds)

	App.Post("StreamSummary/GetLastSixStreamSummaries", middleware.UseExtractor(), StreamSummary.GetLastSixStreamSummaries)
	App.Post("StreamSummary/AverageViewers", StreamSummary.AverageViewers)

}
