package withdrawapplication

import (
	withdrawalsdomain "PINKKER-BACKEND/internal/withdraw/withdraw"
	withdrawalstinfrastructure "PINKKER-BACKEND/internal/withdraw/withdrawals-infrastructure"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WithdrawalsService struct {
	withdrawalsRepository *withdrawalstinfrastructure.WithdrawalsRepository
}

func NewwithdrawalsService(withdrawalsRepository *withdrawalstinfrastructure.WithdrawalsRepository) *WithdrawalsService {
	return &WithdrawalsService{
		withdrawalsRepository: withdrawalsRepository,
	}
}
func (s *WithdrawalsService) WithdrawalRequest(StreamerID primitive.ObjectID, data withdrawalsdomain.WithdrawalRequestReq) error {
	err := s.withdrawalsRepository.WithdrawalRequest(StreamerID, data)
	return err
}
