package helpers

import (
	"PINKKER-BACKEND/config"

	"github.com/resend/resend-go/v2"
)

func ResendConfirmMail(code, To string) error {
	apikey := config.ResendApi()
	client := resend.NewClient(apikey)
	params := &resend.SendEmailRequest{
		From:    "onboarding@resend.dev",
		To:      []string{"devspinnker@gmail.com"},
		Subject: "email confirmation code",
		Html:    "<p>codigo de confirmacion</p>" + code,
	}
	_, err := client.Emails.Send(params)
	return err
}

func ResendRecoverPassword(newPassword, To string) error {
	apikey := config.ResendApi()
	client := resend.NewClient(apikey)
	params := &resend.SendEmailRequest{
		From:    "onboarding@resend.dev",
		To:      []string{"devspinnker@gmail.com"},
		Subject: "email confirmation code",
		Html:    "<p>New Password</p>" + newPassword,
	}
	_, err := client.Emails.Send(params)
	return err
}
