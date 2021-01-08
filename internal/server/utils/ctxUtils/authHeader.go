package ctxUtils

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// GetAuthHeaderToken returns the Bearer token from the http authorization header
func GetAuthHeaderToken(ctx *fiber.Ctx) (token string) {
	authHeader := string(ctx.Request().Header.Peek("Authorization"))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		token = authHeader[7:]
	}
	return
}
