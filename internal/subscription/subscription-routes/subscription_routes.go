package subscriptionroutes

import (
	subscriptionapplication "PINKKER-BACKEND/internal/subscription/subscription-application"
	subscriptioninfrastructure "PINKKER-BACKEND/internal/subscription/subscription-infrastructure"
	subscriptioninterfaces "PINKKER-BACKEND/internal/subscription/subscription-interfaces"
	"PINKKER-BACKEND/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func SubsRoutes(App *fiber.App, redisClient *redis.Client, newMongoDB *mongo.Client) {

	userRepository := subscriptioninfrastructure.NewSubscriptionRepository(redisClient, newMongoDB)
	userService := subscriptionapplication.NewChatService(userRepository)
	UserHandler := subscriptioninterfaces.NewSubscriptionHandler(userService)

	App.Post("/Subs/suscribirse", middleware.UseExtractor(), UserHandler.Suscribirse)
	App.Get("/Subs/GetSubsChat", UserHandler.GetSubsChat)

}
