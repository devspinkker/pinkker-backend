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

	App.Post("/tweet/tweetCreate", middleware.UseExtractor(), tweetHandler.CreateTweet)
	App.Post("/tweet/CommentPost", middleware.UseExtractor(), tweetHandler.CommentPost)

	App.Post("/tweet/tweetLike", middleware.UseExtractor(), tweetHandler.TweetLike)
	App.Post("/tweet/tweetDislike", middleware.UseExtractor(), tweetHandler.TweetDislike)

	App.Get("/tweet/tweetGetFollow", middleware.UseExtractor(), tweetHandler.TweetGetFollow)
	App.Get("/tweet/tweetGetCommentPost", middleware.UseExtractor(), tweetHandler.GetCommentPost)
}
