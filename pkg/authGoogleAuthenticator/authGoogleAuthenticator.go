package authGoogleAuthenticator

import (
	"github.com/pquerna/otp/totp"
)

// GenerateKey generates a new TOTP key
func GenerateKey(accountName, nameUser string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Pinkker.tv",
		AccountName: accountName,
	})
	if err != nil {
		return "", "", err
	}

	url := key.URL() + "&issuer=Pinkker.tv:" + nameUser
	return key.Secret(), url, nil
}

// ValidateCode validates the TOTP code
func ValidateCode(secret string, code string) bool {
	return totp.Validate(code, secret)
}
