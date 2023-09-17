package configoauth2

import (
	"PINKKER-BACKEND/config"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	GoogleLoginConfig oauth2.Config
}

var AppConfig Config

func LoadConfig() {
	GOOGLE_CLIENT_ID := config.GOOGLE_CLIENT_ID()
	GOOGLE_CLIENT_SECRET := config.GOOGLE_CLIENT_SECRET()
	AppConfig.GoogleLoginConfig = oauth2.Config{
		ClientID:     os.Getenv(GOOGLE_CLIENT_ID),
		ClientSecret: os.Getenv(GOOGLE_CLIENT_SECRET),
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:8080/google_callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.birthday",
			"https://www.googleapis.com/auth/userinfo.phone",
		},
	}
}
