package clipapplication

import (
	clipdomain "PINKKER-BACKEND/internal/clip/clip-domain"
	clipinfrastructure "PINKKER-BACKEND/internal/clip/clip-infrastructure"
	"PINKKER-BACKEND/pkg/helpers"
	"fmt"
	"path/filepath"
	"strings"
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

func (u *ClipService) CreateClip(idCreator primitive.ObjectID, totalKey string, nameUser string, ClipTitle string, TsUrls []string) (*clipdomain.Clip, error) {
	user, err := u.ClipRepository.FindUser(totalKey)
	if err != nil {
		return &clipdomain.Clip{}, err
	}

	CategorieStream, err := u.ClipRepository.FindCategorieStream(user.ID)
	if err != nil {
		return &clipdomain.Clip{}, err
	}

	UrlsClipFormate, err := u.UrlsGenerateClips(CategorieStream.ID, TsUrls)
	urls, m3u8Content, err := u.GenerateCustomM3U8(UrlsClipFormate)
	if err != nil {
		return &clipdomain.Clip{}, err
	}

	fmt.Println("TS URLs:\n", urls)
	fmt.Println("M3U8 Content:\n", m3u8Content)
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
	imagePrevClip, err := helpers.CopyImageFromURL(CategorieStream.StreamThumbnail)
	if err != nil {
		fmt.Println(err)
		imagePrevClip = CategorieStream.StreamThumbnail
	}

	clip := &clipdomain.Clip{
		NameUserCreator:       nameUser,
		IDCreator:             idCreator,
		NameUser:              user.NameUser,
		UserID:                user.ID,
		Avatar:                user.Avatar,
		ClipTitle:             ClipTitle,
		URL:                   urls,
		Likes:                 []primitive.ObjectID{},
		StreamThumbnail:       imagePrevClip,
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

func (u *ClipService) UrlsGenerateClips(streamID primitive.ObjectID, tsIndexes []string) ([]string, error) {

	Summary, err := u.ClipRepository.GetCurrentStreamSummary(streamID)
	if err != nil {
		return []string{}, err
	}
	return u.ClipRepository.GenerateStreamURLs(Summary, tsIndexes)
}

func (r *ClipService) GenerateCustomM3U8(tsURLs []string) (string, string, error) {
	if len(tsURLs) == 0 {
		return "", "", fmt.Errorf("no TS URLs provided")
	}

	// Construir el contenido del archivo M3U8
	var m3u8Content strings.Builder
	m3u8Content.WriteString("#EXTM3U\n")
	m3u8Content.WriteString("#EXT-X-VERSION:3\n")
	m3u8Content.WriteString("#EXT-X-TARGETDURATION:10\n") // Ajustar duración estimada por segmento
	m3u8Content.WriteString(fmt.Sprintf("#EXT-X-MEDIA-SEQUENCE:%d\n", extractSequenceNumber(tsURLs[0])))

	for _, url := range tsURLs {
		m3u8Content.WriteString("#EXTINF:10.0,\n") // Duración por segmento
		m3u8Content.WriteString(url + "\n")
	}

	m3u8Content.WriteString("#EXT-X-ENDLIST\n")

	// Unir las URLs en un string para referencia
	urlList := strings.Join(tsURLs, ",")

	return urlList, m3u8Content.String(), nil
}

// Función auxiliar para extraer el número de secuencia de una URL
func extractSequenceNumber(tsURL string) int {
	// Asume que el archivo tiene un formato como "index123.ts"
	baseName := filepath.Base(tsURL)
	var seqNum int
	fmt.Sscanf(baseName, "index%d.ts", &seqNum)
	return seqNum
}
