package authGoogleAuthenticator

import (
	"fmt"
	"net/url"
	"os"

	"github.com/mdp/qrterminal/v3"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

func GenerateKey(accountName, nameUser string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Pinkker",
		AccountName: nameUser,
	})
	if err != nil {
		return "", "", err
	}

	otpURL := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		url.QueryEscape("Pinkker"),
		url.QueryEscape(nameUser),
		url.QueryEscape(key.Secret()),
		url.QueryEscape("Pinkker"))

	return key.Secret(), otpURL, nil
}

func GenerateQRCode(otpURL string, filePath string) error {
	err := qrcode.WriteFile(otpURL, qrcode.Medium, 256, filePath)
	if err != nil {
		return err
	}

	qrterminal.GenerateWithConfig(otpURL, qrterminal.Config{
		Level:     qrterminal.L,
		Writer:    os.Stdout,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
	})

	fmt.Println("\nScan the QR code with your authentication app")
	return nil
}

func ValidateCode(secret string, code string) bool {
	return totp.Validate(code, secret)
}
