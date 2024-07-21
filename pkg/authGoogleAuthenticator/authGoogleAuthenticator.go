package authGoogleAuthenticator

import (
	"fmt"
	"net/url"

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

	otpURL := url.URL{
		Scheme: "otpauth",
		Host:   "totp",
		Path:   fmt.Sprintf("%s:%s", url.QueryEscape("Pinkker.tv"), url.QueryEscape(nameUser)),
	}
	params := url.Values{}
	params.Add("secret", key.Secret())
	params.Add("issuer", "Pinkker.tv")
	params.Add("algorithm", "SHA1")
	params.Add("digits", "6")
	params.Add("period", "30")
	otpURL.RawQuery = params.Encode()

	return key.Secret(), otpURL.String(), nil
}

// ValidateCode validates the TOTP code
func ValidateCode(secret string, code string) bool {
	return totp.Validate(code, secret)
}
