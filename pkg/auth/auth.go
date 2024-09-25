package auth

import (
	"context"
	"errors"

	"github.com/pquerna/otp/totp"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TOTPRepository interface {
	GetTOTPSecret(ctx context.Context, userID primitive.ObjectID) (string, error)
}

func TOTPAutheLogin(TOTPCode string, secret string) (bool, error) {
	if !totp.Validate(TOTPCode, secret) {
		return false, errors.New("invalid TOTP code")
	}
	return true, nil

}
