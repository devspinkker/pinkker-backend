package advertisementsapplication

import (
	"PINKKER-BACKEND/internal/advertisements/advertisements"
	advertisementsinfrastructure "PINKKER-BACKEND/internal/advertisements/advertisements-infrastructure"

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
func (s *AdvertisementsService) AcceptPendingAds(StreamerID primitive.ObjectID, data advertisements.AdvertisementGet, nameUser string) error {
	err := s.AdvertisementsRepository.AutCode(StreamerID, data.Code)
	if err != nil {
		return err
	}
	return s.AdvertisementsRepository.AcceptPendingAds(nameUser)
}

func (s *AdvertisementsService) GetAllPendingAds(StreamerID primitive.ObjectID, data advertisements.AdvertisementGet, page int64) ([]advertisements.Advertisements, error) {
	err := s.AdvertisementsRepository.AutCode(StreamerID, data.Code)
	if err != nil {
		return []advertisements.Advertisements{}, err
	}
	return s.AdvertisementsRepository.GetAllPendingAds(page)
}
func (s *AdvertisementsService) RemovePendingAds(StreamerID primitive.ObjectID, data advertisements.AdvertisementGet, nameUser string) error {
	err := s.AdvertisementsRepository.AutCode(StreamerID, data.Code)
	if err != nil {
		return err
	}
	return s.AdvertisementsRepository.RemovePendingAds(nameUser)
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
