package tweetroutes

import (
	tweetapplication "PINKKER-BACKEND/internal/tweet/tweet-application"
	tweetinfrastructure "PINKKER-BACKEND/internal/tweet/tweet-infrastructure"
	tweetinterfaces "PINKKER-BACKEND/internal/tweet/tweet-interface"
	"PINKKER-BACKEND/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func TweetdRoutes(App *fiber.App, redisClient *redis.Client, newMongoDB *mongo.Client) {

	tweetRepository := tweetinfrastructure.NewTweetRepository(redisClient, newMongoDB)
	tweetService := tweetapplication.NewTweetService(tweetRepository)
	tweetHandler := tweetinterfaces.NewTweetService(tweetService)

	App.Post("/tweetCreate", middleware.UseExtractor(), tweetHandler.CreateTweet)

	App.Post("/tweetLike", middleware.UseExtractor(), tweetHandler.TweetLike)
	App.Post("/tweetDislike", middleware.UseExtractor(), tweetHandler.TweetDislike)

	App.Get("/tweetGetFollow", middleware.UseExtractor(), tweetHandler.TweetGetFollow)

}
