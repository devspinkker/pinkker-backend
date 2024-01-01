package clipapplication

import (
	clipdomain "PINKKER-BACKEND/internal/clip/clip-domain"
	clipinfrastructure "PINKKER-BACKEND/internal/clip/clip-infrastructure"
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	"time"

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
func (u *ClipService) CreateClip(idCreator primitive.ObjectID, totalKey string, nameUser string, ClipTitle string, outputFilePath string) (*clipdomain.Clip, error) {
	user, err := u.ClipRepository.FindUser(totalKey)
	if err != nil {
		return &clipdomain.Clip{}, err
	}
	timeStamps := struct {
		CreatedAt time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
		UpdatedAt time.Time `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	}{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	clip := &clipdomain.Clip{

		NameUserCreator: nameUser,
		IDCreator:       idCreator,
		NameUser:        user.NameUser,
		UserID:          user.ID,
		Avatar:          user.Avatar,
		ClipTitle:       ClipTitle,
		URL:             outputFilePath,
		Likes:           []string{},
		Duration:        10,
		Views:           0,
		Cover:           "",
		TotalLikes:      0,
		TotalRetweets:   0,
		TotalComments:   0,
		Timestamps:      timeStamps,
	}
	clip, err = u.ClipRepository.SaveClip(clip)
	return clip, err
}
func (u *ClipService) UpdateClip(clipUpdate *clipdomain.Clip, ulrClip string) {
	u.ClipRepository.UpdateClip(clipUpdate.ID, ulrClip)
}
func (u *ClipService) GetClipId(clipId primitive.ObjectID) (*clipdomain.Clip, error) {
	clip, err := u.ClipRepository.FindrClipId(clipId)
	return clip, err
}
