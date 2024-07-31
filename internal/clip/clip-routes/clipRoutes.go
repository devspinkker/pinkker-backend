package cliproutes

import (
	clipapplication "PINKKER-BACKEND/internal/clip/clip-application"
	clipinfrastructure "PINKKER-BACKEND/internal/clip/clip-infrastructure"
	clipinterface "PINKKER-BACKEND/internal/clip/clip-interface"
	"PINKKER-BACKEND/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func ClipRoutes(App *fiber.App, redisClient *redis.Client, newMongoDB *mongo.Client) {
	clipRepository := clipinfrastructure.NewClipRepository(redisClient, newMongoDB)
	clipService := clipapplication.NewClipService(clipRepository)
	clipHandler := clipinterface.NewClipHandler(clipService)

	App.Post("/clips/create-clips", middleware.UseExtractor(), clipHandler.CreateClips)
	App.Get("/clips/GetClipId", clipHandler.GetClipId)
	App.Get("/clips/GetClipIdLogueado", middleware.UseExtractor(), clipHandler.GetClipId)

	App.Get("/clips/GetClipsByTitle", clipHandler.GetClipsByTitle)
	App.Get("/clips/GetClipsNameUser", clipHandler.GetClipsNameUser)
	App.Get("/clips/GetClipsCategory", clipHandler.GetClipsCategory)
	App.Get("/clips/GetClipsMostViewed", clipHandler.GetClipsMostViewed)
	App.Get("/clips/GetClipsMostViewedLast48Hours", clipHandler.GetClipsMostViewedLast48Hours)

	App.Post("/clips/ClipLike", middleware.UseExtractor(), clipHandler.CliptLike)
	App.Post("/clips/DisLike", middleware.UseExtractor(), clipHandler.ClipDislike)
	App.Post("/clips/MoreViewOfTheClip", clipHandler.MoreViewOfTheClip)
	App.Post("/clips/ClipsRecommended", middleware.UseExtractor(), clipHandler.ClipsRecommended)

	App.Post("/clips/CommentClip", middleware.UseExtractor(), clipHandler.CommentClip)
	App.Post("/clips/LikeCommentClip", middleware.UseExtractor(), clipHandler.LikeCommentClip)
	App.Post("/clips/UnlikeComment", middleware.UseExtractor(), clipHandler.UnlikeComment)

	App.Get("/clips/GetClipComments", clipHandler.GetClipComments)

	App.Get("/clips/TimeOutClipCreate", middleware.UseExtractor(), clipHandler.TimeOutClipCreate)

}
