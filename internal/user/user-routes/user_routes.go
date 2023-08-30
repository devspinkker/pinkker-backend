package userroutes

import (
	application "PINKKER-BACKEND/internal/user/user-application"
	infrastructure "PINKKER-BACKEND/internal/user/user-infrastructure"
	interfaces "PINKKER-BACKEND/internal/user/user-interfaces"
	"PINKKER-BACKEND/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func UserRoutes(App *fiber.App, redisClient *redis.Client, newMongoDB *mongo.Client) {

	userRepository := infrastructure.NewUserRepository(redisClient, newMongoDB)
	userService := application.NewChatService(userRepository)
	UserHandler := interfaces.NewUserHandler(userService)

	App.Post("/signup", UserHandler.Signup)
	App.Post("/login", UserHandler.Login)

	App.Get("/getUserById", middleware.UseExtractor(), UserHandler.GetUserById)
	//Follow
	App.Post("/follow", middleware.UseExtractor(), UserHandler.Follow)
	App.Post("/Unfollow", middleware.UseExtractor(), UserHandler.Unfollow)

	App.Post("/buyPixeles", middleware.UseExtractor(), UserHandler.BuyPixeles)

}
