package advertisementsroutes

import (
	advertisementsapplication "PINKKER-BACKEND/internal/advertisements/advertisements-application"
	advertisementsinfrastructure "PINKKER-BACKEND/internal/advertisements/advertisements-infrastructure"
	advertisementsinterface "PINKKER-BACKEND/internal/advertisements/advertisements-interface"
	"PINKKER-BACKEND/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func AdvertisementsRoutes(App *fiber.App, redisClient *redis.Client, newMongoDB *mongo.Client) {

	Repository := advertisementsinfrastructure.NewadvertisementsRepository(redisClient, newMongoDB)
	Service := advertisementsapplication.NewAdvertisementsService(Repository)
	Handler := advertisementsinterface.NewwithdrawService(Service)

	App.Get("/advertisements/GetAdvertisements", middleware.UseExtractor(), Handler.GetAdvertisements)
	App.Post("/advertisements/CreateAdvertisement", middleware.UseExtractor(), Handler.CreateAdvertisement)
	App.Post("/advertisements/UpdateAdvertisement", middleware.UseExtractor(), Handler.UpdateAdvertisement)
	App.Post("/advertisements/DeleteAdvertisement", middleware.UseExtractor(), Handler.DeleteAdvertisement)

}
