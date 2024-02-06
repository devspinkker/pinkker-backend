package helpers

import (
	"PINKKER-BACKEND/config"
	"fmt"

	"github.com/resend/resend-go/v2"
)

func ResendConfirmMail(code, To string) error {
	fmt.Println(To)
	apikey := config.ResendApi()
	client := resend.NewClient(apikey)
	params := &resend.SendEmailRequest{
		From:    "onboarding@resend.dev",
		To:      []string{To},
		Subject: "email confirmation code",
		Html:    "<p>codigo de confirmacion</p>" + code,
	}
	_, err := client.Emails.Send(params)
	return err
}

func ResendRecoverPassword(code, To string) error {
	fmt.Println(To)
	html := "<a href='http://localhost:3000/user/password-reset?reset_token=" + code + "'target='_blank'><button style='background-color:blue; color:white;'>restablecer contraseña</button></a>"
	apikey := config.ResendApi()
	client := resend.NewClient(apikey)
	params := &resend.SendEmailRequest{
		From:    "onboarding@resend.dev",
		To:      []string{To},
		Subject: "email confirmation code",
		Html:    html,
	}
	_, err := client.Emails.Send(params)
	return err
}
