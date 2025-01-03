package clipapplication

import (
	clipdomain "PINKKER-BACKEND/internal/clip/clip-domain"
	clipinfrastructure "PINKKER-BACKEND/internal/clip/clip-infrastructure"
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

func (u *ClipService) DeleteClipByIDAndUserID(clipID, userID primitive.ObjectID) error {
	return u.ClipRepository.DeleteClipByIDAndUserID(clipID, userID)
}
func (u *ClipService) UpdateClipTitle(clipID, userID primitive.ObjectID, title string) error {
	return u.ClipRepository.UpdateClipTitle(clipID, userID, title)
}
func (u *ClipService) GetClipsByNameUserIDOrdenación(UserID primitive.ObjectID, filterType string, dateRange string, page int, limit int) ([]clipdomain.GetClip, error) {
	return u.ClipRepository.GetClipsByNameUserIDOrdenación(UserID, filterType, dateRange, page, limit)
}

func (u *ClipService) TimeOutClipCreate(idClip primitive.ObjectID) error {
	err := u.ClipRepository.TimeOutClipCreate(idClip)
	return err
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

		NameUserCreator:       nameUser,
		IDCreator:             idCreator,
		NameUser:              user.NameUser,
		UserID:                user.ID,
		Avatar:                user.Avatar,
		ClipTitle:             ClipTitle,
		URL:                   outputFilePath,
		Likes:                 []primitive.ObjectID{},
		StreamThumbnail:       CategorieStream.StreamThumbnail,
		Category:              CategorieStream.StreamCategory,
		Duration:              10,
		Views:                 0,
		Cover:                 "",
		Comments:              []primitive.ObjectID{},
		Timestamps:            timeStamps,
		IdOfTheUsersWhoViewed: []primitive.ObjectID{},
		Type:                  "clip",
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
func (u *ClipService) GetClipId(clipId primitive.ObjectID) (clipdomain.GetClip, error) {
	clip, err := u.ClipRepository.FindClipById(clipId)
	return clip, err
}
func (u *ClipService) GetClipIdLogueado(clipId, idValueObj primitive.ObjectID) (*clipdomain.GetClip, error) {
	clip, err := u.ClipRepository.GetClipIdLogueado(clipId, idValueObj)
	return clip, err
}

func (u *ClipService) GetClipsNameUser(page int, NameUser string) ([]clipdomain.GetClip, error) {

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
func (u *ClipService) GetClipsWeightedByDate(page int) ([]clipdomain.Clip, error) {

	Clips, err := u.ClipRepository.GetClipsWeightedByDate(page)
	return Clips, err
}

func (u *ClipService) GetClipsMostViewedLast48Hours(page int) ([]clipdomain.Clip, error) {

	Clips, err := u.ClipRepository.GetClipsMostViewedLast48Hours(page)
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
func (u *ClipService) MoreViewOfTheClip(idClip primitive.ObjectID, idt primitive.ObjectID) error {
	err := u.ClipRepository.MoreViewOfTheClip(idClip, idt)
	return err
}
func (u *ClipService) ClipsRecommended(idT primitive.ObjectID, excludeIDs []primitive.ObjectID) ([]clipdomain.GetClip, error) {
	limit := 10
	clips, err := u.ClipRepository.ClipsRecommended(idT, limit, excludeIDs)
	return clips, err
}
func (u *ClipService) GetClipsByTitle(title string) ([]clipdomain.GetClip, error) {
	limit := 10
	clips, err := u.ClipRepository.GetClipsByTitle(title, limit)
	return clips, err
}
func (u *ClipService) CommentClip(clipID, userID primitive.ObjectID, username, comment string) (clipdomain.ClipCommentGet, error) {
	return u.ClipRepository.CommentClip(clipID, userID, username, comment)
}

func (u *ClipService) DeleteComment(commentID primitive.ObjectID, idValueToken primitive.ObjectID) error {
	return u.ClipRepository.DeleteComment(commentID, idValueToken)
}
func (u *ClipService) LikeCommentClip(idClip primitive.ObjectID, idValueToken primitive.ObjectID) error {
	return u.ClipRepository.LikeComment(idClip, idValueToken)
}

func (u *ClipService) UnlikeComment(idClip primitive.ObjectID, idValueToken primitive.ObjectID) error {
	err := u.ClipRepository.UnlikeComment(idClip, idValueToken)
	return err
}
func (u *ClipService) GetClipComments(idClip primitive.ObjectID, page int) ([]clipdomain.ClipCommentGet, error) {
	clips, err := u.ClipRepository.GetClipComments(idClip, page)
	return clips, err
}
func (u *ClipService) GetClipCommentsLoguedo(idClip primitive.ObjectID, page int, idt primitive.ObjectID) ([]clipdomain.ClipCommentGet, error) {
	clips, err := u.ClipRepository.GetClipCommentsLoguedo(idClip, page, idt)
	return clips, err
}
func (u *ClipService) GetClipTheAd(clips []clipdomain.GetClip) (clipdomain.GetClip, error) {

	clipId, err := u.ClipRepository.GetAdClips()
	if err != nil {
		return clipdomain.GetClip{}, err
	}
	clipAds, err := u.ClipRepository.FindClipById(clipId)

	u.ClipRepository.AddAds(clipAds.AdId, clips)

	return clipAds, err
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
