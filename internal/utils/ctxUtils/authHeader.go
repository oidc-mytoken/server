package ctxUtils

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetAuthHeaderToken(ctx *fiber.Ctx) (token string) {
	authHeader := string(ctx.Request().Header.Peek("Authorization"))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		token = authHeader[7:]
	}
	return
}
