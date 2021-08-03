package ctxUtils

import (
	"github.com/gofiber/fiber/v2"

	"github.com/oidc-mytoken/api/v0"
)

// ClientMetaData returns the model.ClientMetaData for a given fiber.Ctx
func ClientMetaData(ctx *fiber.Ctx) *api.ClientMetaData {
	return &api.ClientMetaData{
		IP:        ctx.IP(),
		UserAgent: string(ctx.Request().Header.UserAgent()),
	}
}
