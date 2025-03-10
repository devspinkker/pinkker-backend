package Chatsroutes

import (
	Chatsapplication "PINKKER-BACKEND/internal/Chat/Chats-application"
	Chatsinfrastructure "PINKKER-BACKEND/internal/Chat/Chats-infrastructure"
	Chatsinterface "PINKKER-BACKEND/internal/Chat/Chats-interface"
	"PINKKER-BACKEND/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func ChatsRoutes(app *fiber.App, redisClient *redis.Client, mongoClient *mongo.Client) {
	repository := Chatsinfrastructure.NewChatsRepository(redisClient, mongoClient)
	service := Chatsapplication.NewChatsService(repository)
	handler := Chatsinterface.NewChatsHandler(service, redisClient)
	app.Post("/chats/CreateChatOrGetChats", middleware.UseExtractor(), handler.CreateChatOrGetChats)
	app.Post("/chats/send", middleware.UseExtractor(), handler.SendMessage)
	app.Get("/chats/GetChatsByUserIDWithStatus", middleware.UseExtractor(), handler.GetChatsByUserIDWithStatus)
	app.Get("/chats/messages", middleware.UseExtractor(), handler.GetMessages)
	app.Post("/chats/seen/:id", middleware.UseExtractor(), handler.MarkMessageAsSeen)
	app.Get("/ws/chat/:roomID/:token", websocket.New(handler.WebSocketHandler))

	app.Post("/chats/UpdateUserStatus", middleware.UseExtractor(), handler.UpdateUserStatus)

	app.Post("/chats/DeleteAllMessages", middleware.UseExtractor(), handler.DeleteAllMessages)
	app.Post("/chats/UpdateChatBlockStatus", middleware.UseExtractor(), handler.UpdateChatBlockStatus)

}
