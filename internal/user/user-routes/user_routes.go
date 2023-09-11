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

	App.Post("/user/signup", UserHandler.Signup)
	App.Post("/user/login", UserHandler.Login)

	App.Get("/user/getUserById", middleware.UseExtractor(), UserHandler.GetUserById)
	App.Post("/user/get_user_by_key", UserHandler.GetUserBykey)

	//Follow
	App.Post("/user/follow", middleware.UseExtractor(), UserHandler.Follow)
	App.Post("/user/Unfollow", middleware.UseExtractor(), UserHandler.Unfollow)

	App.Post("/user/buyPixeles", middleware.UseExtractor(), UserHandler.BuyPixeles)

}
