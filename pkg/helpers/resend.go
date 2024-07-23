package helpers

import (
	"PINKKER-BACKEND/config"
	"fmt"

	"github.com/resend/resend-go/v2"
)

func ResendConfirmMail(code, To string) error {
	apikey := config.ResendApi()
	client := resend.NewClient(apikey)
	RESENDDOMINIO := config.ResendDominio()
	params := &resend.SendEmailRequest{
		From:    RESENDDOMINIO,
		To:      []string{To},
		Subject: "confirmacion de mail - pinkker",
		Html:    "<p>codigo de confirmacion</p>" + code,
	}
	_, err := client.Emails.Send(params)
	fmt.Println(err)
	return err
}

func ResendRecoverPassword(code, To string) error {
	html := "<a href='https://www.pinkker.tv/user/password-reset?reset_token=" + code + "'target='_blank'><button style='background-color:blue; color:white;'>restablecer contraseña</button></a>"
	apikey := config.ResendApi()
	RESENDDOMINIO := config.ResendDominio()
	client := resend.NewClient(apikey)
	params := &resend.SendEmailRequest{
		From:    RESENDDOMINIO,
		To:      []string{To},
		Subject: "recuperación de contraseña - pinkker",
		Html:    html,
	}
	_, err := client.Emails.Send(params)
	return err
}
func ChangeGoogleAuthenticator(code, To string) error {
	html := "<a href='https://www.pinkker.tv/user/password-reset?reset_token=" + code + "'target='_blank'><button style='background-color:blue; color:white;'>restablecer contraseña</button></a>"
	apikey := config.ResendApi()
	RESENDDOMINIO := config.ResendDominio()
	client := resend.NewClient(apikey)
	params := &resend.SendEmailRequest{
		From:    RESENDDOMINIO,
		To:      []string{To},
		Subject: "recuperación de contraseña - pinkker",
		Html:    html,
	}
	_, err := client.Emails.Send(params)
	return err
}
func ResendNotificationStreamerOnline(nameUser string, To []string) error {
	html := "<h1>" + nameUser + " Online<h1/>"
	apikey := config.ResendApi()
	RESENDDOMINIO := config.ResendDominio()
	client := resend.NewClient(apikey)
	params := &resend.SendEmailRequest{
		From:    RESENDDOMINIO,
		To:      To,
		Subject: nameUser + " acaba de prender en pinkker !!!",
		Html:    html,
	}
	_, err := client.Emails.Send(params)
	return err

}
