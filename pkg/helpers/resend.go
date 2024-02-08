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
		Subject: "email confirmation code",
		Html:    "<p>codigo de confirmacion</p>" + code,
	}
	_, err := client.Emails.Send(params)
	fmt.Println(err)
	return err
}

func ResendRecoverPassword(code, To string) error {
	// http://vps-acad4de5.vps.ovh.ca/
	html := "<a href='http://localhost:3000/user/password-reset?reset_token=" + code + "'target='_blank'><button style='background-color:blue; color:white;'>restablecer contraseña</button></a>"
	apikey := config.ResendApi()
	RESENDDOMINIO := config.ResendDominio()
	client := resend.NewClient(apikey)
	params := &resend.SendEmailRequest{
		From:    RESENDDOMINIO,
		To:      []string{To},
		Subject: "email confirmation code",
		Html:    html,
	}
	_, err := client.Emails.Send(params)
	return err
}
