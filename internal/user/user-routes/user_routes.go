package userroutes

import (
	application "PINKKER-BACKEND/internal/user/user-application"
	infrastructure "PINKKER-BACKEND/internal/user/user-infrastructure"
	interfaces "PINKKER-BACKEND/internal/user/user-interfaces"
	"PINKKER-BACKEND/pkg/jwt"
	"PINKKER-BACKEND/pkg/middleware"
	"PINKKER-BACKEND/pkg/utils"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func UserRoutes(App *fiber.App, redisClient *redis.Client, newMongoDB *mongo.Client) {

	userRepository := infrastructure.NewUserRepository(redisClient, newMongoDB)
	userService := application.NewChatService(userRepository)
	UserHandler := interfaces.NewUserHandler(userService)

	App.Post("/user/signupNotConfirmed", UserHandler.SignupSaveUserRedis)
	App.Post("/user/SaveUserCodeConfirm", UserHandler.SaveUserCodeConfirm)
	App.Post("/user/login", UserHandler.Login)
	App.Post("/user/LoginTOTPSecret", UserHandler.LoginTOTPSecret)
	App.Post("/user/GetRecommendedUsers", middleware.UseExtractor(), UserHandler.GetRecommendedUsers)
	// recuperacion
	App.Post("/user/Get_Recover_lost_password", UserHandler.Get_Recover_lost_password)
	App.Post("/user/account-recovery", UserHandler.RestorePassword)
	App.Get("/user/PurchasePinkkerPrime", middleware.UseExtractor(), UserHandler.PurchasePinkkerPrime)

	App.Post("/user/ChangeGoogleAuthenticator", middleware.UseExtractor(), UserHandler.ChangeGoogleAuthenticator)
	App.Post("/user/DeleteGoogleAuthenticator", middleware.UseExtractor(), UserHandler.DeleteGoogleAuthenticator)

	// oauth2
	App.Get("/user/google_login", UserHandler.GoogleLogin)
	App.Get("/user/google_callback", UserHandler.Google_callback)
	App.Post("/user/Google_callback_Complete_Profile_And_Username", UserHandler.Google_callback_Complete_Profile_And_Username)

	App.Post("/generate-totp-key", middleware.UseExtractor(), UserHandler.GenerateTOTPKey)
	App.Post("/validate-totp-code", middleware.UseExtractor(), UserHandler.ValidateTOTPCode)

	App.Get("/user/getUserByNameUser", UserHandler.GetUserByNameUser)
	App.Get("/user/GetStreamAndUserData", middleware.UseExtractor(), UserHandler.GetStreamAndUserData)

	App.Get("/user/getUserByNameUserIndex", UserHandler.GetUserByNameUserIndex)
	App.Get("/user/get_user_cmt", UserHandler.GetUserByCmt)
	App.Get("/user/get_user_by_key", UserHandler.GetUserBykey)
	App.Get("/user/GetUserBanInstream", UserHandler.GetUserBanInstream)

	App.Get("/user/getUserById", middleware.UseExtractor(), UserHandler.GetUserByIdTheToken)
	App.Get("/user/GetNotificacionesLastConnection", middleware.UseExtractor(), UserHandler.GetNotificacionesLastConnection)
	//Follow
	App.Post("/user/follow", middleware.UseExtractor(), UserHandler.Follow)
	App.Post("/user/Unfollow", middleware.UseExtractor(), UserHandler.Unfollow)
	App.Post("/user/buyPixeles", middleware.UseExtractor(), UserHandler.BuyPixeles)

	// edit user information
	App.Post("/user/EditProfile", middleware.UseExtractor(), UserHandler.EditProfile)
	App.Post("/user/ChangeNameUser", middleware.UseExtractor(), middleware.TOTPAuthMiddleware(userRepository), UserHandler.ChangeNameUser)
	App.Post("/user/EditAvatar", middleware.UseExtractor(), UserHandler.EditAvatar)
	App.Post("/user/EditBanner", middleware.UseExtractor(), UserHandler.EditBanner)

	App.Post("/user/EditSocialNetworks", middleware.UseExtractor(), UserHandler.EditSocialNetworks)

	App.Post("/user/PanelAdminPinkker/InfoUser", middleware.UseExtractor(), UserHandler.PanelAdminPinkkerInfoUser)
	App.Post("/user/PanelAdminPinkker/CreateAdmin", middleware.UseExtractor(), UserHandler.CreateAdmin)
	App.Post("/user/PanelAdminPinkker/ChangeNameUserCodeAdmin", middleware.UseExtractor(), UserHandler.ChangeNameUserCodeAdmin)

	App.Post("/user/PanelAdminPinkker/banStreamer", middleware.UseExtractor(), UserHandler.PanelAdminPinkkerbanStreamer)
	App.Post("/user/PanelAdminPinkker/RemoveBanStreamer", middleware.UseExtractor(), UserHandler.PanelAdminRemoveBanStreamer)
	App.Post("/user/PanelAdminPinkker/PanelAdminPinkkerPartnerUser", middleware.UseExtractor(), UserHandler.PanelAdminPinkkerPartnerUser)

	App.Get("/ws/notification/ActivityFeed/:user", websocket.New(func(c *websocket.Conn) {
		user := c.Params("user") + "ActivityFeed"
		chatService := utils.NewChatService()
		client := &utils.Client{Connection: c}
		chatService.AddClientToRoom(user, client)

		defer func() {
			chatService.RemoveClientFromRoom(user, client)
			_ = c.Close()
		}()

		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	App.Get("/ws/pinker_notifications/:token", websocket.New(func(c *websocket.Conn) {
		UserHandler.Pinker_notifications(c)
		defer func() {
			_ = c.Close()
		}()
		token := c.Params("token", "null")
		var id primitive.ObjectID

		if token != "null" {
			_, idToken, _, err := jwt.ExtractDataFromToken(token)
			if err != nil {
				return
			}
			IdUserTokenP, errinObjectID := primitive.ObjectIDFromHex(idToken)
			if errinObjectID != nil {
				c.WriteMessage(websocket.TextMessage, []byte(errinObjectID.Error()))
				c.Close()
			}
			id = IdUserTokenP
		}
		errReceiveMessageFromRoom := UserHandler.Pinker_notifications(c)
		if errReceiveMessageFromRoom != nil {
			err := UserHandler.UpdateLastConnection(id)
			if err != nil {
				fmt.Println(err)
			}
			c.WriteMessage(websocket.TextMessage, []byte(errReceiveMessageFromRoom.Error()))
			c.Close()
			return
		}

	}))

}
