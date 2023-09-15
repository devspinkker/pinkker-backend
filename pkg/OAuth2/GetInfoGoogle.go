package oauth2

import (
	userdomain "PINKKER-BACKEND/internal/user/user-domain"
	configoauth2 "PINKKER-BACKEND/pkg/OAuth2/configOAuth2"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

func GetUserInfoFromGoogle(token *oauth2.Token) (*userdomain.UserInfoOAuth2, error) {
	// Crea un cliente HTTP con el token de acceso.
	client := configoauth2.AppConfig.GoogleLoginConfig.Client(context.Background(), token)

	// Hacer una solicitud GET a la API de Google para obtener información del usuario.
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error al obtener información del usuario de Google. Código de estado: %d", resp.StatusCode)
	}

	var userInfo userdomain.UserInfoOAuth2
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
