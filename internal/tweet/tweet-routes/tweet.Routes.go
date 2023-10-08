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

	App.Post("/post/postCreate", middleware.UseExtractor(), tweetHandler.CreatePost)
	App.Post("/post/CommentPost", middleware.UseExtractor(), tweetHandler.CommentPost)
	App.Post("/post/Repost", middleware.UseExtractor(), tweetHandler.RePost)
	App.Post("/post/Citapost", middleware.UseExtractor(), tweetHandler.CitaPost)

	App.Post("/post/posttLike", middleware.UseExtractor(), tweetHandler.PostLike)
	App.Post("/post/postDislike", middleware.UseExtractor(), tweetHandler.PostDislike)

	App.Get("/post/postGetFollow", middleware.UseExtractor(), tweetHandler.TweetGetFollow)
	App.Get("/post/PostGets", tweetHandler.PostGets)

	App.Post("/post/GetCommentPost", middleware.UseExtractor(), tweetHandler.GetCommentPost)
}
