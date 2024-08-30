package PinkkerProfitPerMonthapplication

import (
	PinkkerProfitPerMonthdomain "PINKKER-BACKEND/internal/PinkkerProfitPerMonth/PinkkerProfitPerMonth-domain"
	PinkkerProfitPerMonthinfrastructure "PINKKER-BACKEND/internal/PinkkerProfitPerMonth/PinkkerProfitPerMonth-infrastructure"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PinkkerProfitPerMonthService struct {
	PinkkerProfitPerMonthRepository *PinkkerProfitPerMonthinfrastructure.PinkkerProfitPerMonthRepository
}

func NewPinkkerProfitPerMonthService(PinkkerProfitPerMonthRepository *PinkkerProfitPerMonthinfrastructure.PinkkerProfitPerMonthRepository) *PinkkerProfitPerMonthService {
	return &PinkkerProfitPerMonthService{
		PinkkerProfitPerMonthRepository: PinkkerProfitPerMonthRepository,
	}
}

func (s *PinkkerProfitPerMonthService) GetProfitByMonth(streamerID primitive.ObjectID, Code string, Time time.Time) (PinkkerProfitPerMonthdomain.PinkkerProfitPerMonth, error) {

	err := s.PinkkerProfitPerMonthRepository.AutCode(streamerID, Code)
	if err != nil {
		return PinkkerProfitPerMonthdomain.PinkkerProfitPerMonth{}, err
	}
	GetEarningsByWeek, err := s.PinkkerProfitPerMonthRepository.GetProfitByMonth(Time)

	return GetEarningsByWeek, err
}
func (s *PinkkerProfitPerMonthService) GetProfitByMonthRange(streamerID primitive.ObjectID, Code string, Time1, Time2 time.Time) ([]PinkkerProfitPerMonthdomain.PinkkerProfitPerMonth, error) {
	err := s.PinkkerProfitPerMonthRepository.AutCode(streamerID, Code)
	if err != nil {
		return []PinkkerProfitPerMonthdomain.PinkkerProfitPerMonth{}, err
	}
	GetEarningsByMonth, err := s.PinkkerProfitPerMonthRepository.GetProfitByMonthRange(Time1, Time2)
	return GetEarningsByMonth, err
}
