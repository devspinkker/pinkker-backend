package auth

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TOTPRepository interface {
	GetTOTPSecret(ctx context.Context, userID primitive.ObjectID) (string, error)
}
