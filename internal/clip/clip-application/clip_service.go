package clipapplication

import (
	clipdomain "PINKKER-BACKEND/internal/clip/clip-domain"
	clipinfrastructure "PINKKER-BACKEND/internal/clip/clip-infrastructure"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ClipService struct {
	ClipRepository *clipinfrastructure.ClipRepository
}

func NewClipService(ClipRepository *clipinfrastructure.ClipRepository) *ClipService {
	return &ClipService{
		ClipRepository: ClipRepository,
	}
}

func (u *ClipService) FindUser(NameUser string) (*userdomain.User, error) {
	findUserInDbExist, errCollUsers := u.ClipRepository.FindUser(NameUser)
	return findUserInDbExist, errCollUsers
}
func (u *ClipService) CreateClip(idCreator primitive.ObjectID, streamer primitive.ObjectID, clipName string, outputFilePath string) (*clipdomain.Clip, error) {
	clip := &clipdomain.Clip{
		StreamerID:    idCreator,
		UserID:        streamer,
		ClipName:      clipName,
		URL:           outputFilePath,
		Likes:         []string{},
		Duration:      10,
		Views:         0,
		Cover:         "",
		TotalLikes:    0,
		TotalRetweets: 0,
		TotalComments: 0,
	}
	clip, err := u.ClipRepository.SaveClip(clip)
	return clip, err
}
func (u *ClipService) UpdateClip(clipUpdate *clipdomain.Clip, ulrClip string) {
	u.ClipRepository.UpdateClip(clipUpdate.ID, ulrClip)
}
