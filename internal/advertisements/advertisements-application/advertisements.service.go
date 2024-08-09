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
func (s *AdvertisementsService) GetAdvertisements(StreamerID primitive.ObjectID, data advertisements.AdvertisementGet) ([]advertisements.Advertisements, error) {
	err := s.AdvertisementsRepository.AutCode(StreamerID, data.Code)
	if err != nil {
		return []advertisements.Advertisements{}, err
	}
	advertisementsGet, err := s.AdvertisementsRepository.AdvertisementsGet()
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
