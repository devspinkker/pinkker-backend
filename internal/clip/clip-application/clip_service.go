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
	CategorieStream, err := u.ClipRepository.FindCategorieStream(user.ID)
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
		Likes:           []primitive.ObjectID{},
		StreamThumbnail: CategorieStream.StreamThumbnail,
		Category:        CategorieStream.StreamCategory,
		Duration:        10,
		Views:           0,
		Cover:           "",
		Comments:        []primitive.ObjectID{},
		Timestamps:      timeStamps,
	}
	clip, err = u.ClipRepository.SaveClip(clip)
	return clip, err
}
func (u *ClipService) UpdateClip(clipUpdate *clipdomain.Clip, ulrClip string) {
	u.ClipRepository.UpdateClip(clipUpdate.ID, ulrClip)
}
func (u *ClipService) UpdateClipPreviouImage(clipUpdate *clipdomain.Clip, ulrClip string) {
	u.ClipRepository.UpdateClip(clipUpdate.ID, ulrClip)
}
func (u *ClipService) GetClipId(clipId primitive.ObjectID) (*clipdomain.Clip, error) {
	clip, err := u.ClipRepository.FindrClipId(clipId)
	return clip, err
}

func (u *ClipService) GetClipsNameUser(page int, NameUser string) ([]clipdomain.Clip, error) {

	Clips, err := u.ClipRepository.GetClipsNameUser(page, NameUser)
	return Clips, err
}
func (u *ClipService) GetClipsCategory(page int, Category string, lastClipID primitive.ObjectID) ([]clipdomain.Clip, error) {

	Clips, err := u.ClipRepository.GetClipsCategory(page, Category, lastClipID)
	return Clips, err
}
func (u *ClipService) GetClipsMostViewed(page int) ([]clipdomain.Clip, error) {

	Clips, err := u.ClipRepository.GetClipsMostViewed(page)
	return Clips, err
}
func (u *ClipService) LikeClip(idClip primitive.ObjectID, idValueToken primitive.ObjectID) error {
	err := u.ClipRepository.LikeClip(idClip, idValueToken)
	return err
}

func (u *ClipService) ClipDislike(idClip primitive.ObjectID, idValueToken primitive.ObjectID) error {
	err := u.ClipRepository.ClipDislike(idClip, idValueToken)
	return err
}
func (u *ClipService) MoreViewOfTheClip(idClip primitive.ObjectID) error {
	err := u.ClipRepository.MoreViewOfTheClip(idClip)
	return err
}

// func (u *ClipService) ExtractFrameFromVideo(videoPath, outputPath, ffmpegPath string) error {
// 	transcoder.FFmpegBin = ffmpegPath

// 	trans := new(transcoder.Transcoder)
// 	err := trans.Initialize(videoPath, outputPath, ffmpegPath)
// 	if err != nil {
// 		return err
// 	}

// 	err = <-trans.Run(false)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
