package cryptoapplication

import (
	cryptoinfrastructure "PINKKER-BACKEND/internal/Crypto/Crypto-infrastructure"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
)

type CryptoService struct {
	CryptoRepository *cryptoinfrastructure.CryptoRepository
}

func NewryptoService(CryptoRepository *cryptoinfrastructure.CryptoRepository) *CryptoService {
	return &CryptoService{
		CryptoRepository: CryptoRepository,
	}
}

func (s *CryptoService) Subscription() {

}

func (s *CryptoService) UpdataSubscriptionState(SourceAddress string, DestinationAddress string) {
	err := s.CryptoRepository.UpdataSubscriptionState(SourceAddress, DestinationAddress)
	return err
}

// USDToBNB realiza la conversión de USD a tokens BNB (ajusta esto según la tasa de cambio)
func (s *CryptoService) USDToBNB(usdAmount float64) *big.Int {
	// Supongamos una tasa de cambio fija de 1 BNB = 500 USD (ajusta según tus necesidades)
	conversionRate := float64(500)
	amountBNB := usdAmount / conversionRate

	// Convierte el monto BNB a Wei (18 decimales)
	amountWei := new(big.Float).Mul(big.NewFloat(amountBNB), big.NewFloat(1e18))
	amount := new(big.Int)
	amountWei.Int(amount)

	return amount
}

func (s *CryptoService) TransferBNB(client *ethclient.Client, sourceAddress, destinationAddress string, amount *big.Int) (string, error) {
	txHash, err := s.CryptoRepository.TransferBNB(client, sourceAddress, destinationAddress, amount)
	return txHash, err
}
