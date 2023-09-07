package cryptopinterface

import (
	cryptoapplication "PINKKER-BACKEND/internal/crypto/crypto-application"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gofiber/fiber/v2"
)

type CryptoHandler struct {
	CryptoService *cryptoapplication.CryptoService
}

func NewCryptoHandler(CryptoService *cryptoapplication.CryptoService) *CryptoHandler {
	return &CryptoHandler{
		CryptoService: CryptoService,
	}
}

// TransferRequest define la estructura de datos de la solicitud de transferencia
type TransferRequest struct {
	SourceAddress      string `json:"sourceAddress"`
	DestinationAddress string `json:"destinationAddress"`
	SignedTx           string `json:"signedTx"`
}

// TransferResponse define la estructura de datos de la respuesta de transferencia
type TransferResponse struct {
	TxHash string `json:"txHash"`
}

func (crypto *CryptoHandler) Subscription(c *fiber.Ctx) error {
	const quickNodeURL = "https://hardworking-small-breeze.bsc-testnet.discover.quiknode.pro/dcd23c83bc45d4cfe4e01bf88aeb0cc488e34746/"
	// Parsea los datos de la solicitud JSON
	var requestData TransferRequest
	if err := c.BodyParser(&requestData); err != nil {
		fmt.Println(requestData)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Error en los datos de la solicitud",
		})
	}

	// Crea una conexión RPC con QuickNode
	client, err := ethclient.Dial(quickNodeURL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Error al conectar con QuickNode: %v", err),
		})
	}

	txHash, err := crypto.CryptoService.TransferToken(client, requestData.SignedTx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Error al realizar la transferencia en BSC: %v", err),
		})
	}

	err = crypto.CryptoService.UpdataSubscriptionState(requestData.SourceAddress, requestData.DestinationAddress)
	if err != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "ok",
			"data":    err.Error(),
		})
	}

	response := TransferResponse{
		TxHash: txHash,
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"data":    response,
	})
}
