package donationapplication

import (
	donationdomain "PINKKER-BACKEND/internal/donation/donation"
	donationtinfrastructure "PINKKER-BACKEND/internal/donation/donation-infrastructure"

	"github.com/gofiber/websocket/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DonationService struct {
	DonationRepository *donationtinfrastructure.DonationRepository
}

func NewDonationService(DonationRepository *donationtinfrastructure.DonationRepository) *DonationService {
	return &DonationService{
		DonationRepository: DonationRepository,
	}
}
func (D *DonationService) PublishNotification(roomID primitive.ObjectID, noty map[string]interface{}) error {
	stream, err := D.DonationRepository.GetStreamByStreamerID(roomID)
	if err != nil {
		return err
	}
	return D.DonationRepository.PublishNotification(stream.ID.Hex(), noty)

}

// FromUser tiene pixeles?
func (D *DonationService) UserHasNumberPikels(FromUser primitive.ObjectID, Pixeles float64) error {
	err := D.DonationRepository.UserHasNumberPikels(FromUser, Pixeles)
	return err
}
func (D *DonationService) LatestStreamSummaryByUpdateDonations(FromUser primitive.ObjectID, pixeles float64) error {
	err := D.DonationRepository.LatestStreamSummaryByUpdateDonations(FromUser, pixeles)
	return err
}
func (D *DonationService) StateTheUserInChat(Donado primitive.ObjectID, Donante primitive.ObjectID) (bool, string, error) {
	return D.DonationRepository.StateTheUserInChat(Donado, Donante)
}

// donar pixeles de fromUser a ToUser
func (D *DonationService) DonatePixels(FromUser primitive.ObjectID, ToUser primitive.ObjectID, Pixeles float64, text string) error {
	err := D.DonationRepository.DonatePixels(FromUser, ToUser, Pixeles, text)
	return err
}
func (D *DonationService) GetWebSocketActivityFeed(user string) ([]*websocket.Conn, error) {
	client, err := D.DonationRepository.GetWebSocketClientsInRoom(user)
	return client, err
}

// user de token, para ver laas donaciones que le han hecho, solo la que Notified sea false
func (D *DonationService) MyPixelesdonors(id primitive.ObjectID) ([]donationdomain.ResDonation, error) {
	Donations, err := D.DonationRepository.MyPixelesdonors(id)

	return Donations, err
}

// todos los donantes de Pixeles de user token
func (D *DonationService) AllMyPixelesDonors(id primitive.ObjectID) ([]donationdomain.ResDonation, error) {
	Donations, err := D.DonationRepository.AllMyPixelesDonors(id)

	return Donations, err
}
func (D *DonationService) GetPixelesDonationsChat(id primitive.ObjectID) ([]donationdomain.ResDonation, error) {
	Donations, err := D.DonationRepository.GetPixelesDonationsChat(id)

	return Donations, err
}

// actualzaa el Notified
func (D *DonationService) UpdateDonationsNotifiedStatus(donation []donationdomain.ResDonation) error {
	err := D.DonationRepository.UpdateDonationsNotifiedStatus(donation)
	return err
}
