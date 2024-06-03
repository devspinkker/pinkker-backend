package advertisementsapplication

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
func (s *WithdrawalsService) WithdrawalRequest(StreamerID primitive.ObjectID, nameUser string, data withdrawalsdomain.WithdrawalRequestReq) error {
	err := s.withdrawalsRepository.WithdrawalRequest(StreamerID, nameUser, data)
	return err
}
func (s *WithdrawalsService) GetPendingUnnotifiedWithdrawals(id primitive.ObjectID, data withdrawalsdomain.WithdrawalRequestGet) ([]withdrawalsdomain.WithdrawalRequests, error) {

	err := s.withdrawalsRepository.AutCode(id, data.Code)
	if err != nil {
		return nil, err
	}
	Withdrawal, err := s.withdrawalsRepository.GetPendingUnnotifiedWithdrawals(data)
	return Withdrawal, err

}
func (s *WithdrawalsService) AcceptWithdrawal(id primitive.ObjectID, data withdrawalsdomain.AcceptWithdrawal) error {

	err := s.withdrawalsRepository.AutCode(id, data.Code)
	if err != nil {
		return err
	}
	err = s.withdrawalsRepository.AcceptWithdrawal(id, data)

	return err

}
func (s *WithdrawalsService) RejectWithdrawal(id primitive.ObjectID, data withdrawalsdomain.RejectWithdrawal) error {

	err := s.withdrawalsRepository.AutCode(id, data.Code)
	if err != nil {
		return err
	}
	err = s.withdrawalsRepository.RejectWithdrawal(id, data)

	return err

}
func (s *WithdrawalsService) GetWithdrawalToken(id primitive.ObjectID) ([]withdrawalsdomain.WithdrawalRequests, error) {

	data, err := s.withdrawalsRepository.GetWithdrawalToken(id)

	return data, err

}
