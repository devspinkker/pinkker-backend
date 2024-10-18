package advertisementsapplication

import (
	"PINKKER-BACKEND/internal/advertisements/advertisements"
	advertisementsinfrastructure "PINKKER-BACKEND/internal/advertisements/advertisements-infrastructure"
	clipdomain "PINKKER-BACKEND/internal/clip/clip-domain"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdvertisementsService struct {
	AdvertisementsRepository *advertisementsinfrastructure.AdvertisementsRepository
}

func NewAdvertisementsService(AdvertisementsRepository *advertisementsinfrastructure.AdvertisementsRepository) *AdvertisementsService {
	return &AdvertisementsService{
		AdvertisementsRepository: AdvertisementsRepository,
	}
}
func (s *AdvertisementsService) GetAdvertisements(StreamerID primitive.ObjectID, data advertisements.AdvertisementGet, page int64) ([]advertisements.Advertisements, error) {
	err := s.AdvertisementsRepository.AutCode(StreamerID, data.Code)
	if err != nil {
		return []advertisements.Advertisements{}, err
	}
	advertisementsGet, err := s.AdvertisementsRepository.AdvertisementsGet(page)
	return advertisementsGet, err
}
func (s *AdvertisementsService) AcceptPendingAds(StreamerID primitive.ObjectID, data advertisements.AcceptPendingAds) error {
	err := s.AdvertisementsRepository.AutCode(StreamerID, data.Code)
	if err != nil {
		return err
	}
	return s.AdvertisementsRepository.AcceptPendingAds(data.NameUser)
}

func (s *AdvertisementsService) GetAllPendingAds(StreamerID primitive.ObjectID, data advertisements.AdvertisementGet, page int64) ([]advertisements.Advertisements, error) {
	err := s.AdvertisementsRepository.AutCode(StreamerID, data.Code)
	if err != nil {
		return []advertisements.Advertisements{}, err
	}
	return s.AdvertisementsRepository.GetAllPendingAds(page)
}

func (s *AdvertisementsService) GetAllPendingNameUserAds(StreamerID primitive.ObjectID, data advertisements.AcceptPendingAds, page int64) ([]advertisements.Advertisements, error) {
	err := s.AdvertisementsRepository.AutCode(StreamerID, data.Code)
	if err != nil {
		return []advertisements.Advertisements{}, err
	}
	return s.AdvertisementsRepository.GetAllPendingNameUserAds(page, data.NameUser)
}

func (s *AdvertisementsService) RemovePendingAds(StreamerID primitive.ObjectID, data advertisements.AcceptPendingAds) error {
	err := s.AdvertisementsRepository.AutCode(StreamerID, data.Code)
	if err != nil {
		return err
	}
	return s.AdvertisementsRepository.RemovePendingAds(data.NameUser)
}
func (s *AdvertisementsService) GetAdsUser(NameUser string) ([]advertisements.Advertisements, error) {

	advertisementsGet, err := s.AdvertisementsRepository.GetAdsUser(NameUser)
	return advertisementsGet, err
}

func (s *AdvertisementsService) GetAdsUserCode(data advertisements.GetAdsUserCode, StreamerID primitive.ObjectID) ([]advertisements.Advertisements, error) {
	err := s.AdvertisementsRepository.AutCode(StreamerID, data.Code)
	if err != nil {
		return []advertisements.Advertisements{}, err
	}
	advertisementsGet, err := s.AdvertisementsRepository.GetAdsUser(data.NameUser)
	return advertisementsGet, err
}

func (s *AdvertisementsService) CreateAdvertisement(StreamerID primitive.ObjectID, data advertisements.UpdateAdvertisement) (advertisements.Advertisements, error) {
	err := s.AdvertisementsRepository.AutCode(StreamerID, data.Code)
	if err != nil {
		return advertisements.Advertisements{}, err
	}
	advertisementsGet, err := s.AdvertisementsRepository.CreateAdvertisement(data)
	return advertisementsGet, err
}
func (s *AdvertisementsService) BuyadCreate(StreamerID primitive.ObjectID, data advertisements.UpdateAdvertisement) (advertisements.Advertisements, error) {
	advertisementsGet, err := s.AdvertisementsRepository.BuyadCreate(data, StreamerID)
	return advertisementsGet, err
}

func (s *AdvertisementsService) ClipAdsCreate(data advertisements.ClipAdsCreate) (advertisements.Advertisements, error) {
	advertisementsGet, err := s.AdvertisementsRepository.CreateAdsAdvertisement(data)
	return advertisementsGet, err
}

func (s *AdvertisementsService) UpdateAdvertisement(StreamerID primitive.ObjectID, data advertisements.UpdateAdvertisement) (advertisements.Advertisements, error) {
	err := s.AdvertisementsRepository.AutCode(StreamerID, data.Code)
	if err != nil {
		return advertisements.Advertisements{}, err
	}
	advertisementsGet, err := s.AdvertisementsRepository.UpdateAdvertisement(data)
	return advertisementsGet, err
}
func (s *AdvertisementsService) IdOfTheUsersWhoClicked(IdU primitive.ObjectID, idAdvertisements primitive.ObjectID) error {

	return s.AdvertisementsRepository.IdOfTheUsersWhoClicked(IdU, idAdvertisements)

}
func (s *AdvertisementsService) DeleteAdvertisement(StreamerID primitive.ObjectID, data advertisements.DeleteAdvertisement) error {
	err := s.AdvertisementsRepository.AutCode(StreamerID, data.Code)
	if err != nil {
		return err
	}
	err = s.AdvertisementsRepository.DeleteAdvertisement(data.ID)
	return err
}

func (u *AdvertisementsService) CreateClipForAds(idCreator primitive.ObjectID, nameUser string, ClipTitle string, outputFilePath string) (*clipdomain.Clip, error) {
	avatar, err := u.AdvertisementsRepository.FindUser(idCreator)
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
		NameUser:              nameUser,
		UserID:                idCreator,
		Avatar:                avatar,
		ClipTitle:             ClipTitle,
		URL:                   outputFilePath,
		Likes:                 []primitive.ObjectID{},
		StreamThumbnail:       "",
		Category:              "Ad",
		Duration:              10,
		Views:                 0,
		Cover:                 "",
		Comments:              []primitive.ObjectID{},
		Timestamps:            timeStamps,
		Type:                  "Ad",
		IdOfTheUsersWhoViewed: []primitive.ObjectID{},
	}
	clipid, err := u.AdvertisementsRepository.SaveClip(clip)
	clip.ID = clipid
	return clip, err
}
func (u *AdvertisementsService) UpdateClip(clipUpdate *clipdomain.Clip, ulrClip string, ad primitive.ObjectID) {
	u.AdvertisementsRepository.UpdateClip(clipUpdate.ID, ulrClip, ad)
}

func (s *AdvertisementsService) BuyadClipCreate(StreamerID primitive.ObjectID, data advertisements.ClipAdsCreate, clipid primitive.ObjectID) (advertisements.Advertisements, error) {
	advertisementsGet, err := s.AdvertisementsRepository.BuyadClipCreate(data, StreamerID, clipid)
	return advertisementsGet, err
}
