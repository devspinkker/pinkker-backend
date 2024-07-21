package middleware

import (
	"PINKKER-BACKEND/pkg/auth"
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/pquerna/otp/totp"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TOTPAuthMiddleware validates TOTP code using a TOTPRepository
func TOTPAuthMiddleware(repo auth.TOTPRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context (assumed to be set earlier in the request lifecycle)
		IdUserToken := c.Context().UserValue("_id").(string)

		userID, err := primitive.ObjectIDFromHex(IdUserToken)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID"})
		}

		// Get TOTP code from query parameters or body
		totpCode := c.Query("totp_code")
		if totpCode == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "TOTP code is required"})
		}

		// Retrieve the user's TOTP secret from the repository
		secret, err := repo.GetTOTPSecret(context.Background(), userID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Failed to retrieve TOTP secret"})
		}

		// Validate the TOTP code
		if !totp.Validate(totpCode, secret) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid TOTP code"})
		}

		// If everything is fine, continue with the next middleware/handler
		return c.Next()
	}
}
