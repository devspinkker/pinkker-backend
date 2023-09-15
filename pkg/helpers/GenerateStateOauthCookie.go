package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GenerateStateOauthCookie(c *fiber.Ctx) string {
	var expiration = time.Now().Add(2 * time.Minute)
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := fiber.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Expires:  expiration,
		HTTPOnly: true,
	}
	c.Cookie(&cookie)

	return state
}
