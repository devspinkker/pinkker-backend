package middleware

import (
	"PINKKER-BACKEND/pkg/jwt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func UseExtractor() fiber.Handler {

	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		token := strings.Replace(authHeader, "Bearer ", "", 1)

		nameUser, _id, verified, err := jwt.ExtractDataFromToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"Message": "Unauthorized",
			})
		}
		c.Context().SetUserValue("nameUser", nameUser)
		c.Context().SetUserValue("_id", _id)
		c.Context().SetUserValue("partner", verified)
		return c.Next()
	}

}
