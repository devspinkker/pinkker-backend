package cryptoapplication

import cryptoinfrastructure "PINKKER-BACKEND/internal/Crypto/Crypto-infrastructure"

type CryptoService struct {
	CryptoRepository *cryptoinfrastructure.CryptoRepository
}

func NewryptoService(CryptoRepository *cryptoinfrastructure.CryptoRepository) *CryptoService {
	return &CryptoService{
		CryptoRepository: CryptoRepository,
	}
}
