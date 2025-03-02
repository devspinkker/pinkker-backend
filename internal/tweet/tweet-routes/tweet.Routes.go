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
	tweetHandler := tweetinterfaces.NewTweetService(tweetService, redisClient)

	App.Post("/post/postCreate", middleware.UseExtractor(), tweetHandler.CreatePost)
	App.Post("/post/CommentPost", middleware.UseExtractor(), tweetHandler.CommentPost)
	// App.Post("/post/Repost", middleware.UseExtractor(), tweetHandler.RePost)
	App.Post("/post/Citapost", middleware.UseExtractor(), tweetHandler.CitaPost)

	App.Post("/post/posttLike", middleware.UseExtractor(), tweetHandler.PostLike)
	App.Post("/post/postDislike", middleware.UseExtractor(), tweetHandler.PostDislike)

	App.Get("/post/PostGets", tweetHandler.PostGets)
	App.Get("/post/PostGetId", tweetHandler.GetPostId)
	App.Get("/post/PostGetIdLogueado", middleware.UseExtractor(), tweetHandler.GetPostIdLogueado)

	App.Get("/post/postGetFollow", middleware.UseExtractor(), tweetHandler.TweetGetFollow)
	App.Get("/post/getPostUser", tweetHandler.GetPostuser)

	App.Get("/post/GetPostsWithImages", tweetHandler.GetPostsWithImages)

	App.Get("/post/getPostUserLogueado", middleware.UseExtractor(), tweetHandler.GetPostuserLogueado)

	App.Post("/post/GetTweetsRecommended", middleware.UseExtractor(), tweetHandler.GetTweetsRecommended)
	App.Post("/post/GetRandomPostcommunities", middleware.UseExtractor(), tweetHandler.GetRandomPostcommunities)
	App.Post("/post/GetPostCommunitiesFromUser", middleware.UseExtractor(), tweetHandler.GetPostCommunitiesFromUser)

	App.Get("/post/GetTrends", tweetHandler.GetTrends)
	App.Get("/post/GetTweetsByHashtag", tweetHandler.GetTweetsByHashtag)
	App.Get("/post/GetTrendsByPrefix", tweetHandler.GetTrendsByPrefix)
	App.Get("/post/GetCommentPost", middleware.UseExtractor(), tweetHandler.GetCommentPost)

	App.Post("/communities/GetCommunityPosts", middleware.UseExtractor(), tweetHandler.GetCommunityPosts)
	App.Post("/communities/GetCommunityPostsGallery", middleware.UseExtractor(), tweetHandler.GetCommunityPostsGallery)

}
