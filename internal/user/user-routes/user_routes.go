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

	// oauth2
	App.Get("/user/google_login", UserHandler.GoogleLogin)
	App.Get("/user/google_callback", UserHandler.Google_callback)
	App.Post("/user/Google_callback_Complete_Profile_And_Username", UserHandler.Google_callback_Complete_Profile_And_Username)

	App.Get("/user/getUserByNameUser", UserHandler.GetUserByNameUser)
	App.Get("/user/getUserByNameUserIndex", UserHandler.GetUserByNameUserIndex)

	App.Get("/user/get_user_by_key", UserHandler.GetUserBykey)
	App.Get("/user/getUserById", middleware.UseExtractor(), UserHandler.GetUserByIdTheToken)

	//Follow
	App.Post("/user/follow", middleware.UseExtractor(), UserHandler.Follow)
	App.Post("/user/Unfollow", middleware.UseExtractor(), UserHandler.Unfollow)
	App.Post("/user/buyPixeles", middleware.UseExtractor(), UserHandler.BuyPixeles)

	// edit user information
	App.Post("/user/EditProfile", middleware.UseExtractor(), UserHandler.EditProfile)
	App.Post("/user/EditAvatar", middleware.UseExtractor(), UserHandler.EditAvatar)

}
