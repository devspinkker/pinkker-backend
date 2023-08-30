package mercadopago

import (
	"PINKKER-BACKEND/config"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func CreatePayment(User string, monto float64) (string, error) {
	// Datos del pago
	preference := map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"title":       "Pago único",
				"description": "Pago único por nuestro servicio",
				"quantity":    monto,
				"currency_id": "ARS",
				"unit_price":  monto,
			},
		},
		"payer": map[string]interface{}{
			"name":  User,
			"email": User,
			"_id":   User,
		},
		"back_urls": map[string]string{
			"success": "http://localhost:3000/SubscriptionStatus",
			"failure": "http://localhost:3000/SubscriptionStatus",
			"pending": "http://localhost:3000/SubscriptionStatus",
		},
		"auto_return": "approved",
		"payment_methods": map[string]interface{}{
			"excluded_payment_methods": []map[string]string{
				{"id": "amex"},
			},
			"excluded_payment_types": []map[string]string{
				{"id": "atm"},
			},
			"installments": 1,
		},
		"date_of_expiration": time.Now().Add(1 * time.Hour).Format("2006-01-02T15:04:05Z"),
	}

	// Codificar los datos de la preferencia en formato JSON
	payload, err := json.Marshal(preference)
	if err != nil {
		return "", err
	}

	// Enviar la URL de inicio del pago
	initPoint, err := getInitPoint(payload)
	if err != nil {
		return "", err
	}

	return initPoint, nil
}
func getInitPoint(payload []byte) (string, error) {
	// URL de la API de MercadoPago para crear una preferencia de pago
	apiURL := "https://api.mercadopago.com/checkout/preferences"

	ACCESS_TOKEN_PRUEBA_MERCADOPAGO := config.ACCESS_TOKEN_PRUEBA_MERCADOPAGO()

	// Crear una solicitud HTTP POST con el payload de la preferencia
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	// Agregar el encabezado de autorización con el token de acceso
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ACCESS_TOKEN_PRUEBA_MERCADOPAGO))
	req.Header.Set("Content-Type", "application/json")

	// Realizar la solicitud HTTP
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Decodificar la respuesta JSON
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	// Obtener el initPoint de la respuesta
	initPoint, ok := response["init_point"].(string)
	if !ok {
		return "", fmt.Errorf("no se pudo obtener el initPoint de la respuesta")
	}

	return initPoint, nil
}
