package cryptopinterface

import (
	cryptoapplication "PINKKER-BACKEND/internal/crypto/crypto-application"
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
	SourceAddress      string  `json:"sourceAddress"`
	DestinationAddress string  `json:"destinationAddress"`
	AmountUSD          float64 `json:"amountUSD"`
}

// TransferResponse define la estructura de datos de la respuesta de transferencia
type TransferResponse struct {
	TxHash string `json:"txHash"`
}

func (crypto *CryptoHandler) Prueba(c *fiber.Ctx) error {
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

	// Convierte el monto en USD a tokens BNB (ajusta esto según la tasa de cambio)
	amountTokens := USDToBNB(requestData.AmountUSD)

	// Realiza la transferencia de tokens
	txHash, err := TransferBNB(client, requestData.SourceAddress, requestData.DestinationAddress, amountTokens)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Error al realizar la transferencia: %v", err),
		})
	}

	// Prepara una respuesta JSON con el hash de la transacción
	response := TransferResponse{
		TxHash: txHash,
	}

	// Envía la respuesta JSON
	return c.JSON(response)
}

// USDToBNB realiza la conversión de USD a tokens BNB (ajusta esto según la tasa de cambio)
func USDToBNB(usdAmount float64) *big.Int {
	// Supongamos una tasa de cambio fija de 1 BNB = 500 USD (ajusta según tus necesidades)
	conversionRate := float64(500)
	amountBNB := usdAmount / conversionRate

	// Convierte el monto BNB a Wei (18 decimales)
	amountWei := new(big.Float).Mul(big.NewFloat(amountBNB), big.NewFloat(1e18))
	amount := new(big.Int)
	amountWei.Int(amount)

	return amount
}

// TransferBNB realiza la transferencia de tokens BNB
func TransferBNB(client *ethclient.Client, sourceAddress, destinationAddress string, amount *big.Int) (string, error) {
	// Convierte las direcciones en tipos comunes
	source := common.HexToAddress(sourceAddress)
	destination := common.HexToAddress(destinationAddress)

	// Obtiene el nonce del origen (número de transacción)
	nonce, err := client.PendingNonceAt(context.Background(), source)
	if err != nil {
		return "", err
	}
	// Establece la gas price y la cantidad máxima de gas
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	// Establece el límite de gas
	gasLimit := uint64(21000) // Gas limit para transferencias

	// Crea la transacción
	tx := types.NewTransaction(nonce, destination, amount, gasLimit, gasPrice, nil)

	// Firma la transacción
	chainID := big.NewInt(56) // Para Binance Smart Chain (BSC)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), nil)
	if err != nil {
		return "", err
	}

	// Envía la transacción
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	// Devuelve el hash de la transacción
	return signedTx.Hash().Hex(), nil
}
