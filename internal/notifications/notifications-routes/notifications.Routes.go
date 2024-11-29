package notificationsroutes

import (
	notificationsapplication "PINKKER-BACKEND/internal/notifications/notifications-application"
	Notificationstinfrastructure "PINKKER-BACKEND/internal/notifications/notifications-infrastructure"
	Notificationstinterfaces "PINKKER-BACKEND/internal/notifications/notifications-interface"
	"PINKKER-BACKEND/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func NotificationsRoutes(App *fiber.App, redisClient *redis.Client, mongoClient *mongo.Client) {
	repo := Notificationstinfrastructure.NewNotificationsRepository(redisClient, mongoClient)
	service := notificationsapplication.NewNotificationsService(repo)
	handler := Notificationstinterfaces.NewNotificationsHandler(service)

	App.Get("/notifications/recent", middleware.UseExtractor(), handler.GetRecentNotifications)
	App.Get("/notifications/GetOldNotifications", middleware.UseExtractor(), handler.GetOldNotifications)
}
