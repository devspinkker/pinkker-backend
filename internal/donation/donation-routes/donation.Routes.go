package donationroutes

import (
	application "PINKKER-BACKEND/internal/donation/donation-application"
	infrastructure "PINKKER-BACKEND/internal/donation/donation-infrastructure"
	interfaces "PINKKER-BACKEND/internal/donation/donation-interface"
	"PINKKER-BACKEND/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func DonatioRoutes(App *fiber.App, redisClient *redis.Client, newMongoDB *mongo.Client) {

	donationRepository := infrastructure.NewDonationRepository(redisClient, newMongoDB)
	donationService := application.NewDonationService(donationRepository)
	donationHandler := interfaces.NewDonationService(donationService)

	App.Post("/pixel/DonatePixel", middleware.UseExtractor(), donationHandler.Donate)
	App.Get("/pixel/Mydonors", middleware.UseExtractor(), donationHandler.Mydonors)
	App.Get("/pixel/AllMyPixelesDonors", middleware.UseExtractor(), donationHandler.AllMyPixelesDonors)
	App.Get("/pixel/GetPixelesDonationsChat", donationHandler.GetPixelesDonationsChat)

}
