package communitiesapplication

import (
	communitiesdomain "PINKKER-BACKEND/internal/Comunidades/communities"
	communitiestinfrastructure "PINKKER-BACKEND/internal/Comunidades/communities-infrastructure"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommunitiesService struct {
	communitiesRepository *communitiestinfrastructure.CommunitiesRepository
}

func NewCommunitiesService(communitiesRepository *communitiestinfrastructure.CommunitiesRepository) *CommunitiesService {
	return &CommunitiesService{
		communitiesRepository: communitiesRepository,
	}
}

// Crear una nueva comunidad
func (s *CommunitiesService) CreateCommunity(ctx context.Context, name string, creatorID primitive.ObjectID, description string, isPrivate bool, categories []string) (interface{}, error) {
	return s.communitiesRepository.CreateCommunity(ctx, name, creatorID, description, isPrivate, categories)
}

// Agregar un miembro a la comunidad
func (s *CommunitiesService) AddMember(ctx context.Context, communityID, userID primitive.ObjectID) error {
	return s.communitiesRepository.AddMember(ctx, communityID, userID)
}
func (s *CommunitiesService) RemoveMember(ctx context.Context, communityID, userID primitive.ObjectID) error {
	return s.communitiesRepository.RemoveMember(ctx, communityID, userID)
}

// Expulsar un miembro de la comunidad
func (s *CommunitiesService) BanMember(ctx context.Context, communityID, userID, mod primitive.ObjectID) error {
	return s.communitiesRepository.BanMember(ctx, communityID, userID, mod)
}
func (s *CommunitiesService) GetCommunityPosts(ctx context.Context, CommunityID primitive.ObjectID, ExcludeFilterIDs []primitive.ObjectID, idT primitive.ObjectID) ([]communitiesdomain.PostGetCommunityReq, error) {
	return s.communitiesRepository.GetCommunityPosts(ctx, CommunityID, ExcludeFilterIDs, idT, 10)
}
func (s *CommunitiesService) AddModerator(ctx context.Context, communityID, newModID, modID primitive.ObjectID) error {
	return s.communitiesRepository.AddModerator(ctx, communityID, newModID, modID)
}
func (s *CommunitiesService) DeletePost(ctx context.Context, communityID primitive.ObjectID, postID primitive.ObjectID, userID primitive.ObjectID) error {
	return s.communitiesRepository.DeletePost(ctx, communityID, postID, userID)
}
func (s *CommunitiesService) DeleteCommunity(ctx context.Context, communityID primitive.ObjectID, creatorID primitive.ObjectID) error {
	return s.communitiesRepository.DeleteCommunity(ctx, communityID, creatorID)
}
func (s *CommunitiesService) FindCommunityByName(ctx context.Context, community string) ([]communitiesdomain.CommunityDetails, error) {
	return s.communitiesRepository.FindCommunityByName(ctx, community)
}
func (s *CommunitiesService) GetTop10CommunitiesByMembers(ctx context.Context) ([]communitiesdomain.CommunityDetails, error) {
	return s.communitiesRepository.GetTop10CommunitiesByMembers(ctx)
}
func (s *CommunitiesService) GetCommunity(ctx context.Context, community primitive.ObjectID) (*communitiesdomain.CommunityDetails, error) {
	return s.communitiesRepository.GetCommunity(ctx, community)
}
func (s *CommunitiesService) GetCommunityWithUserMembership(ctx context.Context, community, user primitive.ObjectID) (*communitiesdomain.CommunityDetails, error) {
	return s.communitiesRepository.GetCommunityWithUserMembership(ctx, community, user)
}
