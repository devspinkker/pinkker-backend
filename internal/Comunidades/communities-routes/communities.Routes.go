package communitiesroutes

import (
	communitiesapplication "PINKKER-BACKEND/internal/Comunidades/communities-application"
	communitiestinfrastructure "PINKKER-BACKEND/internal/Comunidades/communities-infrastructure"
	communitiestinterfaces "PINKKER-BACKEND/internal/Comunidades/communities-interface"
	"PINKKER-BACKEND/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

func CommunitiesRoutes(App *fiber.App, redisClient *redis.Client, newMongoDB *mongo.Client) {

	Repository := communitiestinfrastructure.NewcommunitiesRepository(redisClient, newMongoDB)
	Service := communitiesapplication.NewCommunitiesService(Repository)
	Handler := communitiestinterfaces.NewCommunitiesHandler(Service)

	App.Post("/communities/CreateCommunity", middleware.UseExtractor(), middleware.TOTPAuthMiddleware(Repository), Handler.CreateCommunity)
	App.Post("/communities/EditCommunity", middleware.UseExtractor(), middleware.TOTPAuthMiddleware(Repository), Handler.EditCommunity)

	App.Post("/communities/AddMember", middleware.UseExtractor(), Handler.AddMember)

	App.Post("/communities/RemoveMember", middleware.UseExtractor(), Handler.RemoveMember)
	App.Post("/communities/FindUserCommunities", Handler.FindUserCommunities)
	App.Post("/communities/CommunityOwnerUser", Handler.CommunityOwnerUser)

	App.Post("/communities/BanMember", middleware.UseExtractor(), Handler.BanMember)
	App.Post("/communities/AddModerator", middleware.UseExtractor(), Handler.AddModerator)
	App.Post("/communities/DeletePost", middleware.UseExtractor(), Handler.DeletePost)

	App.Post("/communities/DeleteCommunity", middleware.UseExtractor(), middleware.TOTPAuthMiddleware(Repository), Handler.DeleteCommunity)
	App.Get("/communities/FindCommunityByName", Handler.FindCommunityByName)
	App.Get("/communities/GetCommunity", Handler.GetCommunity)
	App.Get("/communities/GetCommunityWithUserMembership", middleware.UseExtractor(), Handler.GetCommunityWithUserMembership)
	App.Get("/communities/GetTop10CommunitiesByMembers", middleware.UseExtractor(), Handler.GetTop10CommunitiesByMembers)
	App.Get("/communities/GetTop10CommunitiesByMembersNoMember", middleware.UseExtractor(), Handler.GetTop10CommunitiesByMembersNoMember)
	App.Get("/communities/GetCommunityRecommended", middleware.UseExtractor(), Handler.GetCommunityRecommended)

}
