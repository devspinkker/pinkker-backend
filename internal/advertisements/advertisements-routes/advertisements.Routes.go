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

	App.Post("/advertisements/GetAdvertisements", middleware.UseExtractor(), Handler.GetAdvertisements)

	App.Post("/advertisements/BuyadCreate", middleware.UseExtractor(), middleware.TOTPAuthMiddleware(Repository), Handler.BuyadCreate)
	App.Post("/advertisements/BuyadMuroCommunity", middleware.UseExtractor(), middleware.TOTPAuthMiddleware(Repository), Handler.BuyadMuroCommunity)

	App.Post("/advertisements/CreateAdsClips", middleware.UseExtractor(), middleware.TOTPAuthMiddleware(Repository), Handler.CreateAdsClips)

	App.Post("/advertisements/GetAllPendingAds", middleware.UseExtractor(), Handler.GetAllPendingAds)
	App.Post("/advertisements/GetAdsUserPendingCode", middleware.UseExtractor(), Handler.GetAdsUserPendingCode)

	App.Post("/advertisements/AcceptPendingAds", middleware.UseExtractor(), Handler.AcceptPendingAds)
	App.Post("/advertisements/RemovePendingAds", middleware.UseExtractor(), Handler.RemovePendingAds)

	App.Post("/advertisements/CreateAdvertisement", middleware.UseExtractor(), Handler.CreateAdvertisement)
	App.Post("/advertisements/UpdateAdvertisement", middleware.UseExtractor(), Handler.UpdateAdvertisement)
	App.Post("/advertisements/DeleteAdvertisement", middleware.UseExtractor(), Handler.DeleteAdvertisement)
	App.Post("/advertisements/IdOfTheUsersWhoClicked", middleware.UseExtractor(), Handler.IdOfTheUsersWhoClicked)

	App.Get("/advertisements/GetAdsUser", middleware.UseExtractor(), Handler.GetAdsUser)
	App.Post("/advertisements/GetAdsUserCode", middleware.UseExtractor(), Handler.GetAdsUserCode)
}
